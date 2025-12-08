#!/bin/sh
set -e

PROJECT_DIR="/data/manage_portal/source/walletus" 
OUTPUT_DIR="/data/manage_portal/server"
BINARY_NAME="walletus-api"

echo "switch direction：$PROJECT_DIR"
cd "$PROJECT_DIR"

echo "pull code for update..."
git fetch origin
git reset --hard origin/main
git clean -fd

echo "start compiling..."
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

go build -o "$OUTPUT_DIR/$BINARY_NAME" cmd/modapi/main.go
echo "compiled successfully and moved app to $OUTPUT_DIR/$BINARY_NAME"