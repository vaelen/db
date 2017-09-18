#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
pushd $DIR


echo "========================================"
echo "Reformatting Source"
echo "========================================"

gofmt -w **/*.go
goimports -w **/*.go

echo "Building"
echo "========================================"

go build -o vdb cmd/vdb/vdb.go &&
    go build -o vdb-server cmd/vdb-server/vdb-server.go &&
    go install github.com/vaelen/db/cmd/vdb github.com/vaelen/db/cmd/vdb-server

if [ $? -eq 0 ]; then
    echo "Build Succeeded"
    echo "========================================"
else
    echo "========================================"
    echo "Build Failed"
    echo "========================================"
    popd
    exit 1
fi

echo "Running Linter"
echo "========================================"

golint github.com/vaelen/db/server \
       github.com/vaelen/db/client \
       github.com/vaelen/db/storage \
       github.com/vaelen/db/cmd/vdb \
       github.com/vaelen/db/cmd/vdb-server

popd
