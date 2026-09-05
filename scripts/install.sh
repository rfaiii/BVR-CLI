#!/usr/bin/env bash
set -e

# BVR-CLI Zero-Friction Installation Script
# https://get.bvr-cli.dev/install.sh

GITHUB_REPO="richavery/BVR-CLI" # TODO: Update when repo is renamed/public
BINARY_NAME="bvr-cli"
INSTALL_DIR="/usr/local/bin"

echo "🚀 Installing BVR-CLI..."

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_NAME="Linux";;
    Darwin*)    OS_NAME="Darwin";;
    CYGWIN*|MINGW*|MSYS*) OS_NAME="Windows";;
    *)          echo "Unsupported OS: ${OS}"; exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64|amd64) ARCH_NAME="x86_64";;
    arm64|aarch64) ARCH_NAME="arm64";;
    i386|i686)    ARCH_NAME="i386";;
    *)            echo "Unsupported architecture: ${ARCH}"; exit 1;;
esac

# For Windows, use zip, else tar.gz
EXT="tar.gz"
if [ "${OS_NAME}" = "Windows" ]; then
    EXT="zip"
    INSTALL_DIR="/c/Windows/System32" # Assuming git bash or similar, adjust as needed
fi

echo "🔍 Detected ${OS_NAME} ${ARCH_NAME}..."

# Fetch latest release data from GitHub API
LATEST_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
echo "⬇️  Fetching latest version info..."

LATEST_RELEASE=$(curl -sL "${LATEST_URL}")
VERSION=$(echo "${LATEST_RELEASE}" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${VERSION}" ]; then
    echo "❌ Failed to fetch the latest version. Are you rate-limited by GitHub?"
    exit 1
fi

echo "📦 Found version ${VERSION}"

# Match Goreleaser naming format: bvr-cli_1.2.1_Darwin_x86_64.tar.gz
# goreleaser drops the 'v' from version in the archive name
VERSION_CLEAN=${VERSION#v}
ARCHIVE_NAME="${BINARY_NAME}_${VERSION_CLEAN}_${OS_NAME}_${ARCH_NAME}.${EXT}"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

TMP_DIR=$(mktemp -d)
cd "${TMP_DIR}"

echo "⬇️  Downloading ${ARCHIVE_NAME}..."
if ! curl -sL -f -o "${ARCHIVE_NAME}" "${DOWNLOAD_URL}"; then
    echo "❌ Download failed! URL: ${DOWNLOAD_URL}"
    exit 1
fi

echo "📦 Extracting..."
if [ "${EXT}" = "zip" ]; then
    unzip -q "${ARCHIVE_NAME}"
else
    tar -xzf "${ARCHIVE_NAME}"
fi

echo "🛠️  Installing to ${INSTALL_DIR} (requires sudo)..."
if [ "${OS_NAME}" = "Windows" ]; then
    # Very basic windows fallback
    cp "bvr-cli.exe" "${INSTALL_DIR}/bvr.exe" || echo "Please copy bvr-cli.exe to your PATH manually."
else
    sudo mv "${BINARY_NAME}" "${INSTALL_DIR}/bvr"
    sudo chmod +x "${INSTALL_DIR}/bvr"
fi

# Cleanup
rm -rf "${TMP_DIR}"

echo ""
echo "✅ BVR successfully installed to ${INSTALL_DIR}/bvr!"
echo ""
echo "Next Steps:"
echo "  Run 'bvr auth <LICENSE_KEY>' to authenticate and activate your device."
echo ""
