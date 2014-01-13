package lib
import (
)

const (
    b26_digits = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
    BET_TYPE_NAME = map[int]string{0: "BullCFD", 1:"BearCFD", 2:"Equal", 3:"NotEqual"}
    BET_TYPE_ID   = map[string]int{"BullCFD": 0, "BearCFD":1, "Equal", 2, "NotEqual": 3}
)

func bitcoind_check() {
    block_count := bitcoin.rpc("getblockcount")
}

