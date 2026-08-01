#!/bin/sh

set -eu

REPO="hamidrezaesh/ffd"
INSTALL_DIR="/usr/local/bin"

echo "Installing ffd..."

# Detect operating system

OS="$(uname -s)"

case "$OS" in
Linux)
GOOS="linux"
;;
Darwin)
GOOS="darwin"
;;
*)
echo "Error: unsupported operating system: $OS"
exit 1
;;
esac

# Detect architecture

ARCH="$(uname -m)"

case "$ARCH" in
x86_64|amd64)
GOARCH="amd64"
;;
aarch64|arm64)
GOARCH="arm64"
;;
*)
echo "Error: unsupported architecture: $ARCH"
exit 1
;;
esac

# Get the latest release tag

LATEST_TAG="$(
curl -fsSL "https://api.github.com/repos/hamidrezaesh/ffd/releases/latest" |
grep '"tag_name":' |
sed -E 's/.*"([^"]+)".*/\1/'
)"

if [ -z "$LATEST_TAG" ]; then
echo "Error: could not determine latest release."
exit 1
fi

VERSION="${LATEST_TAG#v}"

# Download URL

ARCHIVE="ffd_${LATEST_TAG#v}_${GOOS}_${GOARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$ARCHIVE"

TMP_DIR="$(mktemp -d)"

cleanup() {
rm -rf "$TMP_DIR"
}

trap cleanup EXIT

echo "Downloading ffd $VERSION..."
curl -fL "$URL" -o "$TMP_DIR/ffd.tar.gz"

echo "Extracting..."
tar -xzf "$TMP_DIR/ffd.tar.gz" -C "$TMP_DIR"

echo "Installing to $INSTALL_DIR..."

if [ ! -w "$INSTALL_DIR" ]; then
sudo install -m 755 "$TMP_DIR/ffd" "$INSTALL_DIR/ffd"
else
install -m 755 "$TMP_DIR/ffd" "$INSTALL_DIR/ffd"
fi


echo
echo "ffd v$VERSION installed successfully!"
echo
echo "Run:"
echo "  ffd --help"
