// Hub transfers messages between client using channels
// and stores list of all clients.

package main

import (
    "log"
    "github.com/gorilla/websocket"
)


type Hub struct {
    clients    map[*Client]bool
    broadcast  chan *Message
    register   chan *Client
    unregister chan *Client
}


func getLastMessages(user *User) ([]string, error) {
    var messages []string

    stmt, err := db.Prepare(`
        SELECT msg
        FROM (
            SELECT u.username || ': ' || m.text AS msg, send_date
            FROM message AS m
            LEFT JOIN auth_user AS u ON u.id = m.sender_id
            WHERE m.recipient_id = $1 OR m.recipient_id IS NULL
            ORDER BY send_date DESC
            LIMIT 10
        ) AS tmp
        ORDER BY send_date ASC
    `)
    if err != nil {
        return []string{}, err
    }

    rows, err := stmt.Query(user.id)
    defer rows.Close()

    var msg string
    for rows.Next() {
        err = rows.Scan(&msg)
        if err != nil {
            return []string{}, err
        }
        messages = append(messages, msg)
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
                err := client.conn.WriteMessage(
                    websocket.TextMessage,
                    []byte(msg),
                )
                if err != nil {
                    log.Println("Write error: ", err)
                    client.conn.Close()
                    delete(h.clients, client)
                }
            }

        // Remove client from chat
        case client := <-h.unregister:
            log.Println("Unregistered: "+client.user.username)
            client.conn.Close()
            delete(h.clients, client)

        // Send message to all and save it to DB as broadcast
        // message (no recepient and recieve date)
        case msg := <-h.broadcast:
            stmt, err := db.Prepare(`
                INSERT INTO message
                (sender_id, text, send_date)
                VALUES
                ($1, $2, CURRENT_TIMESTAMP)
            `)
            if err != nil {
                log.Println("Saving message error:", err)
            }
            _, err = stmt.Exec(&msg.sender.id, &msg.text)
            if err != nil {
                log.Println("Saving message error:", err)
            }

            for client := range h.clients {
                err := client.conn.WriteMessage(
                    websocket.TextMessage,
                    []byte(client.user.username+": "+string(msg.text)),
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
