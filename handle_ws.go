// Serving client websocket connecions.

package main

import (
    "encoding/json"
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


func (c *Client) readWS() {
    defer func() {
        c.conn.Close()
        c.hub.unregister <- c
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        log.Println(c.user.Username+": pong")
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, data, err := c.conn.ReadMessage()
        if err != nil {
            log.Println("Read error: ", err)
            return
        }

        // var msgJson map[string]string
        var msg Message
        err = json.Unmarshal(data, &msg)
        if err != nil {
            log.Println("JSON decode error: ", err)
            return
        }

        msg.Sender = c.user
        c.hub.message <- &msg  // send to all
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
            log.Println(c.user.Username+": ping")

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
