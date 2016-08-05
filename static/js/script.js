// WebSocket init
var socket = new WebSocket("ws://gochat.local/ws");

socket.onopen = function () {
    console.log('Connected');
};

socket.onclose = function (event) {
    if (event.wasClean) {
        console.log('Closed clean');
    } else {
        console.log('Connection lost');
    }
    console.log('Code: ' + event.code + ', reason: ' + event.reason);
};

socket.onmessage = function (event) {
    console.log('New data: ' + event.data);
};

socket.onerror = function (error) {
    console.log('Error: ' + error.message);
};


// Interface
var input = document.getElementById('message');
var btn = document.getElementById('btn-send');

btn.addEventListener('click', function () {
    var message = input.value;
    console.log('click');
    socket.send('message');
});
