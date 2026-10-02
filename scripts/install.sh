#!/bin/sh

set -eu

REPO="hamidrezaesh/ffd"
INSTALL_DIR="/usr/local/bin"
BETA=false

# Parse arguments

while [ $# -gt 0 ]; do
case "$1" in
--beta)
BETA=true
;;
-h|--help)
echo "Usage: install.sh [OPTIONS]"
echo
echo "Options:"
echo "  --beta    Install the latest beta release"
echo "  -h, --help    Show this help message"
exit 0
;;
*)
echo "Error: unknown option: $1"
exit 1
;;
esac
shift
done

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

# Get release tag

if [ "$BETA" = true ]; then
    echo "Fetching latest beta release..."

    RELEASES="$(
        curl -fsSL "https://api.github.com/repos/$REPO/releases"
    )"

    LATEST_TAG="$(
        printf '%s\n' "$RELEASES" |
        grep '"tag_name":' |
        head -n 1 |
        sed -E 's/.*"([^"]+)".*/\1/'
    )"
else
    echo "Fetching latest stable release..."

    LATEST_TAG="$(
        curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
        grep '"tag_name":' |
        sed -E 's/.*"([^"]+)".*/\1/'
    )"

    # Never install a beta in stable mode.
    if printf '%s\n' "$LATEST_TAG" | grep -q '^v.*-beta\.'; then
        echo "Error: refusing to install beta release: $LATEST_TAG"
        exit 1
    fi
fi

if [ -z "$LATEST_TAG" ]; then
    echo "Error: could not determine release."
    exit 1
fi

VERSION="${LATEST_TAG#v}"

# Download URL

ARCHIVE="ffd_${VERSION}_${GOOS}_${GOARCH}.tar.gz"
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
sudo install -Dm 644 "$TMP_DIR/docs/ffd.1" "/usr/local/share/man/man1/ffd.1"
else
install -m 755 "$TMP_DIR/ffd" "$INSTALL_DIR/ffd"
install -Dm 644 "$TMP_DIR/docs/ffd.1" "/usr/local/share/man/man1/ffd.1"
fi

echo
echo "ffd v$VERSION installed successfully!"
echo
echo "Run:"
echo "  ffd --help"