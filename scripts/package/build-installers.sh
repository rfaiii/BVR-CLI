#!/bin/sh
# Build the installer artifacts for every supported platform from a clean
# checkout, into dist/installers/.
#
#   ./scripts/package/build-installers.sh [version]
#
# This is the local path. GoReleaser remains the signing/notarizing release
# path, and its config is a PRO-licensed file, so nothing here signs anything:
# the macOS DMGs are unsigned and the Windows EXE is unsigned. Both will raise
# a Gatekeeper / SmartScreen prompt on first run.
#
# Artifact names match the assets already on the GitHub releases page, which is
# what the winget manifest and the Homebrew tap point at.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
VERSION=${1:-1.2.7}
OUT=${OUT_DIR:-"$ROOT/dist/installers"}
STAGING="$OUT/.staging"

LDFLAGS="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=$VERSION"

for tool in nfpm makensis hdiutil; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		printf 'missing %s; cannot build installers here\n' "$tool" >&2
		exit 1
	fi
done

rm -rf "$STAGING"
mkdir -p "$OUT" "$STAGING"

license_file=
for candidate in LICENSE LICENSE.md LICENSE.txt; do
	if [ -f "$ROOT/$candidate" ]; then
		license_file="$ROOT/$candidate"
		break
	fi
done
[ -n "$license_file" ] || { echo 'no LICENSE file found' >&2; exit 1; }

build_binary() {
	os=$1
	arch=$2
	dst=$3
	mkdir -p "$(dirname "$dst")"
	GOWORK=off CGO_ENABLED=0 GOEXPERIMENT=greenteagc \
		GOOS="$os" GOARCH="$arch" \
		go build -trimpath -ldflags="$LDFLAGS" -o "$dst" "$ROOT"
	printf 'built %s/%s\n' "$os" "$arch"
}

# --- macOS: unsigned UDZO disk images ---------------------------------------
# hdiutil can only lay down a HFS+/APFS volume for the host platform, and both
# Intel and Apple Silicon Macs run the x86_64 build under Rosetta anyway, but
# each architecture still gets its own image so the native one is what installs.
dmg() {
	arch=$1
	stage="$STAGING/dmg-$arch"
	mkdir -p "$stage"
	build_binary darwin "$arch" "$stage/bvr-cli"
	cp "$ROOT/README.md" "$stage/README.txt"
	cp "$license_file" "$stage/LICENSE.txt"
	dmg_out="$OUT/bvr-cli_${VERSION}_darwin_${arch}.dmg"
	rm -f "$dmg_out"
	hdiutil create -quiet -volname "BVR CLI $VERSION" -srcfolder "$stage" \
		-ov -format UDZO "$dmg_out"
	rm -rf "$stage"
	printf 'Packaged %s\n' "$dmg_out"
}

dmg arm64
dmg amd64

# --- Linux: deb for Ubuntu, .pkg.tar.zst for Omarchy/Arch -------------------
# The nfpm configs are tracked templates; nfpm expands ${VARS} from the
# environment itself, so the version never has to be edited into a file.
linux_package() {
	format=$1
	arch=$2
	case "$format" in
		deb)
			# Debian spells architectures the way Go does, so this is the
			# build arch unchanged.
			BVR_DEB_ARCH=$arch
			export BVR_DEB_ARCH
			config=packaging/nfpm/deb.yaml
			;;
		archlinux)
			# Arch does not: pacman matches these against `uname -m`.
			case "$arch" in
				amd64) BVR_ARCH_ARCH=x86_64 ;;
				arm64) BVR_ARCH_ARCH=aarch64 ;;
				*) printf 'unhandled arch %s\n' "$arch" >&2; return 1 ;;
			esac
			export BVR_ARCH_ARCH
			config=packaging/nfpm/archlinux.yaml
			;;
		*) printf 'unknown format %s\n' "$format" >&2; return 1 ;;
	esac

	bin="$ROOT/dist/pkg-staging/bvr-cli"
	rm -rf "$ROOT/dist/pkg-staging"
	build_binary linux "$arch" "$bin"
	export BVR_VERSION="$VERSION"

	# nfpm resolves relative contents paths against the working directory, and
	# the tracked configs use repo-relative paths, so run from the repo root.
	(
		cd "$ROOT"
		nfpm package -f "$config" -p "$format" -t "$OUT"
	)

	case "$format" in
		deb) mv "$OUT/bvr-cli_${VERSION}_$arch.deb" "$OUT/bvr-cli_${VERSION}_linux_$arch.deb" ;;
	esac
	rm -rf "$(dirname "$bin")"
	printf 'Packaged %s/%s\n' "$format" "$arch"
}

linux_package deb amd64
linux_package archlinux amd64
linux_package archlinux arm64

# --- Windows: NSIS per-user installer ----------------------------------------
# winget declares InstallerType: exe, so this slot has to be a real installer.
# The v1.2.5 asset with this name was the bare console binary, which winget
# would have "installed" by launching the TUI.
win_arch=amd64
win_stage="$STAGING/windows-$win_arch"
mkdir -p "$win_stage"
build_binary windows "$win_arch" "$win_stage/bvr-cli.exe"
cp "$ROOT/README.md" "$win_stage/README.md"
cp "$license_file" "$win_stage/LICENSE.md"
makensis -V2 \
	-DVERSION="$VERSION" \
	-DBINARY="$win_stage/bvr-cli.exe" \
	-DREADME="$win_stage/README.md" \
	-DLICENSE="$win_stage/LICENSE.md" \
	-DOUTFILE="$OUT/bvr-cli_${VERSION}_windows_${win_arch}.exe" \
	"$ROOT/packaging/nsis/bvr-cli.nsi"
rm -rf "$win_stage"
printf 'Packaged %s\n' "$OUT/bvr-cli_${VERSION}_windows_${win_arch}.exe"

printf '\nInstallers for %s:\n' "$VERSION"
ls -1 "$OUT" | grep -v '^\.'
