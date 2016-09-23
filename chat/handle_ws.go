// Serving client websocket connecions.

package chat

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "strings"
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
    hub     *Hub
    conn    *websocket.Conn
    user    *User
    message chan *Message
}


func (c *Client) readWS() {
    defer func() {
        c.conn.Close()
        un := &Unreg{
            client: c,
            msg: c.user.Username + " has gone",
        }
        c.hub.unregister <- un
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, data, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
                log.Println("Read error: ", err)
            }
            return
        }

        var msg Message
        err = json.Unmarshal(data, &msg)
        if err != nil {
            log.Println("JSON decode error: ", err)
            return
        }

        msg.Room = c.hub.room

        switch msg.Action {
        // Regular chat message
        case "message":
            msg.Sender = c.user
            c.hub.message <- &msg  // send to all
        // Control message from admin/moder
        case "mute", "kick", "ban":
            if c.user.checkPrivilege(msg.Action) {
                user, err := getUserById(msg.Recipient.Id)
                if err != nil {
                    log.Println("Unknown user: ", err)
                    return
                }
                room := c.hub.room
                err = room.manage(c.user, user, msg.Action)
                if err != nil {
                    log.Println("User manage error: ", err)
                    return
                }
            } else {
                log.Println("Not enough rights for action: "+msg.Action)
                return
            }
        }
    }
}


func (c *Client) writeWS() {
    ticker := time.NewTicker(pingPeriod)

    defer func() {
        ticker.Stop()
        c.conn.Close()
        un := &Unreg{
            client: c,
            msg: c.user.Username + " has gone",
        }
        c.hub.unregister <- un
    }()

    for {
        select {
        // Messages for this client
        case msg := <-c.message:
            msgJson, err := json.Marshal(msg)
            if err != nil {
                log.Println("JSON encoding error: ", err)
                continue
            }

            err = c.conn.WriteMessage(
                websocket.TextMessage,
                msgJson,
            )
            if err != nil {
                log.Println("Write error: ", err)
                c.conn.Close()
                delete(c.hub.clients, c)
            }

        // Heartbeat
        case <- ticker.C:
            err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
            if err != nil {
                return
            }
        }
    }
}


func handlerWS(w http.ResponseWriter, r *http.Request) {
    p := strings.Split(r.URL.Path, "/")
    roomId, err := strconv.Atoi(p[len(p)-1])
    if err != nil {
        log.Println("Invalid room ID: ", err)
        return
    }

    // Get hub from global hubs map
    hub, ok := hubs[roomId]
    if !ok {
        log.Println("No hub for this room: ", err)
        return
    }

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
        message: make(chan *Message),
    }
    client.hub.register <- &Reg{client: client}

    go client.writeWS()
    client.readWS()
}
