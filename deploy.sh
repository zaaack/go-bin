#!/usr/bin/env bash
set -euo pipefail

export REMOTE_HOST=$MINIPC_HOST
export REMOTE_DIR="\\\\DESKTOP-4PGV4PO\\Users\\z\\迷你主机\\webdav\\minipc-bin"
export SERVICE_NAME="go-bin"

echo ">>> Building go-bin.exe..."
GOOS=windows GOARCH=amd64 go build -o ./go-bin.exe ./cmd/go-bin

echo ">>> Stopping $SERVICE_NAME on $REMOTE_HOST..."
winrs.exe -r:$REMOTE_HOST "supervisorctl stop $SERVICE_NAME"

echo ">>> Copying go-bin.exe to $REMOTE_HOST..."
robocopy.exe . "$REMOTE_DIR" go-bin.exe /IS /IT || true

echo ">>> Starting $SERVICE_NAME on $REMOTE_HOST..."
winrs.exe -r:$REMOTE_HOST "supervisorctl start $SERVICE_NAME"

echo ">>> Deploy done."
