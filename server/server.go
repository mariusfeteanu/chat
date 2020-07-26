package main

import (
    "net/http"
    //"html"
    "fmt"
    "log"
    "io/ioutil"
)

func main() {
    log.Println("Starting server")

    // http.Handle("/foo", fooHandler)

    http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
        message_bytes, _ := ioutil.ReadAll(r.Body)
        log.Println(fmt.Sprintf("%v -> %v: %v",
            r.Header["Chat-From"],
            r.Header["Chat-To"],
            string(message_bytes)))
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
