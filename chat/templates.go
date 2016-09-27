// Parsed HTML-templates.

package chat

import (
    "html/template"
    "log"
)

var (
    tplLogin *template.Template
    tplChat  *template.Template
    tplIndex *template.Template
)


func parseTpl(tpls ...string) *template.Template {
    tpl, err := template.ParseFiles(tpls...)
    if err != nil {
        log.Fatal("Could not parse template:", err)
    }
    return tpl
}


func initTpls() {
    tplLogin = parseTpl(
        "templates/login.html",
        "templates/base.html",
    )
    tplChat = parseTpl(
        "templates/chat.html",
        "templates/base.html",
    )
    tplIndex = parseTpl(
        "templates/index.html",
        "templates/base.html",
    )
}
