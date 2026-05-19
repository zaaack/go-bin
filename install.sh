#!/usr/bin/env bash
set -euo pipefail

REPO="zaaack/go-bin"
BINARY="go-bin"
INSTALL_DIR="${INSTALL_DIR:-.}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)            echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

PLATFORM="${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/go-bin-${PLATFORM}.tar.gz"

echo "Downloading ${BINARY} for ${PLATFORM}..."
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "$TMPDIR/go-bin.tar.gz"
tar xzf "$TMPDIR/go-bin.tar.gz" -C "$TMPDIR"

mkdir -p "$INSTALL_DIR"
mv "$TMPDIR/go-bin-${PLATFORM}" "$INSTALL_DIR/${BINARY}"
chmod +x "$INSTALL_DIR/${BINARY}"

echo "Installed to $INSTALL_DIR/${BINARY}"
echo "Run: ./${BINARY} serve"
