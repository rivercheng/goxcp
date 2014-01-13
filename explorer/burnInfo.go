package main
import (
    "net/http"
    "io/ioutil"
    "fmt"
    "encoding/json"
    "errors"
    "sort"
    "sync"
)

// ============== for Parsing =======================
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
    Txs           []Transaction `json:"txs"`
}


// ================ For stats ===============================
type Record struct {
    From     string
    BtcSpent int64
    XcpGet   float64
    BlockHeight int
    Hash     string
    Time     int
    Error    error
}

func (record *Record) String() string {
    if record.Error == nil {
        return fmt.Sprintf("sent BTC %.8f, get XCP %.8f, block height: %v, tx: %v", float64(record.BtcSpent)*1e-8, record.XcpGet, record.BlockHeight, record.Hash)
    } else {
        return fmt.Sprintf("sent BTC %.8f, block height: %v, tx: %v, invalid: %s", float64(record.BtcSpent)*1e-8, record.BlockHeight, record.Hash, record.Error.Error())
    }
}

type Status struct {
    BtcSpent int64
    XcpSum   float64
    Records  []*Record
    Overflowed bool
    Error    error
}

func (status *Status) String() string {
    if status.Error == nil {
        return fmt.Sprintf("BTC spent: %.8f\tXCP balance: %.8f", float64(status.BtcSpent) * 1e-8, status.XcpSum)
    }
    return fmt.Sprintf("BTC spent: %.8f\tXCP balance: %.8f\terror: %v", float64(status.BtcSpent) * 1e-8, status.XcpSum, status.Error)
}

var (
    globalStatusMap = make(map[string]*Status)
    globalTransactionCount int
    globalBlockHeight      string
    globalOutputText       string
    gMutex sync.Mutex
)

// =============== Config ===========================
const (
    MIN_BLOCK_HEIGHT = 278310
    MAX_BLOCK_HEIGHT = 283810
    BURN_ADDRESS = "1CounterpartyXXXXXXXXXXXXXXXUWLpVr"
    MAX_BURNT_BTC = 100000000
)

func validate(transaction *Transaction) (err error) {
    if transaction.BlockHeight == 0 {
        return errors.New("not confirmed yet")
    }
    /*
    // remove outputs with type != 0
    if len(transaction.Out) > 1 {
        outputs := make([]Output, 0, len(transaction.Out))
        for _, output := range transaction.Out {
            if output.Type == 0 {
                outputs = append(outputs, output)
            }
        }
        if len(outputs) < len(transaction.Out) {
            transaction.Out = outputs
        }
    }
    */
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
    }/* else if len(transaction.Out) > 2 {
        return errors.New("more than 2 outputs")
    }*/
    if len(transaction.Out) == 0 {
        return errors.New("no output")
    }
    /*
    if len(transaction.Out) = 2 {
        if transaction.Out[0].Addr != BURN_ADDRESS && transaction.Out[1].Addr != BURN_ADDRESS {
            return errors.New("no output address is the burn address")
        }
        if transaction.Out[1].Addr != transaction.Inputs[0].PrevOutput.Addr &&
           transaction.Out[0].Addr != transaction.Inputs[0].PrevOutput.Addr {
            return errors.New("the change address is not the sending address")
        }
    }
    */
    // Now output count is not checked. The only requirement is the first output should be the burn address
    if transaction.Out[0].Addr != BURN_ADDRESS {
        return errors.New("the first output address is not the burn address")
    }
    return nil
}

