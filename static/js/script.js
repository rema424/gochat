var input = document.getElementById('message');
var btn = document.getElementById('btn-send');
var board = document.getElementById('board');

// Display message
function showMessage(msg) {
    board.innerHTML += '<p>' + msg + '</p>';
}

// WebSocket init
var socket = new WebSocket("ws://gochat.local/ws");

// WebSocket events
socket.onopen = function () {
    showMessage('Connected');
};

socket.onclose = function (event) {
    if (event.wasClean) {
        showMessage('Closed clean');
    } else {
        showMessage('Connection lost');
    }
    console.log('Code: ' + event.code + ', reason: ' + event.reason);
};

socket.onmessage = function (event) {
    showMessage('>> ' + event.data);
};

socket.onerror = function (error) {
    showMessage('Error: ' + error.message);
};

// Send message
btn.addEventListener('click', function () {
    var message = input.value;
    socket.send(message);
});
