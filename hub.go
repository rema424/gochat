// Hub transfers messages between clients using channels
// and stores list of all clients.

package main

import (
    "log"
    "encoding/json"

    "github.com/gorilla/websocket"
)


type Hub struct {
    clients    map[*Client]bool
    broadcast  chan *Message
    info       chan *Message
    register   chan *Client
    unregister chan *Client
}

func getLastMessages(user *User) ([]map[string]string, error) {
    var messages []map[string]string

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
        return []map[string]string{}, err
    }

    rows, err := stmt.Query(user.id)
    defer rows.Close()

    var msg Message
    var sender User
    var isBroadcast bool
    for rows.Next() {
        err = rows.Scan(
            &sender.id, &sender.username, &msg.id, &msg.role,
            &msg.text, &msg.send_date, &isBroadcast)
        if err != nil {
            return []map[string]string{}, err
        }
        msg.sender = &sender
        if !isBroadcast {
            msg.recipient = user
        }

        messages = append(messages, msg.toMap())
    }

    return messages, nil
}


func (h *Hub) run() {
    for {
        select {
        // Add client to chat
        case client := <-h.register:
            log.Println("Registered: "+client.user.username)
            h.clients[client] = true

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

            // Tell everybody about new user
            msg := &Message{
                role: "new_user",
                text: client.user.username + " joined the room",
            }
            h.broadcast <- msg

        // Remove client from chat
        case client := <-h.unregister:
            log.Println("Unregistered: "+client.user.username)
            client.conn.Close()
            delete(h.clients, client)

        // Send message to all and save it to DB as broadcast
        // message (no recipient and recieve date)
        case msg := <-h.broadcast:
            // Store only users' messages in DB
            if msg.role == "message" {
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
                    &msg.sender.id, nil,
                    &msg.text, &msg.send_date)
                if err != nil {
                    log.Println("Saving message error:", err)
                    continue
                }
            }

            for client := range h.clients {
                msgJson, err := json.Marshal(msg.toMap())
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
}
