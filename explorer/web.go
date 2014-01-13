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
    fmt.Fprintln(res, getResult())
}
