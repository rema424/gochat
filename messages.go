// Structure and methods for Message objects.

package main

import (
    "errors"
    "log"
)


type Message struct {
    Id        int       `json:"id"`
    Role      string    `json:"role"`
    Sender    *User     `json:"sender"`
    Recipient *User     `json:"recipient"`
    Text      string    `json:"text"`
    SendDate  int64     `json:"send_date"`  // timestamp
}


// Save message to database
func (m *Message) save() error {
    // Prohibit empty messages from users
    if m.Text == "" {
        return errors.New("Text cannot be empty")
    }

    log.Println(m.Sender.Username+": "+m.Text)

    stmt, err := db.Prepare(`
        INSERT INTO message
        (sender_id, recipient_id, text, send_date)
        VALUES
        ($1, $2, $3, to_timestamp($4))
    `)
    if err != nil {
        return err
    }
    _, err = stmt.Exec(
        &m.Sender.Id, nil,
        &m.Text, &m.SendDate)
    if err != nil {
        return err
    }

    return nil
}


func getLastMessages(user *User, limit int) ([]Message, error) {
    var messages []Message

    stmt, err := db.Prepare(`
        SELECT *
        FROM (
            SELECT u.id, u.username, u.full_name, u.email,
                m.id, 'message', m.text, EXTRACT(epoch FROM m.send_date)::int AS send_date,
                m.recipient_id IS NULL
            FROM message AS m
            LEFT JOIN auth_user AS u ON u.id = m.sender_id
            WHERE m.recipient_id = $1 OR m.recipient_id IS NULL
            ORDER BY m.send_date DESC
            LIMIT $2
        ) AS tmp
        ORDER BY send_date ASC
    `)
    if err != nil {
        return []Message{}, err
    }

    rows, err := stmt.Query(user.Id, limit)
    if rows != nil {
        defer rows.Close()
    } else {
        return []Message{}, nil
    }

    var msg Message
    var sender User
    var isBroadcast bool

    for rows.Next() {
        err = rows.Scan(
            &sender.Id, &sender.Username, &sender.Fullname, &sender.Email,
            &msg.Id, &msg.Role, &msg.Text, &msg.SendDate, &isBroadcast)
        if err != nil {
            return []Message{}, err
        }
        msg.Sender = &sender
        if !isBroadcast {
            msg.Recipient = user
        }

        messages = append(messages, msg)
    }

    return messages, nil
}