func CalculateXcp(btcSpent int64, blockHeight int) (xcpGet float64) {
    return float64(btcSpent) * 1e-8 * (1000 * (1 + 0.5 * (float64(MAX_BLOCK_HEIGHT - blockHeight)  / float64(MAX_BLOCK_HEIGHT - MIN_BLOCK_HEIGHT))))
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

// ===== help sort the map by xcp count =================
type StatusPair struct {
    Addr  string
    Value *Status
}

type StatusPairList []StatusPair
func (li StatusPairList) Swap(i, j int) {
    li[i], li[j] = li[j], li[i]
}
func (li StatusPairList) Len() int {
    return len(li)
}
func (li StatusPairList) Less(i, j int) bool {
    return li[i].Value.XcpSum < li[j].Value.XcpSum
}


func getResult() (output string) {
    currentHeightResp, err := http.Get("https://blockchain.info/q/getblockcount")
    if err != nil {
        return err.Error()
    }
    currentHeight, err := ioutil.ReadAll(currentHeightResp.Body)
    if err != nil {
        return err.Error()
    }
    currentHeightResp.Body.Close()
    gMutex.Lock()
    defer gMutex.Unlock()
    if string(currentHeight) == globalBlockHeight {
        return "current block height: " + string(currentHeight) + "\n" + globalOutputText
    }
    globalBlockHeight = string(currentHeight)
    statusMap := make(map[string]*Status)
    offset := 0
    transactionCount := 0
    for {
        requestStr := fmt.Sprintf("https://blockchain.info/address/%s?format=json&limit=50&offset=%d", BURN_ADDRESS, offset)
        resp, err := http.Get(requestStr)
        if err != nil {
            return err.Error()
        }
        res, err := ioutil.ReadAll(resp.Body)
        resp.Body.Close()
        var result Result
        err = json.Unmarshal(res, &result)
        if err != nil {
            return err.Error()
        }
        /*
        if result.NTx == globalTransactionCount {
            return fmt.Sprintf("current block height: %s\ncurrent transaction count: %d\n%d\n%s", string(currentHeight), result.NTx, globalTransactionCount, globalOutputText)
        }
        */
        globalTransactionCount = result.NTx
        if offset == 0 {
            output += fmt.Sprintf("BTC burnt:\t%.8f\n", float64(result.TotalReceived) * 1e-8)
        }
        for _, transaction := range result.Txs {
            record := parseTransaction(&transaction)
            status, present := statusMap[record.From]
            if !present {
                status = &Status{}
                statusMap[record.From] = status
            }
            status.BtcSpent += record.BtcSpent
            status.XcpSum += record.XcpGet
            status.Records = append(status.Records, record)
            if record.Error != nil {
                status.Error = record.Error
            }
            if status.BtcSpent > MAX_BURNT_BTC {
                status.Overflowed = true
                status.Error = errors.New("Exceeds 1 BTC limit")
            }
        }
        if len(result.Txs) < 50 {
            transactionCount = offset + len(result.Txs)
            break
        } else {
            offset += len(result.Txs)
        }
    }
    totalXcp := float64(0)
    statusList := make(StatusPairList, 0, len(statusMap))
    for addr, status := range statusMap {
        if status.Overflowed {
            fixOverflowedStatus(status)
        }
        totalXcp += status.XcpSum
        statusList = append(statusList, StatusPair{addr, status})
    }
    output += fmt.Sprintf("XCP created:\t%.8f\n", totalXcp)
    output += fmt.Sprintf("%d transactions\n", transactionCount)
    output += fmt.Sprintf("%d addresses\n", len(statusMap))
    sort.Sort(sort.Reverse(statusList))
    for i, pair := range statusList {
        output += fmt.Sprintf("%d\t%s\t%s\n", i+1, pair.Addr, pair.Value)
    }
    globalOutputText = output
    return "current block height: " + string(currentHeight) + "\n" + globalOutputText
}

func fixOverflowedStatus(status *Status) {
    records := make([]*Record, len(status.Records))
    for i := 0; i < len(status.Records); i++ {
        records[i] = status.Records[len(status.Records)-1-i]
    }
    status.XcpSum = 0
    status.BtcSpent = 0
    for _, record := range records {
        if status.BtcSpent + record.BtcSpent <= MAX_BURNT_BTC {
            status.BtcSpent += record.BtcSpent
            status.XcpSum += record.XcpGet
        } else if status.BtcSpent < MAX_BURNT_BTC {
            validSpent := MAX_BURNT_BTC - status.BtcSpent
            status.BtcSpent += record.BtcSpent
            status.XcpSum += CalculateXcp(validSpent, record.BlockHeight)
            record.Error = errors.New("partly dropped due to 1 BTC limit")
        } else { // ingore the BTC
            status.BtcSpent += record.BtcSpent
            record.Error = errors.New("completely dropped due to 1 BTC limit")
        }
    }
}
