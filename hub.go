package main

import (
    "fmt"

    "github.com/gorilla/websocket"
)


type Hub struct {
    clients map[*Client]bool
    broadcast chan []byte
    register chan *Client
    unregister chan *Client
}


func (h *Hub) run() {
    for {
        select {
        case client := <-h.register:
            fmt.Println("Registered: ", client)
            h.clients[client] = true
        case client := <-h.unregister:
            fmt.Println("Unregistered: ", client)
            client.conn.Close()
            delete(h.clients, client)
        case msg := <-h.broadcast:
            for client := range h.clients {
                err := client.conn.WriteMessage(websocket.TextMessage, msg)
                if err != nil {
                    fmt.Println("Error while sending: ", err)
                    client.conn.Close()
                    delete(h.clients, client)
                }
            }
        }
    }
}
