#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o fgm cli/fauxgomo.go
