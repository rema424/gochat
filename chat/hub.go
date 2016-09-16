// Hub transfers messages between clients using channels
// and stores list of all clients.

package chat

import (
    "log"
    "encoding/json"
    "time"

    "github.com/gorilla/websocket"
)


type Hub struct {
    clients    map[*Client]bool
    message    chan *Message
    info       chan *Message
    register   chan *Client
    unregister chan *Client
}


func makeHub() *Hub {
    h := &Hub{
        clients: make(map[*Client]bool),
        message: make(chan *Message),
        register:  make(chan *Client),
        unregister: make(chan *Client),
    }

    return h
}


func (h *Hub) send(msg *Message) {
    for client := range h.clients {
        // Don't send self messages
        toSelf := client.user == msg.Sender
        // Send only to recipient or if it is broadcast
        isBroadcast := msg.Recipient == nil
        isRecipient := !isBroadcast && (client.user.Id == msg.Recipient.Id)

        if !toSelf && (isBroadcast || isRecipient) {
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
                SendDate: time.Now(),
            }
            h.send(msg)

        // Remove client from chat
        case client := <-h.unregister:
            _, alive := h.clients[client]
            if alive {
                log.Println("Unregistered: "+client.user.Username)

                // Tell everyone about user has gone
                msg := &Message{
                    Role: "gone_user",
                    Sender: client.user,
                    Text: client.user.Username + " has gone",
                    SendDate: time.Now(),
                }
                h.send(msg)

                client.conn.Close()
                delete(h.clients, client)
            }

        // Send message to all and save it to DB as broadcast
        // message (no recipient and recieve date)
        case msg := <-h.message:
            // Store only users' messages in DB
            if msg.Role == "message" {
                err := msg.save()
                if err != nil {
                    log.Println("Saving message error: ", err)
                    continue
                }
            }

            if msg.Recipient != nil {
                log.Println(msg.Sender.Username+" TO "+msg.Recipient.Username+": "+msg.Text)
            } else {
                log.Println(msg.Sender.Username+": "+msg.Text)
            }

            h.send(msg)
        }
    }
}
