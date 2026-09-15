#!/bin/sh
set -eu

# Build locally testable beta artifacts for every currently supported
# OS/architecture pair. GoReleaser remains the release/signing path.
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
VERSION=1.2.4
[ "$#" -ge 1 ] && VERSION=$1
OUT="$ROOT/dist/packages"
STAGING="$OUT/.staging"

rm -rf "$STAGING"
mkdir -p "$OUT" "$STAGING"

license_file=
for candidate in LICENSE LICENSE.md LICENSE.txt; do
	if [ -f "$ROOT/$candidate" ]; then
		license_file="$ROOT/$candidate"
		break
	fi
done

package_target() {
	os=$1
	arch=$2
	archive_kind=$3
	target="$os"_"$arch"
	folder="bvr-cli_$VERSION"_"$target"
	stage="$STAGING/$folder"
	binary="bvr-cli"
	[ "$os" = "windows" ] && binary="bvr-cli.exe"

	mkdir -p "$stage"
	GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc GOOS="$os" GOARCH="$arch" \
		go build -trimpath \
		-ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=$VERSION" \
		-o "$stage/$binary" "$ROOT"
	cp "$ROOT/README.md" "$stage/README.md"
	[ -z "$license_file" ] || cp "$license_file" "$stage/LICENSE"

	if [ "$archive_kind" = "zip" ]; then
		(
			cd "$STAGING"
			zip -qr "$OUT/$folder.zip" "$folder"
		)
	else
		(
			cd "$STAGING"
			tar -czf "$OUT/$folder.tar.gz" "$folder"
		)
	fi
	rm -rf "$stage"
	printf 'Packaged %s\n' "$OUT/$folder.$archive_kind"
}

package_target darwin arm64 tar.gz
package_target darwin amd64 tar.gz
package_target linux amd64 tar.gz
package_target linux arm64 tar.gz
package_target windows amd64 zip
package_target windows arm64 zip

rm -rf "$STAGING"
printf 'Beta packages are in %s\n' "$OUT"
