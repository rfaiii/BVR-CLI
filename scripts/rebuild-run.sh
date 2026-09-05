#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"

GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc go build -o "$ROOT/bvr-cli" .
exec "$ROOT/bvr-cli" "$@"
