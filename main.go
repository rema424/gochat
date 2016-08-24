// Chat app entry point.

package main

import (
    "log"
    "net/http"
    "os"
)


func main() {
    var err error

    // Log file
    // f, err = os.OpenFile("log.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
    // if err != nil {
    //     log.Println("Cannot open log file for wrinting: ", err)
    // }
    // defer f.Close()
    // log.SetOutput(f)

    // Log to stdout
    log.SetOutput(os.Stdout)

    // Static files
    // fs := http.Dir("static")
    // fileHandler := http.FileServer(fs)
    // http.Handle("/static/", http.StripPrefix("/static/", fileHandler))

    // Messages exchanging
    hub := &Hub{
        clients: make(map[*Client]bool),
        broadcast: make(chan []byte),
        register:  make(chan *Client),
        unregister: make(chan *Client),
    }
    go hub.run()

    // Routing
    http.HandleFunc("/login", handlerLoginPage)
    http.HandleFunc("/logout", authMiddleware(handlerLogout))
    http.HandleFunc("/ws", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
        handlerWS(w, r, hub)
    }))
    http.HandleFunc("/", authMiddleware(handlerIndexPage))

    // Run server
    port := "8080"
    log.Printf("Server is running on %s port...\n", port)
    err = http.ListenAndServe(":"+port, nil)
    if err != nil {
        log.Println("ListenAndServe error: ", err)
    }
}
