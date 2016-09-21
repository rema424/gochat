// Structure and methods for User objects.

package chat

import (
    "database/sql"
    "time"
    // "log"
)


type User struct {
    Id       int        `json:"id"`
    Fullname string     `json:"fullname"`
    Username string     `json:"username"`
    Email    string     `json:"email"`
    Password string     `json:"-"`
    Role     string     `json:"role"`
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
            email = $4
        WHERE id = $1
    `)
    if err != nil {
        return err
    }

    _, err = stmt.Exec(
        &u.Id, &u.Fullname, &u.Username, &u.Email,
    )
    if err != nil {
        return err
    }

    return nil
}


// Check if user can do action
func (u *User) checkPrivilege(act string) bool {
    return contains(permissions[u.Role], act)
}


// Add role and mute data in the room for the user
func (u *User) addRoomInfo(roomId int) error {
    stmt, err := db.Prepare(`
        SELECT
            CASE
                WHEN rn.name IS NULL THEN 'user'
                ELSE rn.name
            END AS role,
            m.date
        FROM room AS r
        LEFT JOIN room_role AS rr ON rr.room_id = r.id AND rr.user_id = $1
        LEFT JOIN role_name AS rn ON rr.role_id = rn.id
        LEFT JOIN mute AS m ON m.room_id = r.id AND rr.user_id = $1
        WHERE r.id = $2
    `)
    if err != nil {
        return err
    }

    err = stmt.QueryRow(u.Id, roomId).Scan(&u.Role, &u.MuteDate)
    if err != nil && err != sql.ErrNoRows {
        return err
    } else {
        return nil
    }
}


func getUserById(id int) (*User, error) {
    stmt, err := db.Prepare(`
        SELECT id, full_name, username, email
        FROM auth_user
        WHERE id = $1
    `)
    if err != nil {
        return nil, err
    }

    var user User
    err = stmt.QueryRow(id).Scan(
        &user.Id, &user.Fullname, &user.Username, &user.Email,
    )
    if err != nil {
        return nil, err
    }

    return &user, nil
}
