#!/bin/sh
set -eu

# Build locally testable beta artifacts for every currently supported
# OS/architecture pair. GoReleaser remains the release/signing path.
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
VERSION=1.2.7
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
	GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc GOGC=20 GOOS="$os" GOARCH="$arch" \
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

# Arch Linux packages (native .pkg.tar.zst) for local AUR/smoke testing.
# Uses nfpm if available; otherwise skips with a note.
package_archlinux() {
	os=$1
	arch=$2
	if ! command -v nfpm >/dev/null 2>&1; then
		printf 'nfpm not installed; skipping archlinux package for %s/%s (install from goreleaser/nfpm)\n' "$os" "$arch"
		return
	fi
	pkgarch=${arch#arm}
	[ "$arch" = "arm64" ] && pkgarch="arm64" || pkgarch="x86_64"
	# Build the nfpm config dynamically.
	bin_src="$OUT/bvr-cli_${VERSION}_${os}_${arch}"
	cat > "$STAGING/nfpm_$arch.yaml" <<EOF
name: "bvr-cli"
arch: "$arch"
platform: "linux"
version: "$VERSION"
section: "default"
priority: "extra"
maintainer: "BVR <bvrcli@users.noreply.github.com>"
description: "BVR is a keyboard-first terminal AI workspace for projects, local models, tools, and automation."
vendor: "BVR"
homepage: "https://bvr-cli.com"
license: "FSL-1.1-MIT"
provides:
  - bvr-cli
conflicts:
  - bvr-cli
contents:
  - src: "$bin_src/bvr-cli"
    dst: /usr/bin/bvr-cli
  - src: /dev/null
    dst: /usr/share/bash-completion/completions/bvr-cli
  - src: /dev/null
    dst: /usr/share/zsh/site-functions/_bvr-cli
  - src: /dev/null
    dst: /usr/share/fish/vendor_completions.d/bvr-cli.fish
  - src: /dev/null
    dst: /usr/share/man/man1/bvr-cli.1.gz
  - src: "$bin_src/README.md"
    dst: /usr/share/doc/bvr-cli/README.md
  - src: "$license_file"
    dst: /usr/share/licenses/bvr-cli/LICENSE.md
EOF
	(
		cd "$STAGING"
		nfpm package -f "nfpm_$arch.yaml" -p archlinux -t "$OUT" 2>/dev/null \
			|| nfpm package -c "nfpm_$arch.yaml" -p archlinux -t "$OUT"
	)
	rm -f "$STAGING/nfpm_$arch.yaml"
	# nfpm emits pkgver-pkgrel-arch.pkg.tar.zst; rename to include OS too.
	if [ -f "$OUT/bvr-cli_${VERSION}-1_${pkgarch}.pkg.tar.zst" ]; then
		mv "$OUT/bvr-cli_${VERSION}-1_${pkgarch}.pkg.tar.zst" \
		   "$OUT/bvr-cli_${VERSION}_${os}_${pkgarch}.pkg.tar.zst"
		printf 'Packaged archlinux %s/%s -> %s\n' "$os" "$arch" \
			"$OUT/bvr-cli_${VERSION}_${os}_${pkgarch}.pkg.tar.zst"
	else
		printf 'nfpm failed to build archlinux package for %s/%s\n' "$os" "$arch"
	fi
}

# Arch packages only if the OS/arch tarball already exists.
for spec in "linux amd64" "linux arm64"; do
	set -- $spec
	if [ -f "$OUT/bvr-cli_${VERSION}_$1_$2.tar.gz" ]; then
		package_archlinux "$1" "$2"
	fi
done

rm -rf "$STAGING"
printf 'Beta packages are in %s\n' "$OUT"
