#!/usr/bin/env bash
# Crush the bundled sounds until they stop costing anything measurable. These
# are short UI cues, so fidelity buys nothing: the whole set goes to mono,
# 8-bit unsigned PCM at 11.025 kHz — the rate and depth classic Mac system
# sounds shipped at. The output stays a plain WAV PCM file, so afplay (macOS),
# paplay and aplay (Linux) and Windows' MediaPlayer still need no codec.
#
# Re-encoding the whole file instead of editing headers in place also drops the
# 4 KB Pro Tools `FLLR` filler chunk every one of these cues carries (~158 KB
# across the set).
#
# Requires ffmpeg, with afconvert (macOS) as the fallback.
set -euo pipefail

cd "$(dirname "$0")/.."
sounds="audio/sounds"
rate="11025"

file_size() {
  if stat -f%z "$1" >/dev/null 2>&1; then
    stat -f%z "$1"
  else
    stat -c%s "$1"
  fi
}

total() {
  local sum=0 f
  for f in "$sounds"/*.wav; do
    sum=$(( sum + $(file_size "$f") ))
  done
  echo "$sum"
}

convert() {
  if command -v ffmpeg >/dev/null 2>&1; then
    ffmpeg -nostdin -loglevel error -y -i "$1" -ac 1 -ar "$rate" -c:a pcm_u8 \
      -map_metadata -1 -fflags +bitexact "$2"
  else
    afconvert -f WAVE -d "UI8@${rate}" -c 1 "$1" "$2"
  fi
}

before=$(total)

for f in "$sounds"/*.wav; do
  if [ "$(file_size "$f")" -lt 1024 ]; then
    continue
  fi
  tmp=$(mktemp "${TMPDIR:-/tmp}/bvr-sound-XXXXXX.wav")
  if convert "$f" "$tmp"; then
    mv "$tmp" "$f"
  else
    rm -f "$tmp"
    echo "skipped (conversion failed): $f" >&2
  fi
done

after=$(total)
echo "embedded sounds: $before -> $after bytes ($(( (before - after) / 1024 / 1024 )) MB saved)"
