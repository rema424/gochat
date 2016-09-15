// DB-related functions.

package main

import (
    "fmt"
    "database/sql"

    _ "github.com/lib/pq"
)


func dbConnect() (*sql.DB, error) {
    const (
        dbUser = "pguser"
        dbPass = "123"
        dbName = "db_gochat"
    )
    var err error

    dbConnection := fmt.Sprintf("user=%s password=%s dbname=%s", dbUser, dbPass, dbName)
    db, err = sql.Open("postgres", dbConnection)
    if err != nil {
        return nil, err
    }
    err = db.Ping()
    if err != nil {
        return nil, err
    }

    return db, nil
}
