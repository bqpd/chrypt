/*  

        GGGGGGGGGGGGG                 
     GGG::::::::::::G                 
   GG:::::::::::::::G                 
  G:::::GGGGGGGG::::G                 
 G:::::G       GGGGGG   ooooooooooo   
G:::::G               oo:::::::::::oo 
G:::::G              o:::::::::::::::o
G:::::G    GGGGGGGGGGo:::::ooooo:::::o
G:::::G    G::::::::Go::::o     o::::o
G:::::G    GGGGG::::Go::::o     o::::o
G:::::G        G::::Go::::o     o::::o
 G:::::G       G::::Go::::o     o::::o
  G:::::GGGGGGGG::::Go:::::ooooo:::::o
   GG:::::::::::::::Go:::::::::::::::o
     GGG::::::GGG:::G oo:::::::::::oo 
        GGGGGG   GGGG   ooooooooooo   

*/

package main

import (
    "os"
    "fmt"
    "io"
    "log"
    "net/http"
    "code.google.com/p/go.net/websocket"
    "sync"
    "path/filepath"
)

func main() {
    http.HandleFunc("/", rootHandler)
    http.Handle("/socket/", websocket.Handler(socketHandler))
    err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
    if err != nil {
        log.Fatal(err)
    }
}

type socket struct {
    io.ReadWriter
    done chan bool
    loc string
    id string
}

func (s socket) Close() error {
    s.done <- true
    return nil
}

var socketmap = make( map[string]chan socket )

var checkingSocketMap = new(sync.Mutex)

func socketHandler(ws *websocket.Conn) {
    loc := ws.Config().Location.String()
    var id string
    websocket.Message.Receive(ws, &id)
    s := socket{ws, make(chan bool), loc, id}
    
    checkingSocketMap.Lock()
    if _, exist := socketmap[loc]; !exist {
        socketmap[loc] = make(chan socket)
    }
    checkingSocketMap.Unlock()

    go match(s)

    <-s.done
    fmt.Println("[ws] closing connection to "+id+" on channel "+loc)
}

func match(c socket) {
    fmt.Println("[ws] "+c.id+" added to channel "+c.loc)
    fmt.Fprint(c, "/sys Waiting for a partner...")
    select {
    case socketmap[c.loc] <- c:
        // now handled by the other goroutine
    case p := <-socketmap[c.loc]:
        if p.id != c.id {
            chat(p, c)
        } else {
            match(c)
        }
    }
}

func chat(a, b socket) {
    fmt.Println("[ws] matched "+a.id+" and "+b.id+" on channel "+a.loc)
    fmt.Fprint(a, "/sys ...we found one!")
    fmt.Fprint(b, "/sys ...we found one!")

    //fmt.Fprint(a, "/sys You are talking to "+b.id)
    //fmt.Fprint(b, "/sys You are talking to "+a.id)

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

func rootHandler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    base := filepath.Base(path)
    isfile, _ := filepath.Match("*.*", base)
    if isfile {
        base = ""
    }

    fmt.Println("[http] serving "+path)

    http.ServeFile(w, r, "./"+base)
}