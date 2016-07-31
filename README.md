Description
===========
Simple chat in Go.

Install
=======
```
apt-get install golang git

export PATH=$PATH:/usr/local/go/bin
export GOPATH=~/go

mkdir -p ~/go/bin ~/go/pkg ~/go/src/

cd go/src/
git clone https://github.com/tetafro/gochat.git
cd gochat
go get
```

Compile and run
===============
```
go install
cd ~/go/bin
./gochat
```
