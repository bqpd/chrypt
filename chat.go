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
    fmt.Fprint(a, "/sys ...we found one!")
    fmt.Fprint(b, "/sys ...we found one!")
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
<link rel="stylesheet" href="https://raw.github.com/necolas/normalize.css/master/normalize.css">
<style>

body {  color: #222;
        font-family: Arial, arial;
        background-color: #707070;
        font-size: 12.8px; }
#chatwindow {   margin: 0 auto;
        width: 20em;
        border-top: 0.5px solid white; }
table { width: 100%; 
        text-align: center;}

#header {   text-align: center;
            background-color: rgb(64,64,64);
            color: white;
            font-size: 2em; }
#key { width: 30%; cursor: pointer; }
#status { width: 30%; }
#channel { color: #aaa; }

.question { font-size: 1.5em;
            color: rgb(64,64,64);
            font-weight: 700;
            background-color: #aaa;
            padding: 1em;
            padding-bottom: 0.25em;}
.answer {   border: none;
            background-color: rgba(255, 255, 255, 0.5);
            padding: 0.5em;
            color: rgb(64,64,64);
            -webkit-border-radius: 0.5em;
            -webkit-border-bottom-right-radius: 0;
            -moz-border-radius: 0.5em;
            -moz-border-radius-bottomright: 0;
            border-radius: 0.5em;
            border-bottom-right-radius: 0; }

.textarea { background-color: white;
            word-wrap: break-word;
            overflow-y: scroll;
            height: 20em;
            padding: 0.5em;
            padding-left: 1.5em;
            text-indent: -1em; }
.textarea p { margin: 0.25em; }
.emote { color: #777; font-style: italic; }
.timer { color: #aaa; }

#textentry {    text-align: center;
                background-color: white;
                padding: 0.25em; }

.tab {  font-size: 0.75em;
        padding: 0.25em;
        background-color: white;
        color: #777;
        border-bottom: 1px solid #aaa; }
.inactive { background-color: #aaa;
            color: white;
            cursor: pointer; }

.invisible { display: none; }
.selected { background-color: #76DAFF; }

</style>
<title> chrypt </title>
</head>
<body>
    <div id="chatwindow">

        <table id="header"><tr>
            <td onclick="ToggleQuestionsDisplay()" id="key">
                <img src="https://dl.dropboxusercontent.com/u/4646709/key3.svg" height=20 />
            </td>
            <td id="channel"> </td>
            <td id="status"> ● </td>
        </tr></table>

        <div id="questionbox"></div>

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
<script src="https://crypto-js.googlecode.com/svn/tags/3.1.2/build/rollups/aes.js"></script>
<script src="https://crypto-js.googlecode.com/svn/tags/3.1.2/build/rollups/sha3.js"></script>
<script>

    var conn,
        outbox = $("#outbox"),
        detext = document.getElementById("detxt"),
        pingSound = new Audio('https://dl.dropboxusercontent.com/u/4646709/ping.wav'),
        SwitchViews = function() {
            $(".tab").toggleClass("inactive")
            $(".textarea").toggleClass("invisible") },
        ToggleQuestionsDisplay = function() {
            $("#key").toggleClass("selected")
            $("#questionbox").toggleClass("invisible") }

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
            if (cmsg === "Waiting for a partner...") {
                $("#status").css({color: "yellow"})  }
            else if (cmsg === "...we found one!") {
                $("#status").css({color: "green"})  }
            else if (cmsg === "The connection has closed.") {
                $("#status").css({color: "red"})  }
            else {
                console.log(cmsg)
                $("<p class='emote'>"+cmsg+"</p>").appendTo("#detxt")  }
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
                $('<div class="question selected">'+dmsg+'<p align="right"><input class="answer" /></div>').appendTo("#questionbox")
                // Show the questions box
                $("#key").addClass("selected")
                $("#questionbox").removeClass("invisible")
            } else {
                // If it's a regular chat, format it a la gmail
                dmsg = dmsg.replace(/(^| )\*(.+?)\*( |$)/g,"$1<strong>\$2</strong>$3")
                dmsg = dmsg.replace(/(^| )\_(.+?)\_( |$)/g,"$1<em>\$2</em>$3")
                dmsg = dmsg.replace(/(^| )\-(.+?)\-( |$)/g,"$1<del>\$2</del>$3")
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
        var decrypted = CryptoJS.AES.decrypt(ciphertext, hash.toString()).toString(CryptoJS.enc.Utf8)
        return decrypted.toString() }

    addChat("sys", "/sys Welcome to socially encrypted chat! <br><br>&nbsp; 1) By starting a message with '/nq' you can ask the other person a question. (Try '/nq What is my nickname for you?'). The answers to all of your questions are concatenated, hashed, and used as an encryption key, so if you and the other person have different answers you'll be unable to communicate! <br><br>&nbsp; 2) You can see what you're actually receiving and sending in the 'ENCRYPTED' tab.")

</script>
</html>
`) }