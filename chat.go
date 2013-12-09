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
    fmt.Println("[http] serving "+r.URL.String())
    if r.URL.String() != "/favico.ico" {
        fmt.Fprint(w, `
<html>
<meta http-equiv="Content-Type" content="text/html; Charset=UTF-8"  />
<head>
<link rel="stylesheet" href="https://dl.dropboxusercontent.com/u/4646709/normalize.css">
<style>
/*

        cccccccccccccccc    ssssssssss       ssssssssss   
      cc:::::::::::::::c  ss::::::::::s    ss::::::::::s  
     c:::::::::::::::::css:::::::::::::s ss:::::::::::::s 
    c:::::::cccccc:::::cs::::::ssss:::::ss::::::ssss:::::s
    c::::::c     ccccccc s:::::s  ssssss  s:::::s  ssssss 
    c:::::c                s::::::s         s::::::s      
    c:::::c                   s::::::s         s::::::s   
    c::::::c     cccccccssssss   s:::::s ssssss   s:::::s 
    c:::::::cccccc:::::cs:::::ssss::::::ss:::::ssss::::::s
     c:::::::::::::::::cs::::::::::::::s s::::::::::::::s 
      cc:::::::::::::::c s:::::::::::ss   s:::::::::::ss  
        cccccccccccccccc  sssssssssss      sssssssssss     

*/

body {  color: #222;
        font-family: Arial, arial;
        background-color: #707070;
        font-size: 12.8px; }
#chatwindow {   margin: 0 auto;
        max-width: 20em; }
table { width: 100%; 
        text-align: center; }
a { color: #0065cc; }

#header {   text-align: center;
            background-color: rgb(64,64,64);
            color: white;
            font-size: 2em; }
#key { width: 30%; cursor: pointer; }
#status {   width: 10%;
            padding: 0.25em 0.5em;
            cursor: default; }
#chan {
    text-align: center;
    border: none;
    background: none;
    width: 80%;
    color: white;
    font-weight: 700; }

#questionbox { 
    background-color: #76DAFF;
    padding-top: 1.5em;
    padding-bottom: 0.5em; }
.qbutton {
    background-color: rgba(255, 255, 255, 0.5);
    padding: 0.5em;
    font-weight: 700;
    font-size: 0.75 em;
    color: rgb(64,64,64);
    cursor: pointer;
    margin-top: -1em; }
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
            min-height: 10em;
            height: 23em;
            padding: 0.5em;
            padding-left: 1.5em;
            text-indent: -1em; }
.inbox p { margin: 0.25em; }
.emote { color: #777; font-style: italic; }
.notice { color: #aaa; }

#textentry { padding: 0.35em;  background-color: white; text-align: center; color: #aaa; }
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
<!-- 

    hhhhhhh                     tttt                                  lllllll 
    h:::::h                  ttt:::t                                  l:::::l 
    h:::::h                  t:::::t                                  l:::::l 
    h:::::h                  t:::::t                                  l:::::l 
     h::::h hhhhh      ttttttt:::::ttttttt       mmmmmmm    mmmmmmm    l::::l 
     h::::hh:::::hhh   t:::::::::::::::::t     mm:::::::m  m:::::::mm  l::::l 
     h::::::::::::::hh t:::::::::::::::::t    m::::::::::mm::::::::::m l::::l 
     h:::::::hhh::::::htttttt:::::::tttttt    m::::::::::::::::::::::m l::::l 
     h::::::h   h::::::h     t:::::t          m:::::mmm::::::mmm:::::m l::::l 
     h:::::h     h:::::h     t:::::t          m::::m   m::::m   m::::m l::::l 
     h:::::h     h:::::h     t:::::t          m::::m   m::::m   m::::m l::::l 
     h:::::h     h:::::h     t:::::t    ttttttm::::m   m::::m   m::::m l::::l 
     h:::::h     h:::::h     t::::::tttt:::::tm::::m   m::::m   m::::ml::::::l
     h:::::h     h:::::h     tt::::::::::::::tm::::m   m::::m   m::::ml::::::l
     h:::::h     h:::::h       tt:::::::::::ttm::::m   m::::m   m::::ml::::::l
     hhhhhhh     hhhhhhh         ttttttttttt  mmmmmm   mmmmmm   mmmmmmllllllll

-->
    <div id="chatwindow">

        <table id="header">
            <td id="status"> ●
            <td id="channel"> <span id="chan" />
            <td id="key" onclick="ToggleQuestionsDisplay()">
                <img src="https://dl.dropboxusercontent.com/u/4646709/key3.svg" height=20 />
        </table>

        <div id="questionbox" style="display:none">
            <span onclick="startNewQ()" class="qbutton" style="float: left">+ new question</span>
            <span onclick="showOldQs()" id="showq" class="qbutton" style="float: right">show old questions</span>
            <br>
        </div>

        <div id="detxt" class="inbox"></div>
        <div id="pltxt" class="inbox invisible"></div>

        <div id="textentry"><textarea id="outbox" rows="4"></textarea>
        </div>

        <table id="tabs">
            <td onclick="SwitchViews()" class="tab"> DECRYPTED
            <td onclick="SwitchViews()" class="tab inactive"> ENCRYPTED
        </table>

    </div>
</body>

<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.min.js"></script>
<script src="https://crypto-js.googlecode.com/svn/tags/3.1.2/build/rollups/aes.js"></script>
<script src="https://crypto-js.googlecode.com/svn/tags/3.1.2/build/rollups/sha3.js"></script>
<script>
/*

             jjjj                  
            j::::j                 
             jjjj                  
                                   
           jjjjjjj    ssssssssss   
           j:::::j  ss::::::::::s  
            j::::jss:::::::::::::s 
            j::::js::::::ssss:::::s
            j::::j s:::::s  ssssss 
            j::::j   s::::::s      
            j::::j      s::::::s   
            j::::jssssss   s:::::s 
            j::::js:::::ssss::::::s
            j::::js::::::::::::::s 
            j::::j s:::::::::::ss  
            j::::j  sssssssssss    
            j::::j                 
  jjjj      j::::j                 
 j::::jj   j:::::j                 
 j::::::jjj::::::j                 
  jj::::::::::::j                  
    jjj::::::jjj                   
       jjjjjj

*/

/* TODO *\
on running connectWS, set exponentially growing timeout, with a 'retry' link popping up somewhere
send keep-alive messages every 30s or so?
prevent identical clients from connecting to each other
\*      */

    /*
     * Event-driven functions *
                              */
    var SwitchViews = function() {
            $(".tab").toggleClass("inactive")
            $(".inbox").toggleClass("invisible")
        },

    ToggleQuestionsDisplay = function() {
        $("#key").addClass("selected")
        $("#questionbox").slideToggle( function() {
            if ( $("#questionbox").is(":hidden") ) {
                $("#key").removeClass("selected")
                $(".question:not(.selected)").hide()
                $('#showq').text("show old questions")
            }
        } )
    },

    startNewQ = function() {
        $("#outbox").val("/nq Type your question here.")
        $("#outbox").focus()
        $("#outbox")[0].setSelectionRange(4,28)
    },

    showOldQs = function() {
        switch ( $('#showq').text().substring(0,4) ) {
            case "show":
                $( '.question:not(.selected)' ).slideDown()
                $('#showq').text("hide old questions")
                break
            case "hide":
                $( '.question:not(.selected)' ).slideUp()
                $('#showq').text("show old questions") } }, 

    outbox = $("#outbox")
    outbox.focus( function() { document.title = "chrypt" } )
    outbox.keydown( function(e) {
        if (e.which === 13 && !e.shiftKey) {
            msg = outbox.val()
                if (msg) {
                    if (msg.substr(0,5) === "/plt ") {
                        addChat("me", msg)
                        conn.send(msg)
                    } else {
                        addChat("me", encrypt(msg))
                        conn.send(encrypt(msg))

                    }
                    outbox.val("")
                }
            return false
        }
    } )
    
    $("#questionbox").delegate(".closeq", "click", function() {
        $(this).parents(".question").slideUp( function() { $(this).remove() } )
    })

    $("#questionbox").delegate(".answer", "blur", function(){ 
        $(this).parents(".question").removeClass("selected")
    })

    /*
     * Websocket interfacing *
                             */
    var subd = location.href.replace(/^.*\//, "")
    // Are we at a hashed subdomain? If not, make one up!
    if ( subd === '') {
        subd = Math.random().toString(36).slice(2).substring(7,13)
        location.href = location.origin+'/'+subd
    }
    $('#chan').text(subd)

    var last_connection_time = 0,
        connecting = false,
        host = location.origin.replace(/^http/, "ws"),
        conn,
        client_id = Math.random().toString(36).slice(2).substring(0,13),

    connectWS = function() {
        if (window["WebSocket"]) {
            conn = new WebSocket(host+"/socket/"+subd)
            setTimeout('conn.send(client_id)', 1000)
            conn.onclose = function (evt) {
                addChat("me", "/sys The connection has closed.")
                delayedConnect() }
            conn.onmessage = function (evt) {
                addChat("them", evt.data) }
        } else {
            addChat("me", "/sys Sadly, your browser does not support WebSockets.")
        }
        connecting = false
    },

    delayedConnect = function() {
        if ( !connecting ) {
            var time = new Date().getTime()
            if ( time > last_connection_time ) {
                var delay = Math.max(last_connection_time+5000, time) - time
                last_connection_time = time + delay
                connecting = true
                setTimeout( connectWS, delay )
            }
        }
    }

    /*
     * Message parsing and display *
                                   */
    var statuses = { "Waiting for a partner..." : "yellow",
                     "...we found one!" : "green",
                     "The connection has closed." : "red" },

    ciphertext_interpreters = {
        "/sys ": function(who, cmsg) {
            if ( statuses[cmsg] ) {
                $("#status").css({"color": statuses[cmsg]}) }
            else {
                console.log(cmsg)
                $("<p class='emote'>"+cmsg+"</p>").appendTo("#detxt")  }
        },

        "/ntc ": function(who, cmsg) {
            $("<p class='notice'>"+cmsg+"</p>").appendTo("#detxt")
        },

        "/plt ": function(who, cmsg) {
            $("<p class='"+who+"'style='color: red'><b>"+who+":  (not encrypted) <br></b>"+cmsg+"</p>").appendTo("#detxt")
            detext.scrollTop = detext.scrollHeight;
            if ( !document.hasFocus() ) {
                // Is the user away? Alert them!
                pingSound.play()
                document.title = "(new) chrypt" }
        } }, 
    
    deciphered_text_interpreters = {
        "/me " : function(who, dmsg) {
            $("<p class='emote'>"+dmsg+"</p>").appendTo("#detxt")
        },

        "/nq " : function(who, dmsg) {
            $("<div class='question selected' style='display:none'><table><tr><td class='qtxt qbubble'>"+dmsg+"<td class='closeq'>&times;<tr><td style='height: 0.35em'><tr><td><input class='answer qbubble'/><td class='submitq'>✓</table></div>").appendTo("#questionbox")
            // Show the questions box
            if (!$("#key").hasClass("selected")) {
                $(".question.selected").show()
                ToggleQuestionsDisplay()
            } else {
                $(".question.selected").slideDown() }
        } },

    pingSound = new Audio("https://dl.dropboxusercontent.com/u/4646709/ping.wav"),
    messages_missed = 0,
    previous_speaker,
    detext = document.getElementById("detxt"),

    addChat = function(who, cmsg) {
        // Print the exact message received
        $("<p class='"+who+"''><b>"+who+": </b>"+cmsg+"</p>").appendTo("#pltxt")

        // If there's a matching ciphertext interpreter, run it and skip the rest.
        if ( ciphertext_interpreters[cmsg.substring(0,5)] ) {
            ciphertext_interpreters[cmsg.substring(0,5)](who, cmsg.substring(5))
        } else {
            // Decrypt and translate newlines.
            dmsg = decrypt(cmsg)
            dmsg = dmsg.replace(/\n/g,"<br/>")

            if ( !document.hasFocus() ) {
            // Is the user away? Alert them!
                pingSound.play()
                document.title = "(new) chrypt" }

            // If there's a matching deciphered-text interpreter, use that.
            if ( deciphered_text_interpreters[dmsg.substring(0,4)] ) {
                deciphered_text_interpreters[dmsg.substring(0,4)](who, dmsg.substring(4))
            } else {
                // Is it a blank message? Probably a mismatch between answers.
                if ( dmsg === "" ) {
                    messages_missed++
                    if (messages_missed === 1) {
                        dmsg = "_The decrypted message is empty, most likely because your answers are not the same._"
                    } else if (messages_missed === 3) {
                        dmsg = "_You can send a plaintext message by starting it with '/plt'. If you're having trouble coordinating answers, try deleting all questions and starting over. Giving them hints is not only bad security, it's less fun!_"
                    } else {
                        dmsg = "<br> _Another empty decrypted message._ <br>"
                    }
                }
                // Format it a la gmail
                dmsg = dmsg.replace(/(^| )(http.+?)( |$)/g,"$1<a href='$2'>$2</a>$3")
                dmsg = dmsg.replace(/(^| )\*(.+?)\*( |$)/g,"$1<strong>\$2</strong>$3")
                dmsg = dmsg.replace(/(^| )\_(.+?)\_( |$)/g,"$1<em>\$2</em>$3")
                dmsg = dmsg.replace(/(^| )\-(.+?)\-( |$)/g,"$1<del>\$2</del>$3")

                // Don't sign the message if the previous one was 'normal' and from the same person
                if ($("#detxt p:last").hasClass(who) && $("#detxt p:last").hasClass('normal')) {
                   $("#detxt p:last").append("<br> "+dmsg)
                } else {
                    $("<p class='"+who+" normal'><b>"+who+": </b>"+dmsg+"</p>").appendTo("#detxt")
                }
            }
        // After everything's written, scroll the textbox to the bottom
        detext.scrollTop = detext.scrollHeight;
        }
    }

    /*
     * Crypto helper functions *
                               */
    var last_singlehash,
        last_thousandhash,

    updateHashes = function() {
        if (last_singlehash !== getSecret(1)) {
            addChat('me', '/ntc Encryption key was changed')
            last_singlehash = getSecret(1)
            last_thousandhash = getSecret(1000)
        }
    },

    getSecret = function(hashes) {
            var secret = ""
            $(".answer").each(function(i, element) {
                secret = secret + element.value })

            var hash = CryptoJS.SHA3(secret).toString()
            for (var i = hashes; i > 0; i--) {
                hash =  CryptoJS.SHA3(hash).toString() }
            return hash
        },

        encrypt = function(plaintext) {
            updateHashes()
            return CryptoJS.AES.encrypt(plaintext, last_thousandhash).toString()
        },
        
        decrypt = function(ciphertext) {
            updateHashes()
            return CryptoJS.AES.decrypt(ciphertext, last_thousandhash).toString(CryptoJS.enc.Utf8).toString()
        }

    /*
     * Runtime *
               */

    addChat("me", "/sys Welcome to socially encrypted chat! <br><br> By starting a message with '/nq' you can ask the other person a question. (Try '/nq What is my nickname for you?'). <br><br> The answers to all questions are concatenated, hashed, and used as an encryption key, so if your answers are different you'll be unable to communicate. <br><br> You can see exactly what you're sending and receiving in the 'ENCRYPTED' tab.")

    last_singlehash = getSecret(1)
    last_thousandhash = getSecret(1000)

    addChat("me", "/sys Connecting to: "+host+"/socket/"+subd)
    
    $(document).ready(connectWS)

</script>
</html>
`) } }