#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"

VERSION="${1:-1.2.1}"
OUT="$ROOT/dist/release"
mkdir -p "$OUT"

GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc GOOS=darwin GOARCH=arm64 \
  go build -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=${VERSION}" \
  -o "$OUT/bvr-cli_${VERSION}_darwin_arm64" .

GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc GOOS=darwin GOARCH=amd64 \
  go build -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=${VERSION}" \
  -o "$OUT/bvr-cli_${VERSION}_darwin_amd64" .

GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=${VERSION}" \
  -o "$OUT/bvr-cli_${VERSION}_linux_amd64" .

GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc GOOS=linux GOARCH=arm64 \
  go build -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=${VERSION}" \
  -o "$OUT/bvr-cli_${VERSION}_linux_arm64" .

GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc GOOS=windows GOARCH=amd64 \
  go build -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=${VERSION}" \
  -o "$OUT/bvr-cli_${VERSION}_windows_amd64.exe" .

GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc GOOS=windows GOARCH=arm64 \
  go build -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=${VERSION}" \
  -o "$OUT/bvr-cli_${VERSION}_windows_arm64.exe" .

ls -lh "$OUT"
