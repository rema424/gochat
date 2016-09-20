// Structure and methods for User objects.

package chat

import (
    "time"
    "errors"
    "log"
)


type User struct {
    Id       int        `json:"id"`
    Fullname string     `json:"fullname"`
    Username string     `json:"username"`
    Email    string     `json:"email"`
    Password string     `json:"-"`
    Role     string     `json:"role"`
    // Penalties
    Ban      bool       `json:"ban"`
    BanDate  *time.Time `json:"ban_date"`
    Mute     bool       `json:"mute"`
    MuteDate *time.Time `json:"mute_date"`
}

// Permissions for managing users
var permissions = map[string][]string{
    "admin": []string{"mute", "kick", "ban"},
    "moder": []string{"mute", "kick"},
    "user":  []string{},
}


// TODO: Add insert/update argument
func (u *User) save() error {
    stmt, err := db.Prepare(`
        UPDATE auth_user
        SET
            full_name = $2,
            username = $3,
            email = $4,
            password = $5,
            role = $6,
            is_muted = $7,
            mute_date = $8,
            is_banned = $9,
            ban_date = $10
        WHERE id = $1
    `)
    if err != nil {
        return err
    }

    _, err = stmt.Exec(
        &u.Id, &u.Fullname, &u.Username,
        &u.Email, &u.Password, &u.Role,
        &u.Mute, &u.MuteDate,
        &u.Ban, &u.BanDate)
    if err != nil {
        return err
    }

    return nil
}


// Mute, kick, ban
func (u *User) manage(h *Hub, admin *User, act string) error {
    var t string
    now := time.Now().UTC()

    switch act {
    case "mute":
        u.Mute = !u.Mute
        if u.Mute {
            u.MuteDate = &now
            t = u.Username + " has been muted"
        } else {
            u.MuteDate = nil
            t = u.Username + " has been unmuted"
        }

        err := u.save()
        if err != nil {
            return err
        }

        msg := &Message{
            Action: "mute",
            Sender: admin,
            Recipient: u,
            Text: t,
            SendDate: time.Now().UTC(),
        }
        h.message <- msg
        log.Println("Muted: "+u.Username)

        return nil

    case "kick":
        for c := range h.clients {
            if c.user.Id == u.Id {
                un := &Unreg{
                    client: c,
                    msg: u.Username + " has been kicked",
                }
                h.unregister <- un
            }
        }
        log.Println("Kicked: "+u.Username)
        return nil

    case "ban":
        u.Ban = true
        u.BanDate = &now

        err := u.save()
        if err != nil {
            return err
        }

        for c := range h.clients {
            if c.user.Id == u.Id {
                un := &Unreg{
                    client: c,
                    msg: u.Username + " has been banned",
                }
                h.unregister <- un
            }
        }

        return nil
    }


    return errors.New("Wrong action: "+act)
}


// Check if user can do action
func (u *User) checkPrivilege(act string) bool {
    return contains(permissions[u.Role], act)
}


func getUserById(id int) (*User, error) {
    stmt, err := db.Prepare(`
        SELECT id, full_name, username, email, role,
            is_muted, mute_date, is_banned, ban_date
        FROM auth_user
        WHERE id = $1
    `)
    if err != nil {
        return nil, err
    }

    var user User
    err = stmt.QueryRow(id).Scan(
        &user.Id, &user.Fullname, &user.Username,
        &user.Email, &user.Role,
        &user.Mute, &user.MuteDate,
        &user.Ban, &user.BanDate,
    )
    if err != nil {
        return nil, err
    }

    return &user, nil
}
