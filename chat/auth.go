// Authentication tools.

package chat

import (
    "database/sql"
    "errors"
    "log"
    "math/rand"
    "net/http"
    "time"

    "github.com/gorilla/context"
    "golang.org/x/crypto/bcrypt"
)


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

    // Delete other user's sessions (only 1 session per user)
    _, err := stmtDeleteUserSessions.Exec(user.Id)
    if err != nil {
        return err
    }

    _, err = stmtMakeSession.Exec(&key, &user.Id, &exp)
    if err != nil {
        return err
    }

    cookie := http.Cookie{
        Name: "SessionID",
        Value: key,
        Expires: exp,
    }
    http.SetCookie(w, &cookie)

    return nil
}


// Remove session and clear cookie
func removeSession(w http.ResponseWriter, r *http.Request) error {
    cookie, err := r.Cookie("SessionID")
    if err != nil {
        return errors.New("No cookie found")
    }

    sessionId := cookie.Value
    _, err = stmtDeleteSession.Exec(sessionId)

    // Clear cookie
    cookie = &http.Cookie{
        Name: "SessionID",
        Value: "",
        Expires: time.Now().AddDate(-1, 0, 0), // -1 year
    }
    http.SetCookie(w, cookie)

    return nil
}


// Check session cookie and get user from database
func getUserFromSession(r *http.Request) (*User, error) {
    cookie, err := r.Cookie("SessionID")
    if err != nil {
        return nil, errors.New("No cookie found")
    }

    sessionId := cookie.Value
    var user User
    err = stmtGetUserBySession.QueryRow(sessionId).Scan(
        &user.Id, &user.Fullname, &user.Username, &user.Email,
    )
    if err == sql.ErrNoRows {
        return nil, errors.New("No session found")
    } else if err != nil {
        return nil, err
    }

    return &user, nil
}


// Check user's credentials
func authenticate(username string, password string) (*User, error) {
    var user User
    var hash string
    err := stmtGetUserByUsername.QueryRow(username).Scan(
        &user.Id,
        &user.Fullname,
        &user.Username,
        &user.Email,
        &hash,
    )
    if err == sql.ErrNoRows {
        return nil, errors.New("Login or password incorrect")
    } else if err != nil {
        return nil, err
    }

    err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err == nil {
        return &user, nil
    } else {
        return nil, errors.New("Login or password incorrect")
    }
}


// Middleware for authentication
func authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user, err := getUserFromSession(r)

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
