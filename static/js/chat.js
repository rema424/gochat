var currentUser = {}
var $input = $('#input-message');
var $btn = $('#btn-send');
var $board = $('#board');
var $userlist = $('#input-recipient')

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

    // Get self info
    $.getJSON({
        url: '/ajax/users/self',
        dataType: 'json',
        success: function(response) {
            currentUser = response;
        },
        error: function(response) {
            console.log(response);
        }
    });

    // Get userlist
    $.getJSON({
        url: '/ajax/users',
        dataType: 'json',
        success: function(response) {
            response.forEach(function (user) {
                $userlist.append(
                    $('<option></option>')
                        .attr('value', user.id)
                        .text(user.username)
                );
            })
        },
        error: function(response) {
            console.log(response);
        }
    });
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

    if (currentUser &&
        msg.sender != currentUser.username &&
        msg.role == 'new_user'
    ) {
        $userlist.append(
            $('<option></option>')
                .attr('value', msg.sender.id)
                .text(msg.sender.username)
        );
        msgString = formatMessage(msg.text, 'info');
        showMessage(msgString);
    } else if (msg.role == 'message') {
        var date = new Date(msg.send_date * 1000);
        date = date.getHours() + ':' + date.getMinutes();
        if (msg.recipient) {
            msgString =
                formatMessage(date + ', ', 'date') +
                formatMessage(msg.sender.username, 'sender') +
                formatMessage(' TO ', 'to') +
                formatMessage(msg.recipient.username + ': ', 'recipient') +
                formatMessage(msg.text, 'text');
        } else {
            msgString =
                formatMessage(date + ', ', 'date') +
                formatMessage(msg.sender.username + ': ', 'sender') +
                formatMessage(msg.text, 'text');
        }
        showMessage(msgString);
    }
};

socket.onerror = function (error) {
    showMessage(formatMessage('Error: ' + error.message, 'error'));
};

// Send message
// role in [message, status]
function sendMessage(text, role, recipient) {
    // Don't send empty messages
    if (!text) {
        return;
    }

    var msg = {
        text: text,
        role: role
    };
    if (recipient) {
        msg.recipient = {id: parseInt(recipient.id)};
    }
    socket.send(JSON.stringify(msg));

    // Immidiatly add to the board
    var now = new Date();
    var msgString;
    if (recipient) {
        msgString =
            formatMessage(now.getHours() + ':' + now.getMinutes() + ', ', 'date') +
            formatMessage(currentUser.username, 'sender') +
            formatMessage(' TO ', 'to') +
            formatMessage(recipient.username + ': ', 'recipient') +
            formatMessage(text, 'text');
    } else {
        msgString =
            formatMessage(now.getHours() + ':' + now.getMinutes() + ', ', 'date') +
            formatMessage(currentUser.username + ': ', 'sender') +
            formatMessage(text, 'text');
    }
    showMessage(msgString);
}

$btn.on('click', function (event) {
    event.preventDefault();
    var message = $input.val();
    var recipient = {
        id: $userlist.val(),
        username: $userlist.find('option:selected').text().trim()
    };
    sendMessage(message, 'message', recipient);
    $input.val('');
});

$input.on('keypress', function (event) {
    if (event.keyCode == 13) {
        event.preventDefault();
        var message = $input.val();
        var recipient = {
            id: $userlist.val(),
            username: $userlist.find('option:selected').text().trim()
        };
        sendMessage(message, 'message', recipient);
        $input.val('');
    }
});