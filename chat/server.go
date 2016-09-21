// Entry point.

package chat

import (
    "log"
    "net/http"
    "os"
    "database/sql"

    _ "github.com/lib/pq"
)

// Global connection to DB
var db *sql.DB

// Global storage of hubs (one per room)
// Room.Id: *Hub
var hubs = make(map[int]*Hub)


// Log either to file or to stdout
func setLogOutput(mode string, dir string, file string) (*os.File, error) {
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
            return nil, err
        }
        log.SetOutput(f)
    } else if mode == "stdout" {
        log.SetOutput(os.Stdout)
    }

    return f, nil
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
    f, err := setLogOutput(
        settings["logMode"],
        settings["logDir"],
        settings["logFile"],
    )
    if err != nil {
        panic(err.Error())
    } else if f != nil {
        defer f.Close()
    }

    // Static files
    setStaticMode(settings["staticMode"])

    // DB connect (using global variable)
    db, err = dbConnect(
        settings["dbUser"],
        settings["dbPass"],
        settings["dbName"],
    )
    if err != nil {
        panic(err.Error())
    } else {
        log.Println("DB connected successfully")
        defer db.Close()
    }

    // Messages exchanging (websockets) for each room
    rooms, err := getAllRooms()
    if err != nil {
        panic(err.Error())
    }

    for _, room := range rooms {
        hub := makeHub(&room)
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
        log.Println("ListenAndServe error: ", err)
    }
}
