// Structure and methods for User objects.

package chat


type User struct {
    Id       int    `json:"id"`
    Fullname string `json:"fullname"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"-"`
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
        &user.Id,
        &user.Fullname,
        &user.Username,
        &user.Email,
    )
    if err != nil {
        return nil, err
    }

    return &user, nil
}
