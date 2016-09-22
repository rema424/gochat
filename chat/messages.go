// Structure and methods for Message objects.

package chat

import (
    "encoding/json"
    "time"
    "database/sql"
)


type Message struct {
    Id        int       `json:"id"`
    Action    string    `json:"action"`
    Sender    *User     `json:"sender"`
    Recipient *User     `json:"recipient"`
    Text      string    `json:"text"`
    SendDate  time.Time `json:"send_date"`
    Room      *Room     `json:"-"`
}


// Custom marshaler: convert date to int64 (timestamp)
func (m *Message) MarshalJSON() ([]byte, error) {
    type Alias Message
    return json.Marshal(&struct {
        SendDate int64 `json:"send_date"`
        *Alias
    }{
        SendDate: m.SendDate.Unix(),
        Alias:    (*Alias)(m),
    })
}

// Custom unmarshaller: add current timestamp as date,
// find recipient by id in database
func (m *Message) UnmarshalJSON(data []byte) error {
    type Alias Message

    // Temporary structure for marshalling JSON
    tmp := &struct {
        SendDate  int64      `json:"send_date"`
        Recipient *struct {
            Id    int        `json:"id"`
        }                    `json:"recipient"`
        *Alias
    }{
        Alias: (*Alias)(m),
    }

    err := json.Unmarshal(data, &tmp);
    if err != nil {
        return err
    }

    // Use UTC for saving
    m.SendDate = time.Now().UTC()

    // Get full recipient info from database
    if tmp.Recipient != nil {
        var recipient User
        stmt, err := db.Prepare(`
            SELECT id, username, full_name, email, role
            FROM auth_user
            WHERE id = $1
        `)
        if err != nil {
            return err
        }

        err = stmt.QueryRow(tmp.Recipient.Id).Scan(
            &recipient.Id,
            &recipient.Username,
            &recipient.Fullname,
            &recipient.Email,
            &recipient.Role,
        )
        if err == sql.ErrNoRows {
            return nil
        } else if err != nil {
            return err
        }

        m.Recipient = &recipient
    }

    return nil
}


// TODO: Add insert/update argument
func (m *Message) save() error {
    stmt, err := db.Prepare(`
        INSERT INTO message
        (room_id, sender_id, recipient_id, text, send_date)
        VALUES
        ($1, $2, $3, $4, $5)
    `)
    if err != nil {
        return err
    }

    var recipientId *int
    if m.Recipient != nil {
        recipientId = &m.Recipient.Id
    }

    _, err = stmt.Exec(
        m.Room.Id, m.Sender.Id, recipientId, m.Text, m.SendDate,
    )
    if err != nil {
        return err
    }

    return nil
}
