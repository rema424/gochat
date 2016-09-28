// URL handlers for AJAX requests.

package chat

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"

    "github.com/gorilla/context"
    "github.com/gorilla/mux"
)


// Return all users from chat
func handlerAjaxGetRoomUsers(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)

    // No need to check errors - regex in mux controlls input
    roomId, _ := strconv.Atoi(vars["id"])

    hub := hubs[roomId]
    users := hub.room.getUsers()

    resp, err := json.Marshal(users)
    if err != nil {
        log.Println("JSON encoding error:", err)
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(resp)
}


// Last <n> messages for current user
func handlerAjaxGetRoomMessages(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    user := context.Get(r, "User").(*User)
    n := 10  // number of last messages
    var err error

    // Get number of messages from query
    q, ok := r.URL.Query()["number"]
    if ok && fmt.Sprintf("%T", q) == "[]string" && len(q) > 0 {
        n, err = strconv.Atoi(q[0])
        if err != nil {
            n = 10  // ignore errors, just use default value
        }
    }

    // No need to check errors - regex in mux controlls input
    roomId, _ := strconv.Atoi(vars["id"])
    room := hubs[roomId].room

    messages, err := room.getMessages(user, n)
    if err != nil {
        log.Println(err)
        return
    }

    resp, _ := json.Marshal(messages)

    w.Header().Set("Content-Type", "application/json")
    w.Write(resp)
}
