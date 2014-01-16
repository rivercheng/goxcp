package main
import (
    "encoding/json"
    "errors"
)

type Input struct {
    PrevOutput Output `json:"prev_out"`
}

type Output struct {
    N     int `json:"n"`
    Value int64 `json:"value"`
    Addr  string `json:"addr"`
    TxIndex int  `json:"tx_index"`
    Type    int  `json:"type"`
}

type Transaction struct {
    BlockHeight int `json:"block_height"`
    Time        int `json:"time"`
    Inputs      []Input `json:"inputs"`
    Out         []Output `json:"out"`
    Hash        string `json:"hash"`
    TxIndex     int `json:"tx_index"`
    Ver         int `json:"ver"`
    Size        int `json:"size"`
}

type Result struct {
    TotalReceived int64 `json:"total_received"`
    FinalBalance  int64 `json:"final_balance"`
    NTx           int   `json:"n_tx"`
    Txs           []*Transaction `json:"txs"`
}

func parseResult(content []byte) (result *Result, err error) {
    result = &Result{}
    err = json.Unmarshal(content, result)
    return
}

func validate(transaction *Transaction) (err error) {
    if transaction.BlockHeight == 0 {
        return errors.New("not confirmed yet")
    }
    if transaction.BlockHeight < MIN_BLOCK_HEIGHT {
        return errors.New("earlier than start block")
    } else if transaction.BlockHeight > MAX_BLOCK_HEIGHT {
        return errors.New("later than last block")
    } else if len(transaction.Inputs) > 1 {
        // pass the one with all same input addresses
        var first_input Input
        for i, input := range transaction.Inputs {
            if i == 0 {
                first_input = input
            } else {
                if input.PrevOutput.Addr != first_input.PrevOutput.Addr {
                    return errors.New("multiple inputs")
                }
            }
        }
    }
    if len(transaction.Out) == 0 {
        return errors.New("no output")
    }
    // Now output count is not checked. The only requirement is the first output should be the burn address
    if transaction.Out[0].Addr != BURN_ADDRESS {
        return errors.New("the first output address is not the burn address")
    }
    return nil
}

func parseTransaction(transaction *Transaction) (record *Record) {
    record = &Record{}
    record.From = transaction.Inputs[0].PrevOutput.Addr
    record.BtcSpent = transaction.Out[0].Value
    record.BlockHeight = transaction.BlockHeight
    record.Hash = transaction.Hash
    record.Time = transaction.Time
    err := validate(transaction)
    if err != nil {
        if len(transaction.Out) > 1 {
            for _, out := range transaction.Out {
                if out.Addr == BURN_ADDRESS {
                    record.BtcSpent = out.Value
                    break
                }
            }
        }
        record.Error = err
        return
    }
    record.XcpGet = CalculateXcp(record.BtcSpent, record.BlockHeight)
    return
}
