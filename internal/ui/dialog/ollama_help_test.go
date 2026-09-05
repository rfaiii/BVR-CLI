package dialog

import (
	"image"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/styles"
	"github.com/stretchr/testify/require"
)

func TestOllamaHowToIncludesPlatformInstructions(t *testing.T) {
	for _, text := range []string{"macOS", "Windows", "Linux", "OLLAMA_HOST", "ctrl+l", "pull"} {
		require.Contains(t, ollamaHowTo, text)
	}
}

func TestOllamaHowToDrawIsBounded(t *testing.T) {
	theme := styles.DarkBvrTheme()
	dialog := NewOllamaHowTo(&common.Common{Styles: &theme})
	screen := uv.NewScreenBuffer(60, 16)
	dialog.Draw(screen, image.Rect(0, 0, 60, 16))
	rendered := screen.Render()
	for _, line := range splitLines(rendered) {
		require.LessOrEqual(t, ansi.StringWidth(line), 60)
	}
	require.Contains(t, rendered, "# Ollama How-To")
	require.Contains(t, rendered, "BVR can discover")
}

func splitLines(value string) []string {
	var lines []string
	start := 0
	for i, r := range value {
		if r == '\n' {
			lines = append(lines, value[start:i])
			start = i + 1
		}
	}
	return append(lines, value[start:])
}
