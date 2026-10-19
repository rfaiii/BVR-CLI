#!/bin/bash
# Build GGWave C library for use with BVR's CGO adapter.
# Run this script before building BVR with -tags=ggwave

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
GGWAVE_DIR="$PROJECT_ROOT/internal/ggwave"
BUILD_DIR="$GGWAVE_DIR/build"

echo "Building GGWave in $GGWAVE_DIR..."

# Check if already built
if [[ -f "$BUILD_DIR/libggwave.a" ]] || [[ -f "$BUILD_DIR/libggwave.dylib" ]] || [[ -f "$BUILD_DIR/libggwave.so" ]]; then
    echo "GGWave already built at $BUILD_DIR"
    exit 0
fi

# Create ggwave directory if it doesn't exist
mkdir -p "$GGWAVE_DIR"

# Check if ggwave source exists
if [[ ! -d "$GGWAVE_DIR/ggwave" ]]; then
    echo "Cloning GGWave repository..."
    git clone --depth 1 --recursive https://github.com/ggerganov/ggwave.git "$GGWAVE_DIR/ggwave"
fi

cd "$GGWAVE_DIR/ggwave"

# Build
mkdir -p build
cd build
cmake .. -DCMAKE_BUILD_TYPE=Release -DGGWAVE_BUILD_EXAMPLES=OFF
make -j$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

# Copy headers to include directory for CGO
mkdir -p "$GGWAVE_DIR/include/ggwave"
cp -r ../include/ggwave/* "$GGWAVE_DIR/include/ggwave/"

echo "GGWave built successfully!"
echo "Library: $BUILD_DIR/libggwave.a"
echo "Headers: $GGWAVE_DIR/include/ggwave/"
echo ""
echo "Now build BVR with: go build -tags=ggwave ."