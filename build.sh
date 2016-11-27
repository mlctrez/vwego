#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o /tmp/vwego vwego/vwego.go

scp /tmp/vwego config.json pe0:/tmp
