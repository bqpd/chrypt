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
}

func (s socket) Close() error {
    s.done <- true
    return nil
}

func socketHandler(ws *websocket.Conn) {
    fmt.Println("[ws] Starting websocket handler...")
    s := socket{ws, make(chan bool)}
    go match(s)
    <-s.done
    fmt.Println("[ws] ...closing websocket handler.")
}

var partner = make(chan io.ReadWriteCloser)

func match(c io.ReadWriteCloser) {
    fmt.Println("[m] Looking for a match...")
    fmt.Fprint(c, "/sys Waiting for a partner...")
    select {
    case partner <- c:
        // now handled by the other goroutine
    case p := <-partner:
        chat(p, c)
    }
}

func chat(a, b io.ReadWriteCloser) {
    fmt.Println("[m] ...found a match!")
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
    fmt.Println("[http] Starting web handler...")
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
#status { width: 10%; }
#chan {
    text-align: center;
    border: none;
    background: none;
    width: 100%;
    color: white;
    font-weight: 700;
}

#questionbox { 
    background-color: #76DAFF;
    text-align: right;
    padding-top: 1.5em;
    padding-bottom: 0.5em; }
#showq {
    background-color: rgba(255, 255, 255, 0.5);
    padding: 0.5em;
    font-weight: 700;
    font-size: 0.75 em;
    color: rgb(64,64,64);
    cursor: pointer;
    margin-top: -1em;
}
.question { background-color: #aaa;
            padding: 0.5em 0.35em; 
            font-size: 1.5em; }
.question td { color: rgb(64,64,64); }
.closeq { font-size: 1.5em; cursor: pointer; }
.submitq { font-size: 1.5em; color: white; opacity: 0; }
.qbubble {  text-align: left;
            padding: 0.5em 0.5em;
            -moz-border-radius: 0.5em;
            -webkit-border-radius: 0.5em;
            border-radius: 0.5em;
            font-weight: 700; }
.qtxt {     background-color: rgba(255, 255, 255, 0.5);
            -webkit-border-top-right-radius: 0;
            -moz-border-radius-topright: 0;
            border-top-right-radius: 0; }
.answer {   border: none;
            background-color: rgba(64, 64, 64, 0.5);
            color: white;
            -webkit-border-bottom-left-radius: 0;
            -moz-border-radius-bottomleft: 0;
            border-bottom-left-radius: 0;
            width: 100%; }

.inbox {    background-color: white;
            word-wrap: break-word;
            overflow-y: scroll;
            height: 20em;
            padding: 0.5em;
            padding-left: 1.5em;
            text-indent: -1em; }
.inbox p { margin: 0.25em; }
.emote { color: #777; font-style: italic; }
.timer { color: #aaa; }

#textentry { padding: 0.35em;  background-color: white; }
#textentry textarea {   border: 1px solid;
                        width: 100%;
                        border-color: #777 #aaa #aaa #aaa;}

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
            <td id="status"> ● </td>
            <td id="channel"> <input id="chan" onkeyup="connectWS()" /></td>
            <td onclick="ToggleQuestionsDisplay()" id="key">
                <img src="https://dl.dropboxusercontent.com/u/4646709/key3.svg" height=20 />
            </td>
        </tr></table>

        <div id="questionbox" style="display:none">
            <span onclick="showOldQs()" id="showq">show older questions</span><p/>
        </div>

        <div id="detxt" class="inbox"></div>
        <div id="pltxt" class="inbox invisible"></div>

        <div id="textentry"><textarea id="outbox" rows="4"></textarea></div>

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
    
    subd = window.location.hash.slice(1)
    if (subd === "") {
        subd = Math.random().toString(36).slice(2).substring(0,7)
    }
    $('#chan').val(subd)
    window.location.hash = subd

    var wholast,
        conn,
        outbox = $("#outbox"),
        detext = document.getElementById("detxt"),
        pingSound = new Audio('https://dl.dropboxusercontent.com/u/4646709/ping.wav'),
        SwitchViews = function() {
            $(".tab").toggleClass("inactive")
            $(".inbox").toggleClass("invisible") },
        ToggleQuestionsDisplay = function() {
            $("#key").addClass("selected")
            $("#questionbox").slideToggle( function() {
                if ( $('#questionbox').is(':hidden') ) {
                    $("#key").removeClass("selected")
                    $( '.question:not(.selected)' ).hide()
                    $('#showq').text("show older questions")
                }
            })
        },
        showOldQs = function() {
            if ($('#showq').text() === "show older questions") {
                $( '.question:not(.selected)' ).slideDown()
                $('#showq').text("hide older questions")
            } else {
                $( '.question:not(.selected)' ).slideUp()
                $('#showq').text("show older questions")
            }
        }

    outbox.keydown(function (e) {
        if (e.which === 13 && !e.shiftKey) {
            msg = outbox.val()
                if (msg) {
                    addChat("me", encrypt(msg))
                    conn.send(encrypt(msg))
                    outbox.val("") }
            return false } })

    var connectWS = function () {
        if (window["WebSocket"]) {
            var host = location.origin.replace(/^http/, 'ws')
            conn = new WebSocket(host+"/socket/"+$('#chan').val())
            console.log("Connecting to: "+host+"/socket/"+$('#chan').val())
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
            dmsg = dmsg.replace(/\n/g,"<br/>")
            if (dmsg.substring(0,4) === "/me ") {
                dmsg = dmsg.substring(4)
                $("<p class='emote'>"+dmsg+"</p>").appendTo("#detxt")
            } else if (dmsg.substring(0,4) === "/nq ") {
                dmsg = dmsg.substring(4)
                $('<div class="question selected" style="display:none"><table><tr><td class="qtxt qbubble">'+dmsg+'<td class="closeq">&times;<tr><td style="height: 0.35em"><tr><td><input class="answer qbubble"/><td class="submitq">✓</table></div>').appendTo("#questionbox")
                $('.closeq').click( function(e){
                    $(this).parents(".question").slideUp(function() { $(this).remove() }) })
                $('.answer').blur( function(e){ 
                    $(this).parents(".question").removeClass('selected') })
                // Show the questions box
                if (!$("#key").hasClass('selected')) {
                    $('.question.selected').show()
                    ToggleQuestionsDisplay()
                } else {
                    $('.question.selected').slideDown() }
            } else {
                // If it's blank, it's probably an encryption error!
                if (dmsg === '') {
                    dmsg = "<i>The decrypted message is empty, most likely because your answers are not the same.</i>"
                }
                // If it's a regular chat, format it a la gmail
                dmsg = dmsg.replace(/(^| )\*(.+?)\*( |$)/g,"$1<strong>\$2</strong>$3")
                dmsg = dmsg.replace(/(^| )\_(.+?)\_( |$)/g,"$1<em>\$2</em>$3")
                dmsg = dmsg.replace(/(^| )\-(.+?)\-( |$)/g,"$1<del>\$2</del>$3")
                if (who === wholast) {
                   $("#detxt p:last").append("<br/> "+dmsg)
                } else {
                    wholast = who
                    $("<p class='"+who+"'><b>"+who+": </b>"+dmsg+"</p>").appendTo("#detxt")
                }
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

    addChat("me", "/sys Welcome to socially encrypted chat! <br><br> By starting a message with '/nq' you can ask the other person a question. (Try '/nq What is my nickname for you?'). <br><br> The answers to all questions are concatenated, hashed, and used as an encryption key, so if you and the other person have different answers you'll be unable to communicate. You can see what you're actually receiving and sending in the 'ENCRYPTED' tab.")

</script>
</html>
`) }