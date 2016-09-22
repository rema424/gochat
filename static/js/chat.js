// NOTE: currentUser variable is in template

// Interface elements
var $board = $('#board');
var $inputMsg = $('input[name="message"]');
var $btnSend = $('button[name="send"]');
var $btnMute = $('button[name="mute"]');
var $btnKick = $('button[name="kick"]');
var $btnBan = $('button[name="ban"]');
var $btnExit = $('button[name="exit"]');
var $userlist = $('select[name="users"]');
var $recipient = $('.recipient');

// Permissions for managing users
var PERMISSIONS = {
    admin: ['mute', 'kick', 'ban'],
    moder: ['mute', 'kick'],
    user: []
}

// All users in the room
var allUsers = {}

//
// WebSockets
//

// WebSocket init
var socket = new WebSocket("ws://gochat.local/ws/rooms/" + currentRoom.id);

// WebSocket events
socket.onopen = function () {
    showMessage(formatMessage('Connected', 'info'));
    getAllUsers();
    getLastMessages();
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
    console.log(msg);
    processMessage(msg);
};

socket.onerror = function (error) {
    showMessage(formatMessage('Error: ' + error.message, 'error'));
};

//
// Processing users
//

// Get all users of the room
function getAllUsers() {
    $.getJSON({
        url: '/ajax//rooms/' + currentRoom.id + '/users',
        dataType: 'json',
        success: function(response) {
            response.forEach(function (user) {
                addUser(user);
            });
        },
        error: function(response) {
            console.log(response);
        }
    });
}

// Add new user to userlist
function addUser(user) {
    allUsers[user.id] = user;

    var $opt = $('<option></option>')
        .attr('value', user.id)
        .html(user.username);

    if (user.role == 'admin') {
        $opt.addClass('admin');
    } else if (user.role == 'admin') {
        $opt.addClass('moder');
    }

    if (user.mute) {
        $opt.addClass('muted');
    }

    $userlist.append($opt);
}

// Remove user from userlist
function removeUser(user) {
    delete allUsers[user.id];

    $userlist
        .find('option[value="' + user.id + '"]')
        .remove();
}

//
// Processing messages
//

