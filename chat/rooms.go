// Structure and methods for Room objects.

package chat

import (
    "errors"
    "log"
    "time"
)

type Room struct {
    Id   int
    Name string
    hub  *Hub
}


// If user is muted in this room
func (r *Room) checkMute(userId int) (bool, error) {
    var isMuted bool
    err := stmtCheckMute.QueryRow(userId, r.Id).Scan(&isMuted)
    if err != nil {
        return false, err
    } else {
        return isMuted, nil
    }
}


// If user is banned in this room
func (r *Room) checkBan(userId int) (bool, error) {
    var isBanned bool
    err := stmtCheckBan.QueryRow(userId, r.Id).Scan(&isBanned)
    if err != nil {
        return false, err
    } else {
        return isBanned, nil
    }
}


// Mute, kick, ban
func (r *Room) manage(admin *User, user *User, act string) error {
    var t string
    var logMsg string

    switch act {
    case "mute":
        isMuted, err := r.checkMute(user.Id)
        if err != nil {
            return err
        }

        if isMuted {
            _, err = stmtUnmute.Exec(user.Id, r.Id)
            logMsg = "Unmuted: "
        } else {
            _, err = stmtMute.Exec(user.Id, r.Id)
            logMsg = "Muted: "
        }
        if err != nil {
            return err
        }

        msg := &Message{
            Action: "mute",
            Sender: admin,
            Recipient: user,
            Text: t,
            SendDate: time.Now().UTC(),
            Room: r,
        }
        r.hub.message <- msg
        log.Println(logMsg+user.Username)

        return nil

    case "kick":
        for c := range r.hub.clients {
            if c.user.Id == user.Id {
                un := &Unreg{
                    client: c,
                    msg: user.Username + " has been kicked by " + admin.Username,
                }
                r.hub.unregister <- un
            }
        }
        log.Println("Kicked: "+user.Username)
        return nil

    case "ban":
        for c := range r.hub.clients {
            if c.user.Id == user.Id {
                un := &Unreg{
                    client: c,
                    msg: user.Username + " has been banned",
                }
                r.hub.unregister <- un
            }
        }

        return nil
    }

    return errors.New("Wrong action: "+act)
}


func (r *Room) getUsers() []*User {
    users := []*User{}
    for c := range(r.hub.clients) {
        u := c.user
        u.addRoomInfo(r.Id)
        users = append(users, u)
    }

    return users
}


func (r *Room) getMessages(user *User, limit int) ([]*Message, error) {
    messages := []*Message{}

    rows, err := stmtGetMessages.Query(user.Id, r.Id, limit)
    if rows != nil {
        defer rows.Close()
    } else {
        return messages, nil
    }

    var msg *Message
    var sender *User
    var isBroadcast bool

    for rows.Next() {
        sender = &User{}
        msg = &Message{}
        err = rows.Scan(
            &sender.Id, &sender.Username, &sender.Fullname, &sender.Email, &sender.Role,
            &msg.Id, &msg.Action, &msg.Text, &msg.SendDate, &isBroadcast)
        if err != nil {
            return messages, err
        }
        msg.Sender = sender
        if !isBroadcast {
            msg.Recipient = user
            msg.Recipient.addRoomInfo(r.Id)
        }

        messages = append(messages, msg)
    }

    return messages, nil
}


func getAllRooms() ([]*Room, error) {
    var rooms []*Room

    rows, err := stmtGetAllRooms.Query()
    if rows != nil {
        defer rows.Close()
    } else {
        return rooms, nil
    }

    var room *Room
    for rows.Next() {
        room = &Room{}
        err = rows.Scan(&room.Id, &room.Name)
        if err != nil {
            return rooms, err
        }
        rooms = append(rooms, room)
    }

    return rooms, nil
}
