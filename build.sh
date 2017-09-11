#!/bin/bash

export GOPATH=~/go

go build -o vdb cmd/vdb/vdb.go &&
    go build -o vdb-server cmd/vdb-server/vdb-server.go &&
    go install github.com/vaelen/db/cmd/vdb github.com/vaelen/db/cmd/vdb-server
