#!/bin/sh

# cx-cli Installation Script
# Automated installer for macOS and Linux.

set -e

# 1. Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    darwin|linux)
        ;;
    *)
        echo "Error: Unsupported operating system: $OS"
        exit 1
        ;;
esac

# 2. Determine target binary name
BINARY_NAME="cx-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/guppshub/cx-cli/releases/latest/download/${BINARY_NAME}"

# Special handling if Windows executable extension is needed (not typical for curl installers, but safe)
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

echo "========================================="
echo " Installing cx-cli"
echo " Detected OS  : $OS"
echo " Detected Arch: $ARCH"
echo "========================================="

# 3. Create a temporary download directory
TEMP_DIR="$(mktemp -d)"
TEMP_BIN="${TEMP_DIR}/cx"

# Cleanup on exit
trap 'rm -rf "$TEMP_DIR"' EXIT

# 4. Fetch the latest release binary
echo "Downloading $DOWNLOAD_URL..."
if curl -sL --fail "$DOWNLOAD_URL" -o "$TEMP_BIN"; then
    echo "Download completed successfully."
else
    echo "Error: Failed to download binary. Please make sure a release tag has been published on GitHub."
    exit 1
fi

# 5. Determine installation folder
INSTALL_DIR="/usr/local/bin"
TARGET_PATH="${INSTALL_DIR}/cx"

echo "Installing to ${TARGET_PATH}..."

# Check if target directory is writable
if [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_BIN" "$TARGET_PATH"
    chmod +x "$TARGET_PATH"
else
    echo "Requires administrator privileges to install in ${INSTALL_DIR}."
    sudo mv "$TEMP_BIN" "$TARGET_PATH"
    sudo chmod +x "$TARGET_PATH"
fi

echo "========================================="
echo " Success! cx-cli is now installed."
echo " Run 'cx init' to get started."
echo "========================================="
