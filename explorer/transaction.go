package main
import (
    "sync"
    "log"
)
var (
    gTransactions = make(map[string]*Transaction)
    gMutex        sync.Mutex
    gBTCBurnt     float64
)

func UpdateTransactions() (transactions []*Transaction, reorganized bool, err error) {
    transactions = make([]*Transaction, 0, 50)
    result, err := getTransactions(50, 0)
    if err != nil {
        return
    }
    gBTCBurnt = float64(result.TotalReceived) * 1e-8
    log.Print("BTC burnt: ", gBTCBurnt)
    gMutex.Lock()
    defer gMutex.Unlock()
    existed := false
    for _, tx := range result.Txs {
        t, present := gTransactions[tx.Hash]
        if !present {
            if existed {
                reorganized = true
            }
            transactions = append(transactions, tx)
        } else if t.BlockHeight == 0 && tx.BlockHeight != 0 {
            transactions = append(transactions, tx)
        } else {
            existed = true
        }
    }
    offset := 50
    for !existed || reorganized {
        result, err = getTransactions(50, offset)
        if err != nil {
            reorganized = false
            return
        }
        for _, tx := range result.Txs {
            _, present := gTransactions[tx.Hash]
            if !present {
                if existed {
                    reorganized = true
                }
                transactions = append(transactions, tx)
            } else {
                existed = true
            }
        }
        if len(result.Txs) < 50 {
            break
        }
        offset += 50
    }
    for _, trans := range transactions {
        gTransactions[trans.Hash] = trans
    }
    return
}
