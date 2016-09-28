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
    var text string
    var title string

    switch act {
    case "mute":
        isMuted, err := r.checkMute(user.Id)
        if err != nil {
            return err
        }

        if isMuted {
            _, err = stmtUnmute.Exec(user.Id, r.Id)
            title = "Unmuted: "
        } else {
            _, err = stmtMute.Exec(user.Id, r.Id)
            title = "Muted: "
        }
        if err != nil {
            return err
        }

        msg := &Message{
            Action: "mute",
            Sender: admin,
            Recipient: user,
            Text: text,
            SendDate: time.Now().UTC(),
            Room: r,
        }
        r.hub.ctl <- msg
        log.Println(title+user.Username)

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
        log.Println("Kicked:", user.Username)
        return nil

    case "ban":
        for c := range r.hub.clients {
            if c.user.Id == user.Id {
                _, err := stmtBan.Exec(user.Id, r.Id)
                if err != nil {
                    return err
                }

                un := &Unreg{
                    client: c,
                    msg: user.Username + " has been banned",
                }
                r.hub.unregister <- un
            }
        }
        log.Println("Banned:", user.Username)
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

    rows, err := stmtGetRoomMessagesByUser.Query(user.Id, r.Id, limit)
    if err == sql.ErrNoRows {
        return []*Message{}, nil
    } else if err != nil {
        return []*Message{}, err
    } else {
        defer rows.Close()
    }

    var msg *Message
    var sId, rId *int
    var sUsername, sFullname, sEmail, sRole *string
    var rUsername, rFullname, rEmail, rRole *string

    for rows.Next() {
        msg = &Message{}
        err = rows.Scan(
            // Message info
            &msg.Id, &msg.Action, &msg.Text, &msg.SendDate,
            // Sender
            &sId, &sUsername, &sFullname, &sEmail, &sRole,
            // Recipient
            &rId, &rUsername, &rFullname, &rEmail, &rRole,
        )
        if err != nil {
            return []*Message{}, err
        }
        if sId != nil {
            msg.Sender = &User{
                Id: *sId,
                Username: *sUsername,
                Fullname: *sFullname,
                Email: *sEmail,
                Role: *sRole,
            }
        }
        if rId != nil {
            msg.Recipient = &User{
                Id: *rId,
                Username: *rUsername,
                Fullname: *rFullname,
                Email: *rEmail,
                Role: *rRole,
            }
        }

        messages = append(messages, msg)
    }

    return messages, nil
}


func getUserRooms(userId int) (map[*Room]bool, error) {
    rooms := make(map[*Room]bool)

    rows, err := stmtGetUserRooms.Query(userId)
    if err == sql.ErrNoRows {
        return map[*Room]bool{}, nil
    } else if err != nil {
        return map[*Room]bool{}, err
    } else {
        defer rows.Close()
    }

    var room *Room
    var isBanned bool
    for rows.Next() {
        room = &Room{}
        err = rows.Scan(&room.Id, &room.Name, &isBanned)
        if err != nil {
            return map[*Room]bool{}, err
        }
        rooms[room] = isBanned
    }

    return rooms, nil
}


func getAllRooms() ([]*Room, error) {
    rooms := []*Room{}

    rows, err := stmtGetAllRooms.Query()
    if err == sql.ErrNoRows {
        return []*Room{}, nil
    } else if err != nil {
        return []*Room{}, err
    } else {
        defer rows.Close()
    }

    var room *Room
    for rows.Next() {
        room = &Room{}
        err = rows.Scan(&room.Id, &room.Name)
        if err != nil {
            return []*Room{}, err
        }
        rooms = append(rooms, room)
    }

    return rooms, nil
}
