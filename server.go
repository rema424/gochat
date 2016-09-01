// URL handlers.

package main

import (
    "fmt"
    "net/http"
    "html/template"
)


func handlerLoginPage(w http.ResponseWriter, r *http.Request) {
    context := make(map[string]string)

    if r.Method == "POST" {
        r.ParseForm()
        login := r.Form["login"][0]
        password := r.Form["password"][0]

        user, err := authenticate(login, password)
        if err != nil {
            context["err"] = err.Error()
        } else {
            makeSession(w, user)
            http.Redirect(w, r, "/", 302)
            return
        }
    }

    tpl, _ := template.ParseFiles("templates/login.html")
    tpl.Execute(w, context)
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

    context := make(map[string]string)
    tpl, err := template.ParseFiles("templates/index.html", "templates/base.html")
    if err != nil {
        fmt.Println("Hello,  world")
    }
    tpl.ExecuteTemplate(w, "base", context)
}
