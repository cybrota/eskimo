#!/usr/bin/env sh
# This script installs eskimo from the GitHub releases.
# It detects the operating system and architecture, downloads
# the appropriate zip file, extracts the binary, and moves
# it to a folder in your PATH: /usr/local/bin
# Usage:
#   curl -sfL https://raw.githubusercontent.com/cybrota/eskimo/refs/heads/main/install.sh | sh

set -e

info() { printf '\033[1;34m==> %s\033[0m\n' "$1"; }
err() { printf '\033[1;31merror: %s\033[0m\n' "$1" >&2; }
success() { printf '\033[1;32m%s\033[0m\n' "$1"; }

# Determine OS
OS=$(uname | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)
    PLATFORM="Linux"
    ;;
  darwin)
    PLATFORM="Darwin"
    ;;
  *)
    err "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Determine architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)
    ARCH="x86_64"
    ;;
  i386|i686)
    ARCH="i386"
    ;;
  armv6l|armv7l)
    ARCH="arm"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    err "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Determine release version
if [ -z "$ESKIMO_VERSION" ] || [ "$ESKIMO_VERSION" = "latest" ]; then
  ESKIMO_VERSION=$(curl -s https://api.github.com/repos/cybrota/eskimo/releases/latest |
    grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$ESKIMO_VERSION" ]; then
    err "Could not determine latest version."
    exit 1
  fi
fi

FILE="eskimo_${PLATFORM}_${ARCH}.zip"
URL="https://github.com/cybrota/eskimo/releases/download/${ESKIMO_VERSION}/${FILE}"

info "Downloading eskimo ${ESKIMO_VERSION} for ${PLATFORM}/${ARCH}"
curl -L -o /tmp/${FILE} "${URL}"

EXTRACT_DIR=$(mktemp -d)
info "Extracting ${FILE}"
unzip -q -o /tmp/${FILE} -d "${EXTRACT_DIR}"
chmod +x "${EXTRACT_DIR}/eskimo"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "${INSTALL_DIR}" ]; then
  info "No write permission for ${INSTALL_DIR}. Attempting to use sudo"
  SUDO="sudo"
else
  SUDO=""
fi

info "Installing eskimo to ${INSTALL_DIR}"
$SUDO mv "${EXTRACT_DIR}/eskimo" "${INSTALL_DIR}/eskimo"

rm -rf "${EXTRACT_DIR}" /tmp/${FILE}

success "eskimo installed successfully. Make sure ${INSTALL_DIR} is in your PATH."

exit 0
