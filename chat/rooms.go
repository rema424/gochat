// Structure and methods for Room objects.

package chat

import (
    "database/sql"
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
    stmt, err := db.Prepare(`
        SELECT 1
        FROM mute
        WHERE user_id = $1 AND room_id = $2
    `)
    if err != nil {
        return false, err
    }

    err = stmt.QueryRow(userId, r.Id).Scan()
    if err == sql.ErrNoRows {
        return false, nil
    } else if err != nil {
        return false, err
    } else {
        return true, nil
    }
}


// If user is banned in this room
func (r *Room) checkBan(userId int) (bool, error) {
    stmt, err := db.Prepare(`
        SELECT 1
        FROM ban
        WHERE user_id = $1 AND room_id = $2
    `)
    if err != nil {
        return false, err
    }

    err = stmt.QueryRow(userId, r.Id).Scan()
    if err == sql.ErrNoRows {
        return false, nil
    } else if err != nil {
        return false, err
    } else {
        return true, nil
    }
}


// Mute, kick, ban
func (r *Room) manage(admin *User, user *User, act string) error {
    var t string

    switch act {
    case "mute":
        isMuted, err := r.checkMute(user.Id)
        if err != nil {
            return err
        }
        var stmt *sql.Stmt
        if isMuted {
            stmt, err = db.Prepare(`
                DELETE FROM mute
                WHERE user_id = $1 AND room_id = $2
            `)
        } else {
            stmt, err = db.Prepare(`
                INSERT INTO mute (user_id, room_id)
                VALUES ($1, $2)
            `)
        }
        if err != nil {
            return err
        }

        _, err = stmt.Exec(user.Id, r.Id)
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
        log.Println("Muted: "+user.Username)

        return nil

    case "kick":
        for c := range r.hub.clients {
            if c.user.Id == user.Id {
                un := &Unreg{
                    client: c,
                    msg: user.Username + " has been kicked",
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
    var users []*User
    for c := range(r.hub.clients) {
        u := c.user
        u.addRoomInfo(r.Id)
        users = append(users, u)
    }

    return users
}


func (r *Room) getMessages(user *User, limit int) ([]*Message, error) {
    var messages []*Message

    stmt, err := db.Prepare(`
        SELECT *
        FROM (
            SELECT u.id, u.username, u.full_name, u.email, rn.name AS role,
                m.id, 'message', m.text, m.send_date,
                m.recipient_id IS NULL
            FROM message AS m
            LEFT JOIN auth_user AS u ON u.id = m.sender_id
            LEFT JOIN room_role AS rr ON rr.room_id = m.room_id AND rr.user_id = u.id
            LEFT JOIN role_name AS rn ON rr.role_id = rn.id
            WHERE (m.recipient_id = $1 OR m.recipient_id IS NULL)
                AND m.room_id = $2
            ORDER BY m.send_date DESC
            LIMIT $3
        ) AS tmp
        ORDER BY send_date ASC
    `)
    if err != nil {
        return messages, err
    }

    rows, err := stmt.Query(user.Id, r.Id, limit)
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

    stmt, err := db.Prepare(`
        SELECT id, name
        FROM room
    `)
    if err != nil {
        return rooms, err
    }

    rows, err := stmt.Query()
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
