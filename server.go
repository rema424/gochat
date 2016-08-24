// URL handlers.

package main

import (
    "time"
    "net/http"
    "html/template"
)


func handlerLoginPage(w http.ResponseWriter, r *http.Request) {
    context := make(map[string]string)

    if r.Method == "POST" {
        r.ParseForm()
        login := r.Form["login"][0]
        password := r.Form["password"][0]

        user, ok := users[login]
        if ok && user.Password == password {
            cookie := http.Cookie{
                Name: "SessionID",
                Value: login + ":abc123",
                Expires: time.Now().Add(365 * 24 * time.Hour),
            }
            http.SetCookie(w, &cookie)
            http.Redirect(w, r, "/", 302)
            return
        } else {
            context["err"] = "Login or password incorrect"
        }
    }

    tpl, _ := template.ParseFiles("templates/login.html")
    tpl.Execute(w, context)
}


func handlerLogout(w http.ResponseWriter, r *http.Request) {
    removeSessionCookie(w)
    http.Redirect(w, r, "/login", 302)
    return
}


func handlerIndexPage(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Error(w, "Not found", 404)
        return
    }

    context := make(map[string]string)
    tpl, _ := template.ParseFiles("templates/index.html")
    tpl.Execute(w, context)
}
