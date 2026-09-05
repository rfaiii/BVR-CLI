#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
VERSION=${1:-dev}
BUILD_DIR="$ROOT/dist/darwin"
MACOS_ZIP="$BUILD_DIR/bvr-cli_${VERSION}_darwin_arm64.zip"

mkdir -p "$BUILD_DIR"
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=greenteagc go build -trimpath -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=${VERSION}" -o "$BUILD_DIR/bvr-cli" "$ROOT"

cp "$ROOT/README.md" "$BUILD_DIR/README.txt"
cp "$ROOT/LICENSE" "$BUILD_DIR/LICENSE.txt" 2>/dev/null || cp "$ROOT/LICENSE.md" "$BUILD_DIR/LICENSE.txt" 2>/dev/null || true
cp "$ROOT/resources/icons/bvr-logo-1024.png" "$BUILD_DIR/bvr-logo.png" 2>/dev/null || true

zip -j "$MACOS_ZIP" "$BUILD_DIR/bvr-cli" "$BUILD_DIR/README.txt" "$BUILD_DIR/LICENSE.txt" "$BUILD_DIR/bvr-logo.png" 2>/dev/null || true

printf 'Packaged macOS zip at: %s\n' "$MACOS_ZIP"
