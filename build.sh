#!/bin/bash

#set -x
set -e

version=$(git describe --dirty)

build() {
    echo "$4"
    env GOOS=$1 GOARCH=$2 go build -ldflags="-X main.version=$version" -o "build/$3"
    cd build
    if [ "$1" = "windows" ]; then
        zip "$4.zip" "$3" --quiet
    else
        tar -zcf "$4.tar.gz" "$3"
    fi
    rm "$3"
    cd ..
}

mkdir -p build
rm -rf build/*

build "windows" "amd64" "petrify.exe" "petrify-$version.windows-64bit"
build "windows" "386" "petrify.exe" "petrify-$version.windows-32bit"
build "darwin" "amd64" "petrify" "petrify-$version.macOS-64bit"
build "darwin" "386" "petrify" "petrify-$version.macOS-32bit"
build "linux" "amd64" "petrify" "petrify-$version.linux-64bit"
build "linux" "386" "petrify" "petrify-$version.linux-32bit"
build "linux" "arm" "petrify" "petrify-$version.linux-arm"
