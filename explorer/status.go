package main
import (
    "fmt"
    "errors"
    "sort"
    "log"
)
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
//=======================================================

func CalculateXcp(btcSpent int64, blockHeight int) (xcpGet float64) {
    return float64(btcSpent) * 1e-8 * (1000 * (1 + 0.5 * (float64(MAX_BLOCK_HEIGHT - blockHeight)  / float64(MAX_BLOCK_HEIGHT - MIN_BLOCK_HEIGHT))))
}

func UpdateStatus(transactions []*Transaction, reorganized bool, statusMap map[string]*Status) {
    if reorganized {
        statusMap = make(map[string]*Status)
    }
    for i := 0; i < len(transactions); i++ {
        updateStatusFromTransaction(transactions[len(transactions)-1-i], statusMap)
    }
}

func updateStatusFromTransaction(transaction *Transaction, statusMap map[string]*Status) {
    record := parseTransaction(transaction)
    status, present := statusMap[record.From]
    if !present {
        status = &Status{}
        statusMap[record.From] = status
    }
    for i, r := range status.Records {
        if r.Hash == record.Hash {
            if r.BlockHeight != 0 {
                log.Print("ERROR: should not have duplicate records", r.Hash)
                return
            } else {
                status.Records[i] = record
                log.Print("INFO: replace unconfirmed record: ", record.Hash, " to block ", record.BlockHeight)
                updateStatusFromRecords(status)
                return
            }
        }
    }
    status.Records = append(status.Records, record)
    updateStatusFromRecords(status)
    return
}

func updateStatusFromRecords(status *Status) {
    status.BtcSpent = 0
    status.XcpSum = 0
    status.Error = nil
    status.Overflowed = false
    for _, record := range status.Records {
        if status.BtcSpent + record.BtcSpent <= MAX_BURNT_BTC {
            status.BtcSpent += record.BtcSpent
            status.XcpSum += record.XcpGet
        } else if status.BtcSpent < MAX_BURNT_BTC {
            validSpent := MAX_BURNT_BTC - status.BtcSpent
            status.BtcSpent += record.BtcSpent
            status.XcpSum += CalculateXcp(validSpent, record.BlockHeight)
            status.Error = errors.New("Exceeds 1 BTC limit")
            record.Error = errors.New("partly dropped due to 1 BTC limit")
        } else { // ingore the BTC
            status.BtcSpent += record.BtcSpent
            status.Error = errors.New("Exceeds 1 BTC limit")
            record.Error = errors.New("completely dropped due to 1 BTC limit")
        }
        if status.Error == nil {
            status.Error = record.Error
        }
        if status.BtcSpent > MAX_BURNT_BTC {
            status.Overflowed = true
        }
    }
}

func OrderByXcp(statusMap map[string]*Status) (totalXcp float64, statusList StatusPairList) {
    statusList = make(StatusPairList, 0, len(statusMap))
    for addr, status := range statusMap {
        totalXcp += status.XcpSum
        statusList = append(statusList, StatusPair{addr, status})
    }
    sort.Sort(sort.Reverse(statusList))
    return
}
