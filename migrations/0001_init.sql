-- DB

CREATE DATABASE db_gochat;
GRANT ALL PRIVILEGES ON DATABASE db_gochat TO pguser;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pguser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO pguser;

-- Tables

CREATE TABLE auth_user (
    id serial NOT NULL,
    full_name varchar(60),
    username varchar(60) NOT NULL,
    email varchar(60) NOT NULL,
    password varchar(60) NOT NULL,
    role varchar(20) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE auth_session (
    id serial NOT NULL,
    key varchar(64) NOT NULL,
    user_id integer NOT NULL,
    create_date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expire_date timestamp NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE room (
    id serial NOT NULL,
    name varchar(60) NOT NULL,
    create_date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE room_role (
    id serial NOT NULL,
    room_id integer NOT NULL,
    user_id integer NOT NULL,
    role_id integer NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE role_name (
    id serial NOT NULL,
    name varchar(16) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE message (
    id serial NOT NULL,
    room_id integer NOT NULL,
    sender_id integer NOT NULL,
    recipient_id integer,
    text text NOT NULL,
    send_date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recieve_date timestamp,
    PRIMARY KEY (id)
);

CREATE TABLE ban (
    id serial NOT NULL,
    room_id integer NOT NULL,
    user_id integer NOT NULL,
    date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE mute (
    id serial NOT NULL,
    room_id integer NOT NULL,
    user_id integer NOT NULL,
    date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE db_version (
    id serial NOT NULL,
    version integer NOT NULL,
    PRIMARY KEY (id)
);

-- Populate

INSERT INTO auth_user (id, full_name, username, email, password, role)
    VALUES (1, 'Administrator', 'admin', 'admin@gochat.local', '123', 'admin');
INSERT INTO auth_user (id, full_name, username, email, password, role)
    VALUES (2, 'Moderator', 'moder', 'moder@gochat.local', '123', 'moder');
INSERT INTO auth_user (id, full_name, username, email, password, role)
    VALUES (3, 'User #1', 'user1', 'user1@gochat.local', '123', 'user');
INSERT INTO auth_user (id, full_name, username, email, password, role)
    VALUES (4, 'User #2', 'user2', 'user2@gochat.local', '123', 'user');

INSERT INTO room (id, name, create_date) VALUES (1, 'Room #1', CURRENT_TIMESTAMP);
INSERT INTO room (id, name, create_date) VALUES (2, 'Room #2', CURRENT_TIMESTAMP);

INSERT INTO role_name (id, name) VALUES (1, 'admin');
INSERT INTO role_name (id, name) VALUES (2, 'moder');
INSERT INTO role_name (id, name) VALUES (3, 'user');

INSERT INTO room_role (id, room_id, user_id, role_id) VALUES (1, 1, 1, 1);
INSERT INTO room_role (id, room_id, user_id, role_id) VALUES (2, 1, 2, 2);
INSERT INTO room_role (id, room_id, user_id, role_id) VALUES (3, 1, 3, 3);
INSERT INTO room_role (id, room_id, user_id, role_id) VALUES (4, 1, 4, 3);
INSERT INTO room_role (id, room_id, user_id, role_id) VALUES (5, 2, 1, 1);
INSERT INTO room_role (id, room_id, user_id, role_id) VALUES (6, 2, 2, 2);
INSERT INTO room_role (id, room_id, user_id, role_id) VALUES (7, 2, 3, 3);
INSERT INTO room_role (id, room_id, user_id, role_id) VALUES (8, 2, 4, 3);

INSERT INTO db_version (id, version) VALUES (1, 1);
