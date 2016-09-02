var $input = $('#input-message');
var $btn = $('#btn-send');
var $board = $('#board');

// Display message
function showMessage(msg) {
    $board.html($board.html() + '<p>' + msg + '</p>');
    $board.scrollTop($board.prop('scrollHeight'));
}

function formatMessage(text, cls) {
    return '<span class="msg-' + cls + '">' + text + '</span>';
}

// WebSocket init
var socket = new WebSocket("ws://gochat.local/ws");

// WebSocket events
socket.onopen = function () {
    showMessage(formatMessage('Connected', 'info'));
};

socket.onclose = function (event) {
    msg = '(code: ' + event.code + ', reason:' + event.reason + ')';
    if (event.wasClean) {
        msg = 'Closed clean ' + msg;
    } else {
        msg = 'Connection lost ' + msg;
    }
    showMessage(formatMessage(msg, 'error'));
};

socket.onmessage = function (event) {
    var msg = JSON.parse(event.data);
    var msgString;
    if (msg.recipient) {
        msgString =
            formatMessage(msg.date + ', ', 'date') +
            formatMessage(msg.sender, 'sender') +
            formatMessage(' TO ', 'to') +
            formatMessage(msg.recipient + ': ', 'recipient') +
            formatMessage(msg.text, 'text')
    } else {
        msgString =
            formatMessage(msg.date + ', ', 'date') +
            formatMessage(msg.sender + ': ', 'sender') +
            formatMessage(msg.text, 'text')
    }
    showMessage(msgString);
};

socket.onerror = function (error) {
    showMessage(formatMessage('Error: ' + error.message, 'error'));
};

// Send message
$btn.on('click', function () {
    event.preventDefault();
    var message = $input.val();
    socket.send(message);
    $input.val('');
});
$input.on('keypress', function (event) {
    if (event.keyCode == 13) {
        event.preventDefault();
        var message = $input.val();
        socket.send(message);
        $input.val('');
    }
});