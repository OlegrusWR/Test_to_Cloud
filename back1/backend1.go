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
        fmt.Fprint(w, "Ответ от первого сервера")
    })


    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    fmt.Println("Backend 1 started on :8081")
    http.ListenAndServe(":8081", nil)
}