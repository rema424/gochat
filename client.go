package main

import (
    "fmt"
    "time"
    "net/http"

    "github.com/gorilla/websocket"
    "github.com/gorilla/context"
)


const (
    writeWait = 10 * time.Second
    pongWait = 3 * time.Second
    pingPeriod = 2 * time.Second
    maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type Client struct {
    hub *Hub
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
        fmt.Println(c.user.Username+": pong")
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, msg, err := c.conn.ReadMessage()
        if err != nil {
            fmt.Println("Read error: ", err)
            return
        }
        fmt.Println(c.user.Username+": "+string(msg))
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
            fmt.Println(c.user.Username+": ping")
            err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
            if err != nil {
                fmt.Println("Write error: ", err)
                return
            }
        }
    }
}


func handlerWS(w http.ResponseWriter, r *http.Request, hub *Hub) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("Open connection error: ", err)
        return
    }
    fmt.Println("Successfully connected")

    user := context.Get(r, "User").(*User)
    fmt.Println(user.FullName)

    client := &Client{
        hub: hub,
        conn: conn,
        user: user,
    }
    client.hub.register <- client

    go client.writeWS()
    client.readWS()
}
