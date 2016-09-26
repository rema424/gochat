// Entry point.

package chat

import (
    "log"
    "net/http"
    "os"

    _ "github.com/lib/pq"
)

// Global storage of hubs (one per room)
// Room.Id: *Hub
var hubs = make(map[int]*Hub)


// Log either to file or to stdout
func setLogOutput(mode string, dir string, file string) *os.File {
    var f *os.File

    if mode == "file" {
        // Make logs dirs if it's not already exists
        _, err := os.Stat(dir)
        if os.IsNotExist(err) {
            os.Mkdir(dir, 0700)
        }
        // Write logs to file
        f, err = os.OpenFile(dir+"/"+file, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0600)
        if err != nil {
            log.Fatal("Init logging failed:", err)
        }
        log.SetOutput(f)
    } else if mode == "stdout" {
        log.SetOutput(os.Stdout)
    }

    return f
}


// Static files served either by Nginx or by Go FileServer
func setStaticMode(mode string) {
    if mode == "self" {
        fs := http.Dir("static")
        fileHandler := http.FileServer(fs)
        http.Handle("/static/", http.StripPrefix("/static/", fileHandler))
    } else if mode == "separate" {
        // do nothing (static files are served by Nginx)
    }
}


func RunServer(settings map[string]string) {
    var err error

    // Logs
    f := setLogOutput(
        settings["logMode"],
        settings["logDir"],
        settings["logFile"],
    )
    if f != nil {
        defer f.Close()
    }

    // Static files
    setStaticMode(settings["staticMode"])

    // DB connect (using global variable)
    db = dbConnect(
        settings["dbUser"],
        settings["dbPass"],
        settings["dbName"],
    )
    log.Println("DB connected successfully")
    defer db.Close()

    // Prepare SQL statements
    initStmts()

    // Run messages exchanging (websockets) for each room
    rooms, err := getAllRooms()
    if err != nil {
        log.Fatal("Could not get the list of rooms")
    }
    for _, room := range rooms {
        hub := makeHub(room)
        hubs[room.Id] = hub
        go hub.run()
    }

    // Bind routes to URLs
    makeRouter()

    // Run server
    port := "8080"
    log.Printf("Server is running on %s port...\n", port)
    err = http.ListenAndServe(":"+port, nil)
    if err != nil {
        log.Fatal("ListenAndServe error:", err)
    }
}
