#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
GOBIN_DIR=$(go env GOBIN)
if [ -z "$GOBIN_DIR" ]; then
	GOBIN_DIR=$(go env GOPATH)/bin
fi

go build -o "$ROOT/bvr-cli" "$ROOT"
mkdir -p "$GOBIN_DIR"
cp "$ROOT/bvr-cli" "$GOBIN_DIR/bvr-cli"

# Keep the common user-local installation synchronized too. This avoids an
# older copy winning when both directories are present on PATH.
LOCAL_BIN="$HOME/.local/bin"
mkdir -p "$LOCAL_BIN"
cp "$ROOT/bvr-cli" "$LOCAL_BIN/bvr-cli"

printf 'Installed latest bvr-cli to:\n  %s\n  %s\n' "$GOBIN_DIR/bvr-cli" "$LOCAL_BIN/bvr-cli"