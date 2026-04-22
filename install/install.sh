#!/bin/sh

set -eu

REPO="${EASYDOCKER_REPO:-joao-zanutto/easydocker}"
BINARY="easydocker"
INSTALL_DIR="${EASYDOCKER_INSTALL_DIR:-/usr/local/bin}"
VERSION="${EASYDOCKER_VERSION:-latest}"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Error: required command not found: $1" >&2
    exit 1
  fi
}

download() {
  url="$1"
  output="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$output"
    return
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "$output" "$url"
    return
  fi

  echo "Error: either curl or wget is required" >&2
  exit 1
}

detect_os() {
  case "$(uname -s)" in
    Linux)
      echo "linux"
      ;;
    Darwin)
      echo "darwin"
      ;;
    *)
      echo "Error: unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)
      echo "amd64"
      ;;
    arm64|aarch64)
      echo "arm64"
      ;;
    *)
      echo "Error: unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

resolve_tag() {
  if [ "$VERSION" = "latest" ]; then
    api_url="https://api.github.com/repos/$REPO/releases/latest"
    if command -v curl >/dev/null 2>&1; then
      tag="$(curl -fsSL "$api_url" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
    else
      tag="$(wget -qO - "$api_url" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
    fi

    if [ -z "$tag" ]; then
      echo "Error: could not resolve latest release tag from $api_url" >&2
      exit 1
    fi

    echo "$tag"
    return
  fi

  case "$VERSION" in
    v*) echo "$VERSION" ;;
    *) echo "v$VERSION" ;;
  esac
}

install_binary() {
  src="$1"
  dst="$INSTALL_DIR/$BINARY"

  if [ -w "$INSTALL_DIR" ]; then
    install -m 0755 "$src" "$dst"
    return
  fi

  if command -v sudo >/dev/null 2>&1; then
    sudo install -m 0755 "$src" "$dst"
    return
  fi

  echo "Error: cannot write to $INSTALL_DIR and sudo is not available" >&2
  echo "Hint: set EASYDOCKER_INSTALL_DIR to a writable path, then add it to PATH" >&2
  exit 1
}

verify_checksum() {
  checksum_file="$1"
  archive_file="$2"
  archive_name="$3"

  expected="$(grep "  $archive_name$" "$checksum_file" | awk '{print $1}' | head -n1 || true)"
  if [ -z "$expected" ]; then
    echo "Error: checksum not found for $archive_name" >&2
    exit 1
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$archive_file" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$archive_file" | awk '{print $1}')"
  else
    echo "Error: sha256sum or shasum is required for checksum verification" >&2
    exit 1
  fi

  if [ "$expected" != "$actual" ]; then
    echo "Error: checksum mismatch for $archive_name" >&2
    exit 1
  fi
}

need_cmd uname
need_cmd tar
need_cmd grep
need_cmd awk
need_cmd install

OS="$(detect_os)"
ARCH="$(detect_arch)"
TAG="$(resolve_tag)"
VERSION_NO_V="${TAG#v}"

ARCHIVE_NAME="${BINARY}_v${VERSION_NO_V}_${OS}_${ARCH}.tar.gz"
RELEASE_URL="https://github.com/$REPO/releases/download/$TAG"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

ARCHIVE_PATH="$TMP_DIR/$ARCHIVE_NAME"
CHECKSUMS_PATH="$TMP_DIR/checksums.txt"

echo "Installing $BINARY $TAG for ${OS}/${ARCH}..."
download "$RELEASE_URL/$ARCHIVE_NAME" "$ARCHIVE_PATH"
download "$RELEASE_URL/checksums.txt" "$CHECKSUMS_PATH"
verify_checksum "$CHECKSUMS_PATH" "$ARCHIVE_PATH" "$ARCHIVE_NAME"

tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

EXTRACTED_BIN="$(find "$TMP_DIR" -type f -name "$BINARY" -perm -u+x | head -n1 || true)"
if [ -z "$EXTRACTED_BIN" ]; then
  EXTRACTED_BIN="$(find "$TMP_DIR" -type f -name "$BINARY" | head -n1 || true)"
fi

if [ -z "$EXTRACTED_BIN" ]; then
  echo "Error: extracted binary '$BINARY' not found" >&2
  exit 1
fi

install_binary "$EXTRACTED_BIN"

echo "Installed: $INSTALL_DIR/$BINARY"
echo "Run: $BINARY"