// Entry point.

package chat

import (
    "fmt"
    "log"
    "net/http"
    "database/sql"

    _ "github.com/lib/pq"
)

// Global DB connection
var (
    db                     *sql.DB
    stmtGetUserFromSession *sql.Stmt
)


func prepareStmt(db *sql.DB, query string) *sql.Stmt {
    stmt, err := db.Prepare(query)
    if err != nil {
        log.Fatal("Could not prepare '" + query + "': " + err.Error())
    }
    return stmt
}


func initStmt() {
    stmtGetUserFromSession = prepareStmt(db, `
        SELECT u.id, u.full_name, u.username, u.email
        FROM auth_session AS s
        LEFT JOIN auth_user AS u ON u.id = s.user_id
        WHERE s.key = $1 AND s.expire_date > CURRENT_TIMESTAMP
    `)
}


func RunServer(settings map[string]string) {
    var err error

    // Connect to DB
    dbConnection := fmt.Sprintf(
        "user=%s password=%s dbname=%s",
        settings["dbUser"],
        settings["dbPass"],
        settings["dbName"],
    )
    db, err = sql.Open("postgres", dbConnection)
    if err != nil {
        panic(err.Error())
    }
    err = db.Ping()
    if err != nil {
        panic(err.Error())
    }
    log.Println("DB connected successfully")
    defer db.Close()

    initStmt()

    // Bind routes to URLs
    http.HandleFunc("/login", handlerLoginPage)
    http.HandleFunc("/logout", authMiddleware(handlerLogout))
    http.HandleFunc("/", authMiddleware(handlerIndexPage))

    // Run server
    port := "8080"
    log.Printf("Server is running on %s port...\n", port)
    err = http.ListenAndServe(":"+port, nil)
    if err != nil {
        log.Println("ListenAndServe error: ", err)
    }
}
