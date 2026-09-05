#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
VERSION=${1:-dev}
BUILD_DIR="$ROOT/dist/darwin"
STAGING="$BUILD_DIR/dmg-staging"
DMG="$BUILD_DIR/bvr-cli_${VERSION}_darwin_arm64.dmg"

mkdir -p "$STAGING" "$BUILD_DIR"

echo "Building bvr-cli..."
GOOS=darwin GOARCH=arm64 go build -o "$STAGING/bvr-cli" "$ROOT"

echo "Copying files..."
cp "$ROOT/README.md" "$STAGING/README.txt" 2>/dev/null || true
cp "$ROOT/LICENSE" "$STAGING/LICENSE.txt" 2>/dev/null || cp "$ROOT/LICENSE.md" "$STAGING/LICENSE.txt" 2>/dev/null || true

echo "Creating DMG..."
rm -f "$DMG"
hdiutil create -volname "BVR-CLI $VERSION" -srcfolder "$STAGING" -ov -format UDZO "$DMG"

echo "Created: $DMG"
ls -lh "$DMG"
