// URL handlers for AJAX requests.

package main

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gorilla/context"
)


// Self user info
func handlerAjaxUserSelf(w http.ResponseWriter, r *http.Request, hub *Hub) {
    user := context.Get(r, "User").(*User)

    resp, err := json.Marshal(user)
    if err != nil {
        log.Println("JSON encoding error", err)
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(resp)
}


// Return all usernames
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