#!/bin/bash

update(){
    go get go@latest || { echo -e "\033[31mFailed: $1.*\033[0m" ; return 1; }
    go get -u || { echo -e "\033[31mFailed: $1.*\033[0m" ; return 1; }
    go mod tidy || { echo -e "\033[31mFailed: $1.*\033[0m" ; return 1; }
    for indirect in $(tail +3 go.mod | grep "// indirect" | awk '{if ($1 =="require") print $2; else print $1;}'); do go get -u "${indirect}"; done
    go get -u || { echo -e "\033[31mFailed: $1.*\033[0m" ; return 1; }
    go mod tidy || { echo -e "\033[31mFailed: $1.*\033[0m" ; return 1; }
}

build(){
    ( env GOOS=linux GOARCH=amd64 go build -o "$1.bin" . && echo -e "\033[32mBuild: $1.bin\033[0m" ) || { echo -e "\033[31mFailed: $1.bin\033[0m" ; return 1; }
    # ( env GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ go build -o "$1.exe" . && echo -e "\033[32mBuild: $1.exe\033[0m" ) || { echo -e "\033[31mFailed: $1.exe\033[0m" ; return 1; }
}

file="ATowerDefense"

update "$file" && build "$file"
