package main

import (
    "fmt"
    "net/http"
    "time"
    "os"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(10 * time.Millisecond)
        fmt.Fprintf(os.Stderr, "Запрос от %s\n", r.RemoteAddr) 
        fmt.Fprint(w, "Ответ от второго сервера")
    })


    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    fmt.Println("Backend 2 started on :8082")
    http.ListenAndServe(":8082", nil)
}