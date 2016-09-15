// DB-related functions.

package chat

import (
    "fmt"
    "database/sql"

    _ "github.com/lib/pq"
)


func dbConnect(dbUser string, dbPass string, dbName string) (*sql.DB, error) {
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
