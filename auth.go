// Authentication tools.

package main

import (
    "errors"
    "log"
    "math/rand"
    "net/http"
    "strings"
    "time"

    "github.com/gorilla/context"
)


type User struct {
    Id       int
    FullName string
    Username string
    Email    string
    Password string
}


func makeSessionKey() string {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    key := make([]byte, 64)
    for i := range key {
        key[i] = chars[rand.Intn(len(chars))]
    }
    return string(key)
}


// Make session and set cookie
func makeSession(w http.ResponseWriter, user *User) error {
    key := makeSessionKey()
    exp := time.Now().Add(365 * 24 * time.Hour)

    stmt, err := db.Prepare(`
        INSERT INTO auth_session
        (key, user_id, create_date, expire_date)
        VALUES
        ($1, $2, CURRENT_TIMESTAMP, $3)
    `)
    if err != nil {
        return err
    }
    _, err = stmt.Exec(&key, &user.Id, &exp)
    if err != nil {
        return err
    }

    cookie := http.Cookie{
        Name: "SessionID",
        Value: user.Username + ":" + key,
        Expires: exp,
    }
    http.SetCookie(w, &cookie)

    return nil
}


func removeSession(w http.ResponseWriter, r *http.Request) error {
    cookie, err := r.Cookie("SessionID")
    if err != nil {
        return errors.New("No cookie found")
    }

    session := strings.Split(cookie.Value, ":")
    username := session[0]
    sessionId := session[1]

    stmt, err := db.Prepare(`
        DELETE FROM auth_session
        WHERE
            user_id = (
                SELECT id FROM auth_user WHERE username = $1
            )
            AND key = $2
    `)
    if err != nil {
        return err
    }
    _, err = stmt.Exec(&username, &sessionId)

    cookie = &http.Cookie{
        Name: "SessionID",
        Value: "",
        Expires: time.Now().AddDate(-1, 0, 0), // -1 year
    }
    http.SetCookie(w, cookie)

    return nil
}


// Check user's credentials
func authenticate(username string, password string) (*User, error) {
    stmt, err := db.Prepare(`
        SELECT id, full_name, username, email, password
        FROM auth_user
        WHERE username = $1
    `)
    if err != nil {
        return nil, err
    }

    var user User
    err = stmt.QueryRow(username).Scan(
        &user.Id,
        &user.FullName,
        &user.Username,
        &user.Email,
        &user.Password,
    )
    if err != nil {
        return nil, err
    }

    if user.Password == password {
        return &user, nil
    } else {
        return nil, errors.New("Login or password incorrect")
    }
}


// Check session cookie
func checkSession(r *http.Request) (*User, error) {
    cookie, err := r.Cookie("SessionID")
    if err != nil {
        return nil, errors.New("No cookie found")
    }

    session := strings.Split(cookie.Value, ":")
    username := session[0]
    sessionId := session[1]

    stmt, err := db.Prepare(`
        SELECT u.id, u.full_name, u.username, u.email, u.password
        FROM auth_session AS s
        LEFT JOIN auth_user AS u ON u.id = s.user_id
        WHERE u.username = $1
            AND s.key = $2
            AND s.expire_date > CURRENT_TIMESTAMP
    `)
    if err != nil {
        return nil, err
    }

    var user User
    err = stmt.QueryRow(username, sessionId).Scan(
        &user.Id,
        &user.FullName,
        &user.Username,
        &user.Email,
        &user.Password,
    )
    if err != nil {
        return nil, err
    }

    return &user, nil
}


// Middleware for authentication
func authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Check auth
        user, err := checkSession(r)

        // User is not authenticated
        if err != nil {
            log.Println("Check session error:", err)
            http.Redirect(w, r, "/login", 302)
            return
        }

        context.Set(r, "User", user)
        handler(w, r)
    }
}
