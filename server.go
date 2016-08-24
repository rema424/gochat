package main

import (
    "fmt"
    "errors"
    "strings"
    "time"
    "net/http"
    "html/template"

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
            fmt.Println(err)
            removeSessionCookie(w)
            http.Redirect(w, r, "/login", 302)
            return
        }

        context.Set(r, "User", user)
        handler(w, r)
    }
}


func handlerLoginPage(w http.ResponseWriter, r *http.Request) {
    context := make(map[string]string)

    if r.Method == "POST" {
        r.ParseForm()
        login := r.Form["login"][0]
        password := r.Form["password"][0]

        user, ok := users[login]
        if ok && user.Password == password {
            cookie := http.Cookie{
                Name: "SessionID",
                Value: login + ":abc123",
                Expires: time.Now().Add(365 * 24 * time.Hour),
            }
            http.SetCookie(w, &cookie)
            http.Redirect(w, r, "/", 302)
            return
        } else {
            context["err"] = "Login or password incorrect"
        }
    }

    tpl, _ := template.ParseFiles("templates/login.html")
    tpl.Execute(w, context)
}


func handlerLogout(w http.ResponseWriter, r *http.Request) {
    removeSessionCookie(w)
    http.Redirect(w, r, "/login", 302)
    return
}


func handlerIndexPage(w http.ResponseWriter, r *http.Request) {
    context := make(map[string]string)
    tpl, _ := template.ParseFiles("templates/index.html")
    tpl.Execute(w, context)
}
