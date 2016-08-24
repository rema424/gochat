// Hub transfers messages between client using channels
// and stores list of all clients.

package main

import (
    "log"
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
            log.Println("Registered: "+client.user.Username)
            h.clients[client] = true
        case client := <-h.unregister:
            log.Println("Unregistered: "+client.user.Username)
            client.conn.Close()
            delete(h.clients, client)
        case msg := <-h.broadcast:
            for client := range h.clients {
                err := client.conn.WriteMessage(websocket.TextMessage, msg)
                if err != nil {
                    log.Println("Write error: ", err)
                    client.conn.Close()
                    delete(h.clients, client)
                }
            }
        }
    }
}
