package main

import (
	"fmt"
	"net/http"
    "log"
)

func main() {
	http.HandleFunc("/", hello)
	fmt.Println("listening...")
	err := http.ListenAndServe("0.0.0.0:80", nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
    log.Print("request from ", req.RemoteAddr)
	fmt.Fprintln(res, "Welcome to Counterpart Explorer (XCP)")
    fmt.Fprintln(res, "Currently, only XCP obtained by burning are listed.")
    fmt.Fprintln(res, "Please use it only to confirm your burning")
    fmt.Fprintln(res, "Burn starts at block: ", MIN_BLOCK_HEIGHT)
    fmt.Fprintln(res, "Burn stops  at block: ", MAX_BLOCK_HEIGHT)
    fmt.Fprintln(res, getResult())
}
