package main

import (
    "os"
    "fmt"
    "io"
    "log"
    "net/http"

    "code.google.com/p/go.net/websocket"
)

const listenAddr = ":"+os.Getenv("PORT")

func main() {
    fmt.Println("Starting web handler...")
    http.HandleFunc("/", rootHandler)
    fmt.Println("Starting websocket handler...")
    http.Handle("/socket", websocket.Handler(socketHandler))
    err := http.ListenAndServe(listenAddr, nil)
    if err != nil {
        log.Fatal(err)
    }
}

type socket struct {
    io.ReadWriter
    done chan bool
}

func (s socket) Close() error {
    s.done <- true
    return nil
}
 github.com/kr/godep
func socketHandler(ws *websocket.Conn) {
    s := socket{ws, make(chan bool)}
    go match(s)
    <-s.done
    fmt.Println("Closing websocket handler...")
}

var partner = make(chan io.ReadWriteCloser)

func match(c io.ReadWriteCloser) {
    fmt.Println("Looking for a match...")
    fmt.Fprint(c, "/sys Waiting for a partner...")
    select {
    case partner <- c:
        // now handled by the other goroutine
    case p := <-partner:
        chat(p, c)
    }
}

func chat(a, b io.ReadWriteCloser) {
    fmt.Println("Found a match!")
    fmt.Fprintln(a, "/sys ...we found one!")
    fmt.Fprintln(b, "/sys ...we found one!")
    errc := make(chan error, 1)
    go cp(a, b, errc)
    go cp(b, a, errc)
    if err := <-errc; err != nil {
        log.Println(err)
    }
    a.Close()
    b.Close()
}

func cp(w io.Writer, r io.Reader, errc chan<- error) {
    _, err := io.Copy(w, r)
    errc <- err
}

