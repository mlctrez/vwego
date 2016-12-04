#!/usr/bin/env bash

set -x

GOPATH=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
export GOPATH

go get github.com/nats-io/gnatsd/
go get github.com/nats-io/nats/
go get github.com/satori/go.uuid

mkdir -p $GOPATH/src/github.com/mlctrez

ln -s `pwd` $GOPATH/src/github.com/mlctrez/vwego

go build -o vw vwego/vwego.go

echo $GOPATH

rm -rf $GOPATH









