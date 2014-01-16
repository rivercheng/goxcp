package main
import (
    "fmt"
    "sync"
    "log"
)

// =============== Config ===========================
const (
    MIN_BLOCK_HEIGHT = 278310
    MAX_BLOCK_HEIGHT = 283810
    BURN_ADDRESS = "1CounterpartyXXXXXXXXXXXXXXXUWLpVr"
    MAX_BURNT_BTC = 100000000
)

var (
    globalStatusMap = make(map[string]*Status)
    gStatusMapMutex sync.Mutex
    lastOutput string
    gTotalXcp  float64
    gStatusList StatusPairList
)

func getResult() (output string) {
    transactions, reorganized, err := UpdateTransactions()
    log.Print("new transactions: ", len(transactions), " reorganized: ", reorganized)
    if err != nil {
        log.Print(err)
        return lastOutput
    }
    currentHeight, err := getBlockHeight()
    if err != nil {
        log.Print(err)
        return lastOutput
    }
    if len(transactions) > 0 {
        gStatusMapMutex.Lock()
        defer gStatusMapMutex.Unlock()
        UpdateStatus(transactions, reorganized, globalStatusMap)
        gTotalXcp, gStatusList = OrderByXcp(globalStatusMap)
    }

    lastOutput = "current block height:\t" + currentHeight + "\n"
    lastOutput += fmt.Sprintf("BTC burnt:\t%.8f\n", gBTCBurnt)
    lastOutput += fmt.Sprintf("XCP created:\t%.8f\n", gTotalXcp)
    lastOutput += fmt.Sprintf("%d transactions\n", len(gTransactions))
    lastOutput += fmt.Sprintf("%d addresses\n", len(globalStatusMap))
    for i, pair := range gStatusList {
        lastOutput += fmt.Sprintf("%d\t%s\t%s\n", i+1, pair.Addr, pair.Value)
    }
    return lastOutput
}
