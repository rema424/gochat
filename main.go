// Entry point.

package main

import "gochat/chat"

var settings = map[string]string{
    // Database settings
    "dbUser":     "postgres",
    "dbPass":     "postgres",
    "dbName":     "db_gochat",
}

func main() {
    chat.RunServer(settings)
}
