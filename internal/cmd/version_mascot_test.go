package cmd

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// TestVersionMascotSprite checks the contract scripts/gen_version_mascot.py
// generates against. A ragged or odd-length sprite silently clips the logo
// rather than failing to build, so assert it here.
func TestVersionMascotSprite(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, mascotRows)
	require.Zero(t, len(mascotRows)%2, "pixel rows must pair up into half blocks")

	for i, row := range mascotRows {
		require.Len(t, row, len(mascotRows[0]), "row %d is ragged", i)
		for _, c := range row {
			require.True(t, c == ' ' || (c >= '1' && c <= '4'),
				"row %d has invalid sprite byte %q", i, c)
		}
	}
}

func TestRenderVersionMascot(t *testing.T) {
	t.Parallel()

	out := renderVersionMascot()
	lines := strings.Split(out, "\n")
	require.Len(t, lines, len(mascotRows)/2)

	for i, line := range lines {
		plain := ansi.Strip(line)
		require.Equal(t, len(mascotRows[0]), ansi.StringWidth(plain),
			"row %d collapsed to the wrong width", i)
		for _, c := range plain {
			require.True(t, c == ' ' || c == '▀' || c == '▄',
				"row %d has unexpected glyph %q", i, c)
		}
	}
	require.Contains(t, out, "38;2;59;246;107", "expected the brand accent in the palette")
}
