// Hub transfers messages between clients using channels
// and stores list of all clients.

package chat

import (
    "log"
    "time"
)


type Hub struct {
    clients    map[*Client]bool
    room       *Room
    message    chan *Message
    info       chan *Message
    register   chan *Reg
    unregister chan *Unreg
}

type Reg struct {
    client *Client
}

type Unreg struct {
    client *Client
    msg    string
}

func makeHub(room *Room) *Hub {
    h := &Hub{
        clients: make(map[*Client]bool),
        room: room,
        message: make(chan *Message),
        register:  make(chan *Reg),
        unregister: make(chan *Unreg),
    }

    // Make reverse relation
    h.room.hub = h

    return h
}


func (h *Hub) send(msg *Message) {
    for client := range h.clients {
        // Gone messages are sent to everyone (including sender)
        toAll := contains([]string{"gone_user", "mute", "ban"}, msg.Action)
        // Don't send self messages
        toSelf := msg.Sender != nil && client.user.Id == msg.Sender.Id
        // Send only to recipient or if it is broadcast
        isBroadcast := msg.Recipient == nil
        isRecipient := !isBroadcast && (client.user.Id == msg.Recipient.Id)

        doSend := toAll || (!toSelf && (isBroadcast || isRecipient))

        if doSend {
            client.message <- msg
        }
    }
}


func (h *Hub) run() {
    for {
        select {
        // Add client to chat
        case reg := <-h.register:
            client := reg.client
            log.Println("Registered: "+client.user.Username)
            h.clients[client] = true

            // Tell everyone (except new user) about new user
            msg := &Message{
                Action: "new_user",
                Sender: client.user,
                Text: client.user.Username + " joined the room",
                SendDate: time.Now().UTC(),
            }
            h.send(msg)

        // Remove client from chat
        case unreg := <-h.unregister:
            client := unreg.client
            msg := unreg.msg
            _, alive := h.clients[client]
            if alive {
                log.Println("Unregistered: "+client.user.Username)

                // Tell everyone about user has gone
                msg := &Message{
                    Action: "gone_user",
                    Sender: client.user,
                    Text: msg,
                    SendDate: time.Now().UTC(),
                }
                h.send(msg)

                client.conn.Close()
                delete(h.clients, client)
            }

        // Send message to all and save it to DB as broadcast
        // message (no recipient and recieve date)
        case msg := <-h.message:
            // If user is muted - do nothing
            if msg.Sender.MuteDate != nil {
                continue
            }

            // Store only users' messages in DB
            if msg.Action == "message" {
                err := msg.save()
                if err != nil {
                    log.Println("Saving message error: ", err)
                    continue
                }

                if msg.Recipient != nil {
                    log.Println(msg.Sender.Username+" TO "+msg.Recipient.Username+": "+msg.Text)
                } else {
                    log.Println(msg.Sender.Username+": "+msg.Text)
                }
            }

            h.send(msg)
        }
    }
}
