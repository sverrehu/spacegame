#!/bin/bash

BUILDS_DIR=builds
APP_NAME="$(basename "$(pwd)")"

test -d "${BUILDS_DIR}" || mkdir "${BUILDS_DIR}"

echo "Running tests"
env CGO_ENABLED=0 go test -a ./...

echo "Building"
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -ldflags="-s -w" -o "${BUILDS_DIR}/${APP_NAME}"-linux-amd64
env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -trimpath -ldflags="-s -w" -o "${BUILDS_DIR}/${APP_NAME}"-macos-amd64
env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -trimpath -ldflags="-s -w" -o "${BUILDS_DIR}/${APP_NAME}"-macos-arm64
env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -trimpath -ldflags="-s -w" -o "${BUILDS_DIR}/${APP_NAME}"-windows-amd64.exe
