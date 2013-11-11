package main

import (
    "fmt"
    "net/http"
    "html/template"
)

func main() {
    http.HandleFunc("/", rootHandler)
    fmt.Println("listening...")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
      panic(err)
    }
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    rootTemplate.Execute(w, ":8000")
}

var rootTemplate = template.Must(template.New("root").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script>
    websocket = new WebSocket("ws://{{.}}/socket");
    websocket.onmessage = onMessage;
    websocket.onclose = onClose;
</html>
`))
