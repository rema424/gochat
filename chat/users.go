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
    _, err := stmtUpdateUser.Exec(
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
    err := stmtGetUserRoomInfo.QueryRow(u.Id, roomId).Scan(&u.Role, &u.MuteDate)
    if err != nil && err != sql.ErrNoRows {
        return err
    } else {
        return nil
    }
}


func getUserById(id int) (*User, error) {
    var user User
    err := stmtGetUserById.QueryRow(id).Scan(
        &user.Id, &user.Fullname, &user.Username, &user.Email,
    )
    if err != nil {
        return nil, err
    }

    return &user, nil
}
