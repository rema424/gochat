// Entry point.

package chat

import (
    "log"
    "net/http"
    "os"
    "database/sql"

    _ "github.com/lib/pq"
)

// Settings
const (
    // Logs
    logsMode = "file"        // or "stdout"
    logDir = "logs"          // necessary only if logsMode = "file"
    logFile = "debug.log"    // necessary only if logsMode = "file"

    // Static files
    staticMode = "separate"  // or "self"

    // Database settings
    dbUser = "pguser"
    dbPass = "123"
    dbName = "db_gochat"
)

// Global connection to DB
var db *sql.DB


// Log either to file or to stdout
func setLogOutput(mode string) (*os.File, error) {
    var f *os.File

    if mode == "file" {
        // Make logs dirs if it's not already exists
        _, err := os.Stat("logs")
        if os.IsNotExist(err) {
            os.Mkdir("logs", 0700)
        }
        // Write logs to file
        f, err = os.OpenFile("logs/debug.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0600)
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


func RunServer() {
    var err error

    // Logs
    f, err := setLogOutput(logsMode)
    if err != nil {
        panic(err.Error())
    } else if f != nil {
        defer f.Close()
    }

    // Static files
    setStaticMode(logsMode)

    // DB connect (using global variable)
    db, err = dbConnect(dbUser, dbPass, dbName)
    if err != nil {
        panic(err.Error())
    } else {
        log.Println("DB connected successfully")
        defer db.Close()
    }

    // Messages exchanging (websockets)
    hub := makeHub()
    go hub.run()

    // Bind routes to URLs
    makeRouter(hub)

    // Run server
    port := "8080"
    log.Printf("Server is running on %s port...\n", port)
    err = http.ListenAndServe(":"+port, nil)
    if err != nil {
        log.Println("ListenAndServe error: ", err)
    }
}
