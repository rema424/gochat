package main


import (
    "fmt"
    "errors"
    "strings"
    "time"

    "net/http"
    "html/template"

    "github.com/gorilla/websocket"
)


var connections = make(map[*websocket.Conn]bool)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type user struct {
    Name     string
    Password string
}

// Maps instead of DB (for testing)
var users = map[string]user{
    "user1": {
        Name: "User No.1",
        Password: "pass1",
    },
    "user2": {
        Name: "User No.2",
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


// Check session cookie and return user if authenticated
func isLoggedIn(r *http.Request) (bool, error) {
    cookie, err := r.Cookie("SessionID")
    if err != nil {
        return false, errors.New("No cookie found")
    }

    session := strings.Split(cookie.Value, ":")
    username := session[0]
    sessionId := session[1]

    currentSessions, ok := sessions[username]
    if !ok {
        return false, errors.New("No session found")
    }

    for _, s := range currentSessions {
        // Session found = user is authenticated
        if sessionId == s {
            return true, nil
        }
    }

    return false, errors.New("No session found")
}


func handlerAuth(w http.ResponseWriter, r *http.Request) {
    url := r.URL.Path

    // Check auth
    logged, err := isLoggedIn(r)

    // No auth for login page
    if url == "/login" && !logged {
        if logged {
            fmt.Println("Already logged in")
            http.Redirect(w, r, "/", 302)
        } else {
            handlerLoginPage(w, r)
        }
        return
    }

    // User is not authenticated
    if err != nil {
        fmt.Println(err)
        http.Redirect(w, r, "/login", 302)
        return
    }

    switch url {
    case "/ws":
        handlerWS(w, r)
    default:
        handlerIndexPage(w, r)
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


func handlerIndexPage(w http.ResponseWriter, r *http.Request) {
    context := make(map[string]string)
    tpl, _ := template.ParseFiles("templates/index.html")
    tpl.Execute(w, context)
}


func handlerWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("Successfully connected")
    connections[conn] = true

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            fmt.Println(err)
            delete(connections, conn)
            conn.Close()
            return
        }
        fmt.Println(string(msg))
        sendAll(msg)
    }
}


func sendAll(msg []byte) {
    for conn := range connections {
        err := conn.WriteMessage(websocket.TextMessage, msg)
        if err != nil {
            delete(connections, conn)
            conn.Close()
        }
    }
}


func main() {
    port := "8080"
    fmt.Printf("Server is running on %s port...\n", port)

    // Static files
    // fs := http.Dir("static")
    // fileHandler := http.FileServer(fs)
    // http.Handle("/static/", http.StripPrefix("/static/", fileHandler))

    http.HandleFunc("/", handlerAuth)
    http.ListenAndServe(":"+port, nil)
}
