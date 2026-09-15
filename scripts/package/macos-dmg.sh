#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
VERSION=${1:-dev}
ARCH=${2:-arm64}
BUILD_DIR="$ROOT/dist/darwin"
STAGING="$BUILD_DIR/dmg-staging"
DMG="$BUILD_DIR/bvr-cli_${VERSION}_darwin_${ARCH}.dmg"

mkdir -p "$STAGING" "$BUILD_DIR"

echo "Building bvr-cli (${ARCH})..."
GOOS=darwin GOARCH=${ARCH} CGO_ENABLED=0 GOEXPERIMENT=greenteagc go build -trimpath \
  -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=${VERSION}" \
  -o "$STAGING/bvr-cli" "$ROOT"

echo "Copying files..."
cp "$ROOT/README.md" "$STAGING/README.txt" 2>/dev/null || true
cp "$ROOT/LICENSE" "$STAGING/LICENSE.txt" 2>/dev/null || cp "$ROOT/LICENSE.md" "$STAGING/LICENSE.txt" 2>/dev/null || true

echo "Creating DMG..."
rm -f "$DMG"
hdiutil create -volname "BVR-CLI $VERSION" -srcfolder "$STAGING" -ov -format UDZO "$DMG"

echo "Created: $DMG"
ls -lh "$DMG"
