// DB-related functions.

package chat

import (
    "fmt"
    "log"
    "database/sql"

    _ "github.com/lib/pq"
)

var (
    // Global connection to DB
    db *sql.DB

    // Users
    stmtGetUserById        *sql.Stmt
    stmtGetUserByUsername  *sql.Stmt
    stmtUpdateUser         *sql.Stmt
    // Authentication
    stmtMakeSession        *sql.Stmt
    stmtGetUserBySession   *sql.Stmt
    stmtDeleteSession      *sql.Stmt
    stmtDeleteUserSessions *sql.Stmt
    // Rooms
    stmtGetAllRooms        *sql.Stmt
    stmtGetUserRoomInfo    *sql.Stmt
    // Room administration (mute, ban)
    stmtCheckMute          *sql.Stmt
    stmtMute               *sql.Stmt
    stmtUnmute             *sql.Stmt
    stmtCheckBan           *sql.Stmt
    // Messages
    stmtInsertMessage      *sql.Stmt
    stmtGetMessages        *sql.Stmt
)


func prepareStmt(db *sql.DB, query string) *sql.Stmt {
    stmt, err := db.Prepare(query)
    if err != nil {
        log.Fatal("Could not prepare '"+query+"': "+err.Error())
    }
    return stmt
}


func initStmts() {
    // Users
    stmtGetUserById = prepareStmt(db, `
        SELECT id, full_name, username, email
        FROM auth_user
        WHERE id = $1
    `)
    stmtGetUserByUsername = prepareStmt(db, `
        SELECT id, full_name, username, email, password
        FROM auth_user
        WHERE username = $1
    `)
    stmtUpdateUser = prepareStmt(db, `
        UPDATE auth_user
        SET
            full_name = $2,
            username = $3,
            email = $4
        WHERE id = $1
    `)

    // Authentication
    stmtMakeSession = prepareStmt(db, `
        INSERT INTO auth_session
        (key, user_id, create_date, expire_date)
        VALUES
        ($1, $2, CURRENT_TIMESTAMP, $3)
    `)
    stmtGetUserBySession = prepareStmt(db, `
        SELECT u.id, u.full_name, u.username, u.email
        FROM auth_session AS s
        LEFT JOIN auth_user AS u ON u.id = s.user_id
        WHERE s.key = $1
            AND s.expire_date > CURRENT_TIMESTAMP
    `)
    stmtDeleteSession = prepareStmt(db, `
        DELETE FROM auth_session
        WHERE key = $1
    `)
    stmtDeleteUserSessions = prepareStmt(db, `
        DELETE FROM auth_session
        WHERE user_id = $1
    `)

    // Rooms
    stmtGetAllRooms = prepareStmt(db, `
        SELECT id, name
        FROM room
    `)
    stmtGetUserRoomInfo = prepareStmt(db, `
        SELECT
            CASE
                WHEN rn.name IS NULL THEN 'user'
                ELSE rn.name
            END AS role,
            m.date
        FROM room AS r
        LEFT JOIN room_role AS rr ON rr.room_id = r.id AND rr.user_id = $1
        LEFT JOIN role_name AS rn ON rr.role_id = rn.id
        LEFT JOIN mute AS m ON m.room_id = r.id AND m.user_id = $1
        WHERE r.id = $2
    `)

    // Room administration (mute, ban)
    stmtCheckMute = prepareStmt(db, `
        SELECT EXISTS(
            SELECT 1
            FROM mute
            WHERE user_id = $1 AND room_id = $2
        )
    `)
    stmtMute = prepareStmt(db, `
        INSERT INTO mute (user_id, room_id)
        VALUES ($1, $2)
    `)
    stmtUnmute = prepareStmt(db, `
        DELETE FROM mute
        WHERE user_id = $1 AND room_id = $2
    `)
    stmtCheckBan = prepareStmt(db, `
        SELECT EXISTS(
            SELECT 1
            FROM ban
            WHERE user_id = $1 AND room_id = $2
        )
    `)

    // Messages
    stmtGetMessages = prepareStmt(db, `
        SELECT *
        FROM (
            SELECT u.id, u.username, u.full_name, u.email, rn.name AS role,
                m.id, 'message', m.text, m.send_date,
                m.recipient_id IS NULL
            FROM message AS m
            LEFT JOIN auth_user AS u ON u.id = m.sender_id
            LEFT JOIN room_role AS rr ON rr.room_id = m.room_id AND rr.user_id = u.id
            LEFT JOIN role_name AS rn ON rr.role_id = rn.id
            WHERE (m.recipient_id = $1 OR m.recipient_id IS NULL)
                AND m.room_id = $2
            ORDER BY m.send_date DESC
            LIMIT $3
        ) AS tmp
        ORDER BY send_date ASC
    `)
    stmtInsertMessage = prepareStmt(db, `
        INSERT INTO message
        (room_id, sender_id, recipient_id, text, send_date)
        VALUES
        ($1, $2, $3, $4, $5)
    `)
}


func dbConnect(dbUser string, dbPass string, dbName string) *sql.DB {
    var err error

    dbConnection := fmt.Sprintf("user=%s password=%s dbname=%s", dbUser, dbPass, dbName)
    db, err = sql.Open("postgres", dbConnection)
    if err != nil {
        log.Fatal("DB connection failed:", err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatal("DB ping failed:", err)
    }

    return db
}
