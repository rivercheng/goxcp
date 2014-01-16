package main
import (
    "net/http"
    "io/ioutil"
    "fmt"
    "log"
)
func getBlockHeight() (currentHeight string, err error) {
    log.Print("get block height")
    currentHeightResp, err := http.Get("https://blockchain.info/q/getblockcount")
    if err != nil {
        return
    }
    currentHeightBytes, err := ioutil.ReadAll(currentHeightResp.Body)
    defer currentHeightResp.Body.Close()
    if err != nil {
        return
    }
    currentHeight = string(currentHeightBytes)
    return
}
func getTransactions(limit, offset int) (result *Result, err error) {
    log.Print("get transactions: ", limit, " ", offset)
    requestStr := fmt.Sprintf("https://blockchain.info/address/%s?format=json&limit=%d&offset=%d", BURN_ADDRESS, limit, offset)
    resp, err := http.Get(requestStr)
    if err != nil {
        return
    }
    res, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    result, err = parseResult(res)
    return
}
