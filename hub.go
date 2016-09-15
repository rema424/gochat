// Hub transfers messages between clients using channels
// and stores list of all clients.

package main

import (
    "log"
    "encoding/json"
    "time"

    "github.com/gorilla/websocket"
)


type Hub struct {
    clients    map[*Client]bool
    broadcast  chan *Message
    info       chan *Message
    register   chan *Client
    unregister chan *Client
}


func makeHub() *Hub {
    h := &Hub{
        clients: make(map[*Client]bool),
        broadcast: make(chan *Message),
        register:  make(chan *Client),
        unregister: make(chan *Client),
    }

    return h
}


func (h *Hub) sendAll(msg *Message) {
    for client := range h.clients {
        if client.user != msg.Sender {
            msgJson, err := json.Marshal(msg)
            if err != nil {
                log.Println("JSON encoding error: ", err)
                continue
            }

            err = client.conn.WriteMessage(
                websocket.TextMessage,
                msgJson,
            )
            if err != nil {
                log.Println("Write error: ", err)
                client.conn.Close()
                delete(h.clients, client)
            }
        }
    }
}


func (h *Hub) run() {
    for {
        select {
        // Add client to chat
        case client := <-h.register:
            log.Println("Registered: "+client.user.Username)
            h.clients[client] = true

            // Tell everyone (except new user) about new user
            msg := &Message{
                Role: "new_user",
                Sender: client.user,
                Text: client.user.Username + " joined the room",
                SendDate: time.Now().Unix(),
            }
            h.sendAll(msg)

            // Send last 10 messages
            messages, err := getLastMessages(client.user, 10)
            if err != nil {
                log.Println(err)
                continue
            }

            for _, msg := range messages {
                msgJson, err := json.Marshal(msg)
                if err != nil {
                    log.Println("JSON encoding error: ", err)
                    continue
                }

                err = client.conn.WriteMessage(
                    websocket.TextMessage,
                    msgJson,
                )
                if err != nil {
                    log.Println("Write error: ", err)
                    client.conn.Close()
                    delete(h.clients, client)
                }
            }

        // Remove client from chat
        case client := <-h.unregister:
            log.Println("Unregistered: "+client.user.Username)
            client.conn.Close()
            delete(h.clients, client)

        // Send message to all and save it to DB as broadcast
        // message (no recipient and recieve date)
        case msg := <-h.broadcast:
            // Store only users' messages in DB
            if msg.Role == "message" {
                err := msg.save()
                if err != nil {
                    log.Println("Saving message error: ", err)
                    continue
                }
            }

            h.sendAll(msg)
        }
    }
}
