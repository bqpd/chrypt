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