package main

import (
    "os"
    "fmt"
    "io"
    "log"
    "net/http"
    "code.google.com/p/go.net/websocket"
)

func main() {
    fmt.Println("Starting web handler...")
    http.HandleFunc("/", rootHandler)
    fmt.Println("Starting websocket handler...")
    http.Handle("/socket", websocket.Handler(socketHandler))
    err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
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

func rootHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, `
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; Charset=UTF-8"  />
<link rel="stylesheet" href="http://necolas.github.com/normalize.css/2.1.3/normalize.css">
<link rel="stylesheet" href="https://dl.dropboxusercontent.com/u/4646709/chrypt.css">
<title> social secret chat </title>
</head>
<body>
    <div id="chatwindow">

        <table id="header"><tr>
            <td onclick="ToggleQuestionsDisplay()" class="key">
                <img src="https://dl.dropboxusercontent.com/u/4646709/key3.svg" height=20 />
            </td>
            <td id="channel"> socket </td>
        </tr></table>

        <div id="questions"></div>

        <div id="detxt" class="textarea"></div>
        <div id="pltxt" class="textarea invisible"></div>

        <div id="textentry"><textarea id="outbox" rows="4" cols="36"></textarea></div>

        <table id="tabs"><tr>
            <td onclick="SwitchViews()" class="tab"> DECRYPTED </td>
            <td onclick="SwitchViews()" class="tab inactive"> ENCRYPTED </td>
        </tr></table>

    </div>
</body>
<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.min.js"></script>
<script src="https://dl.dropboxusercontent.com/u/4646709/aes.js"></script>
<script src="https://dl.dropboxusercontent.com/u/4646709/sha3.js"></script>
<script>

    var conn,
        outbox = $("#outbox"),
        detext = document.getElementById("detxt"),
        pingSound = new Audio('https://dl.dropboxusercontent.com/u/4646709/ping.wav'),
        SwitchViews = function() {
            $(".tab").toggleClass("inactive")
            $(".textarea").toggleClass("invisible") },
        ToggleQuestionsDisplay = function() {
            $(".key").toggleClass("selected")
            $("#questions").toggleClass("invisible") }

    outbox.keydown(function (e) {
        if (e.which === 13) {
            msg = outbox.val()
                if (msg) {
                    addChat("me", encrypt(msg))
                    conn.send(encrypt(msg))
                    outbox.val("") }
            return false } })

    var connectWS = function () {
        if (window["WebSocket"]) {
            var host = location.origin.replace(/^http/, 'ws')
            conn = new WebSocket(host+"/socket")
                conn.onclose = function (evt) {
                    addChat("me", "/sys The connection has closed.")
                    connectWS() }
                conn.onmessage = function (evt) {
                    addChat("them", evt.data) }
        } else {
            addChat("me", "/sys Sadly, your browser does not support WebSockets.") } }

    connectWS()

    outbox.focus( function() { document.title = "social secret chat" })

    var addChat = function(who, cmsg) {
        // Print the exact message received
        $("<p class='"+who+"''><b>"+who+": </b>"+cmsg+"</p>").appendTo("#pltxt")

        // Is it a system plaintext message?
        if (cmsg.substring(0,5) === "/sys ") {
            cmsg = cmsg.substring(5)
            $("<p class='emote'>"+cmsg+"</p>").appendTo("#detxt")
        } else {
            // If it's not, play a sound and decrypt it.
            if ( !document.hasFocus() ) {
                pingSound.play()
                document.title = "(new) social secret chat" }

            dmsg = decrypt(cmsg)
            if (dmsg.substring(0,4) === "/me ") {
                dmsg = dmsg.substring(4)
                $("<p class='emote'>"+dmsg+"</p>").appendTo("#detxt")
            } else if (dmsg.substring(0,4) === "/nq ") {
                dmsg = dmsg.substring(4)
                $('<div class="question newq">'+dmsg+'<p align="right"><input class="answer" /></div>').appendTo("#questions")
                // Show the questions box
                $(".key").addClass("selected")
                $("#questions").removeClass("invisible")
            } else {
                // If it's a regular chat, format it a la gmail
                dmsg = dmsg.replace(/\*(.+?)\*/g,"<strong>\$1</strong>")
                dmsg = dmsg.replace(/\_(.+?)\_/g,"<em>\$1</em>")
                dmsg = dmsg.replace(/\-(.+?)\-/g,"<del>\$1</del>")
                $("<p class='"+who+"'><b>"+who+": </b>"+dmsg+"</p>").appendTo("#detxt")
            }

        // Scroll the textbox to the bottom
        detext.scrollTop = detext.scrollHeight;
        }
    }

    var getSecret = function() {
        var secret = ""
        $(".answer").each(function(i, element) {
            secret = secret + element.value })
        return secret }

    var encrypt = function(plaintext) {
        var hash = CryptoJS.SHA3(getSecret())
        var encrypted = CryptoJS.AES.encrypt(plaintext, hash.toString())
        return encrypted.toString() }

    var decrypt = function(ciphertext) {
        var hash = CryptoJS.SHA3(getSecret())
        var decrypted = CryptoJS.AES.decrypt(ciphertext, hash.toString()).toString(CryptoJS.enc.Latin1)
        return decrypted.toString() }

</script>
</html>
`) }