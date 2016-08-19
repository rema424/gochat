package main


import (
    "fmt"
    "net/http"
    "html/template"

    "github.com/gorilla/websocket"
)


var connections = make(map[*websocket.Conn]bool)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}


func handlerLoginPage(w http.ResponseWriter, r *http.Request) {
    fmt.Println(r.Method)
    if r.Method == "POST" {
        r.ParseForm()
        fmt.Println(r.Form["login"][0])
        fmt.Println(r.Form["password"][0])
    }

    context := make(map[string]string)
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
            fmt.Println("err")
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

    // API
    http.HandleFunc("/ws", handlerWS)
    http.HandleFunc("/login", handlerLoginPage)
    http.HandleFunc("/", handlerIndexPage)

    http.ListenAndServe(":"+port, nil)
}
