#!/bin/sh
set -eu

SRC=${BVR_MASTER_SKILLS_DIR:-"$HOME/Documents/AI-SKILLS"}
DST=${BVR_SKILLS_DIR:-"$HOME/.agents/skills"}

if [ ! -d "$SRC" ]; then
	printf 'Master skills directory not found: %s\n' "$SRC" >&2
	exit 1
fi

mkdir -p "$DST"
installed=0
skipped=0
while IFS= read -r skill; do
	name=$(basename "$skill")
	target="$DST/$name"
	if [ -e "$target" ] || [ -L "$target" ]; then
		skipped=$((skipped+1))
		continue
	fi
	ln -s "$skill" "$target"
	installed=$((installed+1))
done <<EOF
$(find "$SRC" -mindepth 2 -maxdepth 2 -type f -name SKILL.md -print | sed 's#/SKILL.md$##' | sort)
EOF

printf 'Installed %s skill links; preserved %s existing entries in %s.\n' "$installed" "$skipped" "$DST"