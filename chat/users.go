// Structure and methods for User objects.

package chat

import (
    "time"
    "errors"
    "strconv"
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


func (u *User) manage(h *Hub, act string) error {
    m := &Message{
        Action: act,
        SendDate: time.Now().UTC(),
    }

    switch act {
    case "kick":
        ctl := map[string]string{"user_id": strconv.Itoa(u.Id)}
        m.Control = &ctl

        for c := range h.clients {
            if c.user.Id == u.Id {
                un := &Unreg{
                    client: c,
                    msg: c.user.Username + " has been kicked",
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
        SELECT id, full_name, username, email, role
        FROM auth_user
        WHERE id = $1
    `)
    if err != nil {
        return nil, err
    }

    var user User
    err = stmt.QueryRow(id).Scan(
        &user.Id,
        &user.Fullname,
        &user.Username,
        &user.Email,
        &user.Role,
    )
    if err != nil {
        return nil, err
    }

    return &user, nil
}
