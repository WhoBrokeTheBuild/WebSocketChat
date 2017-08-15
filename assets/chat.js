$(function () {
    // Readability
    const MILLISECONDS_PER_SECOND = 1000;

    // DOM Lookups
    var ws;
    var name = getName();
    var $container = $("#container");

    var messenger = document.getElementById("messenger");
    var host = messenger.dataset.host;
    var messageLine = document.getElementById("ml");
    var messageLineName = ml.content.querySelector("#ml-n");
    var messageLineTime = ml.content.querySelector("#ml-t");
    var messageLineMessage = ml.content.querySelector("#ml-m");

    if (window.WebSocket === undefined) {
        $container.append("Your browser does not support WebSockets");
        console.error("WebSockets not supported. Aborting...")
        return;
    } else {
        ws = initWS();
    }

    function initWS() {
        var socket = new WebSocket("ws://"+host+"/socket");
        socket.onopen = function() {
            ws.send(JSON.stringify({name: name}));
            $container.append("<p>Socket is open</p>");
        };
        socket.onmessage = function(e) {
            displayMessage(e.data);
        };
        socket.onclose = function (ce) {
            $container.append("<p>Socket closed</p>");
            $container.append(ce.code);
            $container.append(ce.reason);
        }
        return socket;
    }

    $("#messenger").submit(function (e) {
        e.preventDefault();
        ws.send(JSON.stringify({ message: $("#message").val() }));
        displayMessage(JSON.stringify({ 
            message: $("#message").val(),
            name: name,
            time: Date.now() / MILLISECONDS_PER_SECOND | 0
        }));
        $("#message").val("");
        return false;
    });
    function getName() {
        var name = Cookies.get('name'); // => 'value'
        if (typeof name != "undefined"){
        return name;
        }
        do{
        name = prompt("My name's Catbug! What's yours?", "");
        } while (name == null || name == "");
        alert("HAI " + name + "! You're my new friend now. We're having soft tacos later!")
        Cookies.set('name', name, { expires: 365 });
        return name;
    }

    function displayMessage(data) {
        if('content' in document.createElement('template')) {
            data = JSON.parse(data);
            messageLineMessage.textContent = data.message;
            messageLineTime.textContent = new Date(data.time * MILLISECONDS_PER_SECOND).toLocaleTimeString();
            messageLineName.textContent = data.name;
            var clone = document.importNode(messageLine.content, true);
            container.appendChild(clone);
        } else {
            $container.append(data) + "Please upgrade from Internet Explorer 5";
        }
    }
});
