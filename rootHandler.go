package main

import (
    "fmt"
    "net/http"
   // "html/template"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, `
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; Charset=UTF-8"  />
<LINK rel="stylesheet" href="http://necolas.github.com/normalize.css/2.1.3/normalize.css">
<LINK rel="stylesheet" href="https://dl.dropboxusercontent.com/u/4646709/chrypt.css">
</head>
<body>
    <div id="main">
        <div id="header">
            <div id="key">
            <img src="https://dl.dropboxusercontent.com/u/4646709/key6.svg" height=50 />
            </div>
        </div>
        <div id="questions"></div>
        <div id="detxt" class="textarea"></div>
        <div id="pltxt" class="textarea invisible"></div>
        <div id="textentry"><center><textarea id="input" rows="4" cols="36" autofocus></textarea> </center></div>
        <div id="decrypted" class="tab"> DECRYPTED </div>
        <div id="encrypted" class="tab inactive"> ENCRYPTED </div>
        </div>
</body>
<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.min.js"></script>
<script src="https://dl.dropboxusercontent.com/u/4646709/aes.js"></script>
<script src="https://dl.dropboxusercontent.com/u/4646709/sha3.js"></script>
<script>
    var conn

    if (window["WebSocket"]) {
        var host = location.origin.replace(/^http/, 'ws')
        conn = new WebSocket(host)
            conn.onclose = function (evt) {
                addChat("sys", "/sys The connection has closed.") }
            conn.onmessage = function (evt) {
                addChat("them", evt.data) }
    } else {
        addChat("sys", "/sys Sadly, your browser does not support WebSockets.") }

    var outbox = $("#input")
    
    $(".tab").click(function() {
        $(".tab").toggleClass("inactive")
        $(".textarea").toggleClass("invisible")
    })
    
    $("#header").click(function() {
        $("#questions").toggleClass("invisible")
    })

    var addChat = function(who, cmsg) {
        $("<p class='"+who+"''><b>"+who+": </b>"+cmsg+"</p>").appendTo("#pltxt")
        if (cmsg.substring(0,5) === "/sys ") {
            cmsg = cmsg.substring(5)
            $("<p class='emote'>"+cmsg+"</p>").appendTo("#detxt")
        }
        else {
            dmsg = decrypt(cmsg)
            if (dmsg.substring(0,4) === "/me ") {
                dmsg = dmsg.substring(4)
                $("<p class='emote'>"+dmsg+"</p>").appendTo("#detxt")
            }
            else if (dmsg.substring(0,4) === "/nq ") {
                dmsg = dmsg.substring(4)
                $('<div class="newq">'+dmsg+'<p align="right"><input class="answer" /> </div>').appendTo("#questions")
            }
            else {
                $("<p class='"+who+"'><b>"+who+": </b>"+dmsg+"</p>").appendTo("#detxt")
            }
        }
    }

    $(outbox).keydown(function (e) {
        msg = outbox.val()
        if (e.which == 13 && msg) {
            addChat("me", encrypt(msg))
            conn.send(encrypt(msg))
            outbox.val("")
            return false
        }
    })

    var getSecret = function() {
        var secret = ""
        $(".answer").each(function(i, element) {
            secret = secret + element.value
        });
        return secret
    };

    var encrypt = function(plaintext) {
        var hash = CryptoJS.SHA3(getSecret())
        var encrypted = CryptoJS.AES.encrypt(plaintext, hash.toString())
        return encrypted.toString()
    };

    var decrypt = function(ciphertext) {
        var hash = CryptoJS.SHA3(getSecret())
        var decrypted = CryptoJS.AES.decrypt(ciphertext, hash.toString()).toString(CryptoJS.enc.Latin1)
        return decrypted.toString()
    }

</script>
</html>`) }