function getLastMessages() {
    // Get last messages
    $.getJSON({
        url: '/ajax/rooms/' + currentRoom.id + '/messages',
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
}

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

// Get pretty formatted date
function formatDate(timestamp) {
    var monthName = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
        'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    var date = new Date(timestamp * 1000);
    var now = new Date();
    var diffDays = (now.getDate() - date.getDate());

    if (diffDays == 0) {
        dateStr = ("0" + date.getHours()).slice(-2) + ':' +
            ("0" + date.getMinutes()).slice(-2);
    } else if (diffDays == 1) {
        dateStr = 'Yesterday';
    } else {
        dateStr = monthName[date.getMonth()] + ', ' + date.getDate();
    }

    return dateStr;
}

// Parse, format and display
function processMessage(msg) {
    switch (msg.action) {
        case 'new_user':
            if (currentUser) {
                addUser(msg.sender);
                msgString = formatMessage(msg.text, 'info');
                showMessage(msgString);
            }
            break;

        case 'gone_user':
            removeUser(msg.sender);
            msgString = formatMessage(msg.text, 'info');
            showMessage(msgString);
            break;

        case 'mute':
            $userlist
                .find('option[value="' + msg.recipient.id + '"]')
                .toggleClass('muted');
            msgString = formatMessage(msg.text, 'info');
            showMessage(msgString);

            // Toggle disability of inputs
            if (currentUser.id == msg.recipient.id) {
                $inputMsg.prop(
                    'disabled',
                    !$inputMsg.prop('disabled')
                );
                $btnSend.prop(
                    'disabled',
                    !$btnSend.prop('disabled')
                );
            }

            break;

        case 'ban':
            removeUser(msg.recipient);
            msgString = formatMessage(msg.text, 'info');
            showMessage(msgString);
            break;

        case 'message':
            // Timestamp
            var date = formatDate(msg.send_date);

            // Highlight self
            var isSender = msg.sender.username == currentUser.username ? ', self' : '';

            // Format and show message
            if (msg.recipient) {
                var isRecipient = msg.recipient.username == currentUser.username ? ', self' : '';
                msgString =
                    formatMessage(date + ', ', 'date') +
                    formatMessage(msg.sender.username, msg.sender.role + ', sender' + isSender) +
                    formatMessage(' TO ', 'delim') +
                    formatMessage(msg.recipient.username, msg.recipient.role + ', recipient' + isRecipient) +
                    formatMessage(': ', 'delim') +
                    formatMessage(msg.text, 'text');
            } else {
                msgString =
                    formatMessage(date + ', ', 'date') +
                    formatMessage(msg.sender.username, msg.sender.role + ', sender' + isSender) +
                    formatMessage(': ', 'delim') +
                    formatMessage(msg.text, 'text');
            }
            showMessage(msgString);
            break;
    }
}

//
// Sending messages
//

// Send message
// role in [message, status]
function sendMessage(text, action, recipient) {
    // Don't send empty messages
    if (!text) {
        return;
    }

    var msg = {
        text: text,
        action: action
    };
    if (recipient) {
        msg.recipient = recipient;
    }
    socket.send(JSON.stringify(msg));

    // Immidiatly add to the board
    msg.send_date = (new Date())/1000;
    msg.sender = currentUser;
    processMessage(msg);
}

// Get data from form and call sendMessage() for it
function submitMessageForm() {
    var message = $inputMsg.val();
    var recipient;
    var recipientId = $userlist.val();

    if (recipientId) {
        recipient = allUsers[recipientId];
    };

    sendMessage(message, 'message', recipient);
    $inputMsg.val('');
}

// Check if current user can do action
function checkPrivilege(action) {
    if (PERMISSIONS[currentUser.role].indexOf(action) == -1) {
        return false;
    } else {
        return true;
    }
}

// Mute, kick and ban
function manageUser(userId, action) {
    if (!checkPrivilege(action)) {
        return;
    }

    var msg = {
        recipient: {id: userId},
        action: action
    };
    socket.send(JSON.stringify(msg));
}

//
// Interface events
//

// Make active mute, kick and ban buttons for admin
function enableMgntBtns() {
    var active = !Boolean($userlist.val());
    if (['admin', 'moder'].indexOf(currentUser.role) > -1) {
        $btnMute.prop('disabled', active);
        $btnKick.prop('disabled', active);
        $btnBan.prop('disabled', active);
    }
}

$btnSend.on('click', function (event) {
    event.preventDefault();
    submitMessageForm();
});

$inputMsg.on('keypress', function (event) {
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
    $inputMsg.focus();

    enableBanBtn();
});

$userlist.on('change', function () {
    var username = $(this).find('option:selected').text();
    $recipient.text(username);
    $inputMsg.focus();

    enableMgntBtns();
});

// Clear recipient
$recipient.on('click', function () {
    $recipient.text('@');
    $userlist
        .find('option:selected')
        .prop('selected', false);
    $inputMsg.focus();

    enableBanBtn();
});

// Mute, kick and ban
$btnMute.on('click', function (event) {
    event.preventDefault();
    var userId = parseInt($userlist.find('option:selected').val());
    manageUser(userId, 'mute');
});

$btnKick.on('click', function (event) {
    event.preventDefault();
    var userId = parseInt($userlist.find('option:selected').val());
    manageUser(userId, 'kick');
});

$btnBan.on('click', function (event) {
    event.preventDefault();
    var userId = parseInt($userlist.find('option:selected').val());
    manageUser(userId, 'ban');
});

$btnExit.on('click', function (event) {
    event.preventDefault();
    window.location.href = '/';
});