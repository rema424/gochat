package main


import (
    "fmt"
    "net/http"
    "html/template"
)


func handlerIndexPage(w http.ResponseWriter, r *http.Request) {
    context := make(map[string]string)
    tpl, _ := template.ParseFiles("templates/index.html")
    tpl.Execute(w, context)
}


func main() {
    port := "8080"
    fmt.Printf("Server is running on %s port...\n", port)

    http.HandleFunc("/", handlerIndexPage)

    http.ListenAndServe(":"+port, nil)
}
