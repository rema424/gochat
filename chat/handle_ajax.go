// URL handlers for AJAX requests.

package chat

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gorilla/context"
)


// Return all users from chat
func handlerAjaxUsersList(w http.ResponseWriter, r *http.Request, hub *Hub) {
    users := []*User{}
    for c := range(hub.clients) {
        users = append(users, c.user)
    }

    resp, err := json.Marshal(users)

    if err != nil {
        log.Println("JSON encoding error", err)
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(resp)
}


// Last 10 messages for current user
func handlerAjaxGetLastMessages(w http.ResponseWriter, r *http.Request) {
    user := context.Get(r, "User").(*User)

    messages, err := getLastMessages(user, 10)
    // TODO: Error message for client
    if err != nil {
        log.Println(err)
        return
    }

    resp, _ := json.Marshal(messages)

    w.Header().Set("Content-Type", "application/json")
    w.Write(resp)
}
