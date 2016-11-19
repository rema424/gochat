// Entry point.

package main

import "gochat/chat"

var settings = map[string]string{
    // Main port
    "port":       "8080",

    // Logs
    "logMode":    "stdout",     // or "file"
    "logDir":     "logs",       // necessary only if logsMode = "file"
    "logFile":    "debug.log",  // necessary only if logsMode = "file"

    // Static files
    "staticMode": "separate",   // or "self"

    // Database settings
    "dbUser":     "pguser",
    "dbPass":     "123",
    "dbName":     "db_gochat",

    // Redis settings
    "storeProto": "tcp",
    "storeServer": "localhost:6379",
}

func main() {
    chat.RunServer(settings)
}
