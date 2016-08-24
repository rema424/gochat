// Authentication tools.

package main

import (
    "errors"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/gorilla/context"
)


type User struct {
    FullName string
    Username string
    Password string
}

// Maps instead of DB (for testing)
var users = map[string]User{
    "user1": {
        Username: "user1",
        FullName: "User No.1",
        Password: "pass1",
    },
    "user2": {
        Username: "user2",
        FullName: "User No.2",
        Password: "pass2",
    },
}
var sessions = map[string][]string{
    "user1": {"abc123", "def456"},
}


func removeSessionCookie(w http.ResponseWriter) {
    cookie := &http.Cookie{
        Name: "SessionID",
        Value: "",
        Expires: time.Now().AddDate(-1, 0, 0), // -1 year
    }
    http.SetCookie(w, cookie)
}


// Check session cookie
func authenticate(r *http.Request) (*User, error) {
    emptyUser := &User{
        FullName: "",
        Username: "",
        Password: "",
    }

    cookie, err := r.Cookie("SessionID")
    if err != nil {
        return emptyUser, errors.New("No cookie found")
    }

    session := strings.Split(cookie.Value, ":")
    username := session[0]
    sessionId := session[1]

    currentSessions, ok := sessions[username]
    if !ok {
        return emptyUser, errors.New("No session found")
    }

    for _, s := range currentSessions {
        // Session found = user is authenticated
        if sessionId == s {
            user := users[username]
            return &user, nil
        }
    }

    return emptyUser, errors.New("No session found")
}


// Middleware for authentication
// func authMiddleware(w http.ResponseWriter, r *http.Request, hub *Hub) {
func authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Check auth
        user, err := authenticate(r)

        // User is not authenticated
        if err != nil {
            log.Println(err)
            removeSessionCookie(w)
            http.Redirect(w, r, "/login", 302)
            return
        }

        context.Set(r, "User", user)
        handler(w, r)
    }
}
