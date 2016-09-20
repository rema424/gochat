// Bind URLs to handle functions.

package chat


import (
    "net/http"
)


func makeRouter(h *Hub) {
    // AJAX
    http.HandleFunc(
        "/ajax/users",
        authMiddleware(func(w http.ResponseWriter, r *http.Request) {
            handlerAjaxUsersList(w, r, h)
        }),
    )
    http.HandleFunc(
        "/ajax/messages/last",
        authMiddleware(handlerAjaxGetLastMessages),
    )

    // WS
    http.HandleFunc(
        "/ws",
        authMiddleware(func(w http.ResponseWriter, r *http.Request) {
            handlerWS(w, r, h)
        }),
    )

    // Pages
    http.HandleFunc(
        "/login",
        handlerLoginPage,
    )
    http.HandleFunc(
        "/logout",
        authMiddleware(handlerLogout),
    )
    http.HandleFunc(
        "/",
        authMiddleware(handlerIndexPage),
    )
}