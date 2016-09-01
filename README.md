Description
===========
Simple chat in Go.

Install
=======
1. Get project and install Go

    ```
    sudo apt-get install golang git

    export PATH="$PATH:/usr/local/go/bin"
    export GOPATH="$HOME/go"

    mkdir -p ~/go/bin ~/go/pkg ~/go/src/

    cd go/src/
    git clone https://github.com/tetafro/gochat.git
    cd gochat
    go get
    ```

2. PostgreSQL

    ```
    sudo apt-get install postgresql
    sudo su - postgres
    createdb db_gochat
    createuser --no-createdb pguser
    psql < migrations/0001_init.sql
    psql
    \password pguser
    GRANT ALL PRIVILEGES ON DATABASE db_gochat TO pguser;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pguser;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO pguser;
    ```

Compile and run
===============
```
go install
cd ~/go/bin
./gochat
```
