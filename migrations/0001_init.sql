CREATE DATABASE db_gochat;
GRANT ALL PRIVILEGES ON DATABASE db_gochat TO pguser;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pguser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO pguser;

CREATE TABLE auth_user
(
    id serial NOT NULL,
    full_name varchar(60),
    username varchar(60) NOT NULL,
    email varchar(60) NOT NULL,
    password varchar(60) NOT NULL,
    role varchar(20) NOT NULL,
    is_muted boolean DEFAULT FALSE,
    mute_date timestamp,
    is_banned boolean DEFAULT FALSE,
    ban_date timestamp,
    PRIMARY KEY (id)
);

INSERT INTO auth_user (id, full_name, username, email, password, role)
    VALUES (1, 'Administrator', 'admin', 'admin@gochat.local', '123', 'admin');
INSERT INTO auth_user (id, full_name, username, email, password, role)
    VALUES (2, 'Moderator', 'moder', 'moder@gochat.local', '123', 'moder');
INSERT INTO auth_user (id, full_name, username, email, password, role)
    VALUES (3, 'User #1', 'user1', 'user1@gochat.local', '123', 'user');
INSERT INTO auth_user (id, full_name, username, email, password, role)
    VALUES (4, 'User #2', 'user2', 'user2@gochat.local', '123', 'user');

CREATE TABLE auth_session
(
    id serial NOT NULL,
    key varchar(64) NOT NULL,
    user_id integer NOT NULL,
    create_date timestamp NOT NULL,
    expire_date timestamp NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE message
(
    id serial NOT NULL,
    sender_id integer NOT NULL,
    recipient_id integer,
    text text NOT NULL,
    send_date timestamp NOT NULL,
    recieve_date timestamp,
    PRIMARY KEY (id)
);

CREATE TABLE db_version
(
    id serial NOT NULL,
    version integer NOT NULL,
    PRIMARY KEY (id)
);

INSERT INTO db_version (id, version) VALUES (1, 1);
