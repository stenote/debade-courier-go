#!/bin/bash

mkdir /go

export GOROOT=/usr/lib/go
export GOPATH=/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

cd $GOPATH

go get github.com/gobs/simplejson

go get github.com/pebbe/zmq4

go get github.com/streadway/amqp

go get gopkg.in/yaml.v1

git clone https://github.com/gosexy/to.git /go/src/menteslibres.net/gosexy/to
git clone https://github.com/gosexy/dig.git /go/src/menteslibres.net/gosexy/dig
git clone https://github.com/gosexy/yaml.git /go/src/menteslibres.net/gosexy/yaml

git clone https://github.com/stenote/debade-courier-go.git /go/src/github.com/stenote/debade-courier-go

go run /go/src/github.com/stenote/debade-courier-go/debade-courier.go -v -c /etc/debade/courier.yml
