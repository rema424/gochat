// NOTE: currentUser variable is in template

// Interface elements
var $input = $('#input-message');
var $btn = $('#btn-send');
var $board = $('#board');
var $userlist = $('#input-users');
var $recipient = $('#recipient');

//
// WebSockets
//

// WebSocket init
var socket = new WebSocket("ws://gochat.local/ws");

// WebSocket events
socket.onopen = function () {
    showMessage(formatMessage('Connected', 'info'));

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

    // Get last messages
    $.getJSON({
        url: '/ajax/messages/last',
        dataType: 'json',
        success: function(response) {
            response.forEach(function (msg) {
                processMessage(msg);
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
    processMessage(msg);
};

socket.onerror = function (error) {
    showMessage(formatMessage('Error: ' + error.message, 'error'));
};

//
// Processing messages
//

// Display message
function showMessage(msg) {
    $board.html($board.html() + '<p>' + msg + '</p>');
    $board.scrollTop($board.prop('scrollHeight'));
}

// Add classes and tags to string
function formatMessage(text, cls) {
    var clsList = cls.split(',');

    var i;
    for (i = 0; i < clsList.length; i++) {
        clsList[i] = 'msg-'+clsList[i].trim();
    }
    clsList = clsList.join(' ')

    return '<span class="' + clsList + '">' + text + '</span>';
}

// Parse, format and display
function processMessage(msg) {
    if (currentUser && msg.role == 'new_user') {
        $userlist.append(
            $('<option></option>')
                .attr('value', msg.sender.id)
                .text(msg.sender.username)
        );
        msgString = formatMessage(msg.text, 'info');
        showMessage(msgString);
    } else if (msg.role == 'gone_user') {
        $userlist
            .find('option[value="' + msg.sender.id + '"]')
            .remove();
        msgString = formatMessage(msg.text, 'info');
        showMessage(msgString);
    } else if (msg.role == 'message') {
        // Timestamp
        var date = new Date(msg.send_date * 1000);
        date = date.getHours() + ':' + date.getMinutes();

        // Highlight self username
        var isSender = msg.sender.username == currentUser.username ? ', self' : '';

        // Format and show message
        if (msg.recipient) {
            var isRecipient = msg.recipient.username == currentUser.username ? ', self' : '';
            msgString =
                formatMessage(date + ', ', 'date') +
                formatMessage(msg.sender.username, 'sender' + isSender) +
                formatMessage(' TO ', 'delim') +
                formatMessage(msg.recipient.username, 'recipient' + isRecipient) +
                formatMessage(': ', 'delim') +
                formatMessage(msg.text, 'text');
        } else {
            msgString =
                formatMessage(date + ', ', 'date') +
                formatMessage(msg.sender.username, 'sender' + isSender) +
                formatMessage(': ', 'delim') +
                formatMessage(msg.text, 'text');
        }
        showMessage(msgString);
    }
}

//
// Sending messages
//

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
        msg.recipient = {
            id: parseInt(recipient.id),
            username: recipient.username
        };
    }
    socket.send(JSON.stringify(msg));

    // Immidiatly add to the board
    msg.send_date = new Date();
    msg.sender = currentUser;
    processMessage(msg);
}

// Get data from form and call sendMessage() for it
function submitMessageForm() {
    var message = $input.val();
    var recipient;
    var id = $userlist.val();

    if (id) {
        recipient = {
            id: $userlist.val(),
            username: $userlist.find('option:selected').text().trim()
        }
    };

    sendMessage(message, 'message', recipient);
    $input.val('');
}

//
// Interface events
//

$btn.on('click', function (event) {
    event.preventDefault();
    submitMessageForm();
});

$input.on('keypress', function (event) {
    if (event.keyCode == 13) {
        event.preventDefault();
        submitMessageForm();
    }
});

// Pick user as recipient for private message
$board.on('click', '.msg-sender, .msg-recipient', function () {
    var username = $(this).text();
    $userlist
        .find('option:contains("' + username + '")')
        .prop('selected', true);

    $recipient.text(username);
    $input.focus();
});

$userlist.on('change', function () {
    var username = $(this).find('option:selected').text();
    $recipient.text(username);
    $input.focus();
});

// Clear recipient
$recipient.on('click', function () {
    $recipient.text('@');
    $userlist
        .find('option:selected')
        .prop('selected', false);
    $input.focus();
});
