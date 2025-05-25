#!/bin/bash

build_api() {
    go build -o ./bin/api.o ./cmd/api/

    echo "$PWD/bin/api.o"
}

case $1 in
"api")
    build_api;;
esac
