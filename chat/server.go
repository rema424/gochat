// Entry point.

package chat

import (
    "log"
    "net/http"
    "os"
    "path"

    _ "github.com/lib/pq"
)

// Global storage of hubs (one per room)
// Room.Id: *Hub
var hubs = make(map[int]*Hub)


// Log either to file or to stdout
func setLogOutput(mode string, dir string, file string) *os.File {
    var f *os.File

    if mode == "file" {
        // Make logs dir if it doesn't exist yet
        _, err := os.Stat(dir)
        if os.IsNotExist(err) {
            os.Mkdir(dir, 0700)
        } else if err != nil {
            log.Fatal("Init logging failed:", err)
        }

        // Write logs to file
        f, err = os.OpenFile(
            path.Join(dir, file),
            os.O_RDWR | os.O_CREATE | os.O_APPEND,
            0600,
        )
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

    store = storeConnect(
        settings["storeProto"],
        settings["storeServer"],
    )
    log.Println("In-memory storage connected successfully")
    defer store.Close()

    // Prepare SQL statements
    initStmts()

    // Parse HTML-templates
    initTpls()

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

    // Bind handlers to URLs
    makeRouter()

    // Run server
    log.Printf("Server is running on %s port...\n", settings["port"])
    err = http.ListenAndServe(":"+settings["port"], nil)
    if err != nil {
        log.Fatal("ListenAndServe error:", err)
    }
}
