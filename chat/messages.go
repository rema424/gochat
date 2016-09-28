// Structure and methods for Message objects.

package chat

import (
    "encoding/json"
    "time"
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
    tmp := &struct {
        Date int64 `json:"date"`
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
    if m.Recipient != nil {
        m.Recipient, err = getUserById(m.Recipient.Id)
        if err != nil {
            return err
        }
    }

    return nil
}


// TODO: Add insert/update argument
func (m *Message) save() error {
    var recipientId *int
    if m.Recipient != nil {
        recipientId = &m.Recipient.Id
    }

    _, err := stmtInsertMessage.Exec(
        m.Room.Id, m.Sender.Id, recipientId, m.Text, m.SendDate,
    )
    if err != nil {
        return err
    }

    return nil
}
