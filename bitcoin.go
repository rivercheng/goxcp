package lib
import (
    "net/http"
    "encoding/json"
    "github.com/conformal/btcjson"
)
const (
    OP_RETURN    = 0x6a
    OP_PUSHDATA1 = 0x4c
    OP_DUP       = 0x76
    OP_HASH160   = 0xa9
    OP_EQUALVERIFY = 0x88
    OP_CHECKSIG    = 0xac
    OP_1           = 0x51
    OP_2           = 0x52
    OP_CHECKMULTISIG = 0xae
    b58_digits       = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

func connect(host string, payload interface{}, headers http.Header) (resp *http.Response, err error) {
    TRIES := 12
    request, err := http.NewRequest("POST", host, body)
    if err != nil {
        return
    }
    for k, v := range headers {
        http.Header.Add(k, v)
    }
    for i := 0; i < TRIES; i++ {

    }
}
