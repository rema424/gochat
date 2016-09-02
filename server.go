// URL handlers.

package main

import (
    "log"
    "net/http"
    "html/template"

    "github.com/gorilla/context"
)


func handlerLoginPage(w http.ResponseWriter, r *http.Request) {
    log.Println("hi")
    ctx := make(map[string]string)

    if r.Method == "POST" {
        r.ParseForm()
        login := r.Form["login"][0]
        password := r.Form["password"][0]

        user, err := authenticate(login, password)
        if err != nil {
            ctx["err"] = err.Error()
        } else {
            makeSession(w, user)
            http.Redirect(w, r, "/", 302)
            return
        }
    }

    tpl, _ := template.ParseFiles("templates/login.html", "templates/base.html")
    tpl.ExecuteTemplate(w, "base", ctx)
}


func handlerLogout(w http.ResponseWriter, r *http.Request) {
    removeSession(w, r)
    http.Redirect(w, r, "/login", 302)
    return
}


func handlerIndexPage(w http.ResponseWriter, r *http.Request) {
    // Since this is the default handler for all URLs, we
    // must check if it is correct URL
    if r.URL.Path != "/" {
        http.Error(w, "Not found", 404)
        return
    }

    ctx := make(map[string]string)
    tpl, _ := template.ParseFiles("templates/index.html", "templates/base.html")
    ctx["username"] = context.Get(r, "User").(*User).username
    tpl.ExecuteTemplate(w, "base", ctx)
}
