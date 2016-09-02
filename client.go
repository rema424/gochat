// Serving client connecions.

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/context"
    "github.com/gorilla/websocket"
)


const (
    pongWait = 3 * time.Second
    pingPeriod = 2 * time.Second
    maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type Client struct {
    hub  *Hub
    conn *websocket.Conn
    user *User
}

type Message struct {
    id        int
    role      string
    sender    *User
    recipient *User
    text      string
    send_date time.Time
}


// Make map for JSON encoding
func (m *Message) toMap() map[string]string {
    var r string
    if m.recipient != nil {
        r = m.recipient.username
    } else {
        r = ""
    }

    res := map[string]string{
        "role": m.role,
        "sender": m.sender.username,
        "recipient": r,
        "date": fmt.Sprintf("%02d:%02d", m.send_date.Hour(), m.send_date.Minute()),
        "text": m.text,
    }

    return res
}


// Make JSON for sending to websocket
func (m *Message) toJson() ([]byte, error) {
    res, err := json.Marshal(m.toMap())
    return res, err
}


func (c *Client) readWS() {
    defer func() {
        c.conn.Close()
        c.hub.unregister <- c
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        log.Println(c.user.username+": pong")
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, data, err := c.conn.ReadMessage()
        if err != nil {
            log.Println("Read error: ", err)
            return
        }

        var msgJson map[string]string
        err = json.Unmarshal(data, &msgJson)
        if err != nil {
            log.Println("JSON decode error: ", err)
            return
        }

        msg := &Message{
            role: "message",
            sender: c.user,
            recipient: nil,
            text: msgJson["text"],
            send_date: time.Now(),
        }
        log.Println(c.user.username+": "+msg.text)
        c.hub.broadcast <- msg  // send to all
    }
}


func (c *Client) writeWS() {
    ticker := time.NewTicker(pingPeriod)

    defer func() {
        ticker.Stop()
        c.conn.Close()
        c.hub.unregister <- c
    }()

    for {
        select {
        // Heartbeat
        case <- ticker.C:
            log.Println(c.user.username+": ping")

            err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
            if err != nil {
                log.Println("Ping error: ", err)
                return
            }
        }
    }
}


func handlerWS(w http.ResponseWriter, r *http.Request, hub *Hub) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Open connection error: ", err)
        return
    }
    log.Println("Successfully connected")

    user := context.Get(r, "User").(*User)

    client := &Client{
        hub: hub,
        conn: conn,
        user: user,
    }
    client.hub.register <- client

    go client.writeWS()
    client.readWS()
}
