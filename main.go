// Entry point.

package main

import "gochat/chat"

var settings = map[string]string{
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
}

func main() {
    chat.RunServer(settings)
}
