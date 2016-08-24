package main

import (
    "fmt"
    "net/http"
)


func main() {
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
    fmt.Printf("Server is running on %s port...\n", port)
    http.ListenAndServe(":"+port, nil)
}
