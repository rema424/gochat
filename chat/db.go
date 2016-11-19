// DB-related functions.

package chat

import (
    "fmt"
    "log"
    "database/sql"

    _ "github.com/lib/pq"
    "github.com/garyburd/redigo/redis"
)

var (
    // Global connection to DB
    db *sql.DB
    // Global in-memory storage connection
    store redis.Conn

    // Users
    stmtGetUserById           *sql.Stmt
    stmtGetUserByUsername     *sql.Stmt
    // Authentication
    stmtMakeSession           *sql.Stmt
    stmtGetUserBySession      *sql.Stmt
    stmtDeleteSession         *sql.Stmt
    stmtDeleteUserSessions    *sql.Stmt
    // Rooms
    stmtGetAllRooms           *sql.Stmt
    stmtGetUserRooms          *sql.Stmt
    stmtGetUserRoomInfo       *sql.Stmt
    // Room administration (mute, ban)
    stmtCheckMute             *sql.Stmt
    stmtMute                  *sql.Stmt
    stmtUnmute                *sql.Stmt
    stmtCheckBan              *sql.Stmt
    stmtBan                   *sql.Stmt
    // Messages
    stmtInsertMessage         *sql.Stmt
    stmtGetRoomMessagesByUser *sql.Stmt
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
    stmtGetUserRooms = prepareStmt(db, `
        SELECT r.id, r.name,
            CASE WHEN b.id IS NOT NULL THEN true
            ELSE false
            END AS is_banned
        FROM room AS r
        LEFT JOIN ban AS b ON b.room_id = r.id AND b.user_id = $1
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
    stmtBan = prepareStmt(db, `
        INSERT INTO ban (user_id, room_id)
        VALUES ($1, $2)
    `)

    // Messages in room: for the user, from the user or broadcast
    stmtGetRoomMessagesByUser = prepareStmt(db, `
        SELECT *
        FROM (
            SELECT
                m.id, 'message', m.text, m.send_date,
                us.id, us.username, us.full_name, us.email, us_rn.name AS role,
                ur.id, ur.username, ur.full_name, ur.email, ur_rn.name AS role
            FROM message AS m
            -- Sender
            LEFT JOIN auth_user AS us ON us.id = m.sender_id
            LEFT JOIN room_role AS us_rr ON us_rr.room_id = m.room_id AND us_rr.user_id = us.id
            LEFT JOIN role_name AS us_rn ON us_rr.role_id = us_rn.id
            -- Recipient
            LEFT JOIN auth_user AS ur ON ur.id = m.recipient_id
            LEFT JOIN room_role AS ur_rr ON ur_rr.room_id = m.room_id AND ur_rr.user_id = ur.id
            LEFT JOIN role_name AS ur_rn ON ur_rr.role_id = ur_rn.id
            WHERE
                (
                    (m.recipient_id = $1 OR m.recipient_id IS NULL)
                    OR m.sender_id = $1
                )
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

func storeConnect(proto string, srv string) redis.Conn {
    c, err := redis.Dial(proto, srv)
    if err != nil {
        log.Fatal("In-memory store connection failed:", err)
    }
    return c
}
