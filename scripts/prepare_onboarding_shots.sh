#!/usr/bin/env bash
# Regenerate the screenshots embedded in the first-run tour.
#
# The tour paints inside the terminal, at most ~96 cells across. A 2880x1800
# capture carries 30x the pixels that can ever reach a screen, so the committed
# assets are resized to 960x600 and quantized to a terminal-sized palette:
# ~205 KB for all four instead of 3.7 MB. The full-resolution masters stay in
# resources/screenshots/ and are what README and the docs link to.
#
# Usage: scripts/prepare_onboarding_shots.sh
# Requires: ImageMagick (magick) and pngquant.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
src_dir="$root/resources/screenshots"
out_dir="$root/internal/ui/model/shots"

# source -> tour asset. Each target is named for the step that shows it.
mappings=(
  "HOMESCREEN-ZOOM.png:home.png"
  "BVR-CLI-HOME-MENU.png:models.png"
  "COMMAND-MENU.png:commands.png"
  "PINK-THEME-MODELS.png:themes.png"
)

width=960
height=600

command -v magick >/dev/null 2>&1 || { echo "magick not found" >&2; exit 1; }
command -v pngquant >/dev/null 2>&1 || { echo "pngquant not found" >&2; exit 1; }

mkdir -p "$out_dir"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

total_before=0
total_after=0
for mapping in "${mappings[@]}"; do
  src="${mapping%%:*}"
  dst="${mapping##*:}"
  [ -f "$src_dir/$src" ] || { echo "missing source: $src_dir/$src" >&2; exit 1; }

  # -strip drops the editor chunks that can add kilobytes per file on their own.
  magick "$src_dir/$src" -resize "${width}x${height}" -strip "$work/$dst"
  pngquant --quality 60-95 --speed 1 --force --output "$out_dir/$dst" "$work/$dst"

  before=$(stat -f%z "$src_dir/$src" 2>/dev/null || stat -c%s "$src_dir/$src")
  after=$(stat -f%z "$out_dir/$dst" 2>/dev/null || stat -c%s "$out_dir/$dst")
  total_before=$((total_before + before))
  total_after=$((total_after + after))
  printf '%-26s %8s -> %8s  %s\n' "$src" "$before" "$after" "$dst"
done

printf 'tour assets: %d -> %d bytes (%d KB embedded)\n' "$total_before" "$total_after" "$((total_after / 1024))"
