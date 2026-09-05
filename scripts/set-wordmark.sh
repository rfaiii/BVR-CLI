#!/usr/bin/env bash
set -euo pipefail

# Change the terminal UI name in one place.
#
# Usage:
#   ./scripts/set-wordmark.sh BVR-CLI
#
# Instructions:
#   1. Use uppercase letters, numbers, and hyphens only.
#   2. Pass the complete display name, including a suffix such as -CLI.
#   3. Rebuild the binary after changing it: go build -o bvr-cli .
#   4. Run the logo tests: go test ./internal/ui/logo ./internal/ui/model
#
# The script updates internal/ui/logo/logo.go's Wordmark constant. The UI uses
# that constant for the wide logo, compact header, and sidebar wordmark.

if [[ $# -ne 1 ]]; then
	printf 'Usage: %s WORDMARK\n' "$0" >&2
	exit 2
fi

wordmark=$1
if [[ ! $wordmark =~ ^[A-Z0-9]+(-[A-Z0-9]+)*$ ]]; then
	printf 'Wordmark must contain uppercase letters, numbers, and hyphens only.\n' >&2
	exit 2
fi

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
logo_file="$script_dir/../internal/ui/logo/logo.go"

WORDMARK="$wordmark" LOGO_FILE="$logo_file" python3 - <<'PY'
from pathlib import Path
import os
import re

path = Path(os.environ["LOGO_FILE"])
source = path.read_text()
updated, count = re.subn(
    r'const Wordmark = "[A-Z0-9]+(?:-[A-Z0-9]+)*"',
    f'const Wordmark = "{os.environ["WORDMARK"]}"',
    source,
    count=1,
)
if count != 1:
    raise SystemExit("Could not find the Wordmark constant")
path.write_text(updated)
PY

printf 'Updated BVR-CLI wordmark to %s\n' "$wordmark"