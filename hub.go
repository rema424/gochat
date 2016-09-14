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


func getLastMessages(user *User) ([]Message, error) {
    var messages []Message

    stmt, err := db.Prepare(`
        SELECT *
        FROM (
            SELECT u.id, u.username, m.id, 'message',
                m.text, m.send_date, m.recipient_id IS NULL
            FROM message AS m
            LEFT JOIN auth_user AS u ON u.id = m.sender_id
            WHERE m.recipient_id = $1 OR m.recipient_id IS NULL
            ORDER BY m.send_date DESC
            LIMIT 10
        ) AS tmp
        ORDER BY send_date ASC
    `)
    if err != nil {
        return []Message{}, err
    }

    rows, err := stmt.Query(user.Id)
    defer rows.Close()

    var msg Message
    var sender User
    var isBroadcast bool
    for rows.Next() {
        err = rows.Scan(
            &sender.Id, &sender.Username, &msg.Id, &msg.Role,
            &msg.Text, &msg.SendDate, &isBroadcast)
        if err != nil {
            return []Message{}, err
        }
        msg.Sender = &sender
        if !isBroadcast {
            msg.Recipient = user
        }

        messages = append(messages, msg)
    }

    return messages, nil
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
                SendDate: time.Now(),
            }
            h.sendAll(msg)

            // Send last 10 messages
            messages, err := getLastMessages(client.user)
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
                // Prohibit empty messages from users
                if msg.Text == "" {
                    continue
                }

                log.Println(msg.Sender.Username+": "+msg.Text)

                stmt, err := db.Prepare(`
                    INSERT INTO message
                    (sender_id, recipient_id, text, send_date)
                    VALUES
                    ($1, $2, $3, $4)
                `)
                if err != nil {
                    log.Println("Saving message error:", err)
                    continue
                }
                _, err = stmt.Exec(
                    &msg.Sender.Id, nil,
                    &msg.Text, &msg.SendDate)
                if err != nil {
                    log.Println("Saving message error:", err)
                    continue
                }
            }

            h.sendAll(msg)
        }
    }
}
