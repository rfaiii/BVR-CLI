package dialog

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/styles"
	"github.com/stretchr/testify/require"
)

func TestFixedLinesKeepsViewportStable(t *testing.T) {
	lines := fixedLines([]string{
		"/Users/richavery/Projects/bvr-cli-go/this-is-a-very-long-path-that-must-not-wrap",
		"second line",
		"third line should be clipped",
	}, 20, 2)

	require.Len(t, lines, 2)
	for _, line := range lines {
		require.Equal(t, 20, len([]rune(line)))
	}
	require.Contains(t, lines[0], "…")
	require.Equal(t, "second line         ", lines[1])
}

func TestFixedLinesHandlesEmptyViewport(t *testing.T) {
	require.Empty(t, fixedLines([]string{"content"}, 0, 3))
	require.Empty(t, fixedLines([]string{"content"}, 10, 0))
}

func TestFixedLinesClipsUnbrokenTextAndNormalizesCarriageReturns(t *testing.T) {
	lines := fixedLines([]string{"0123456789abcdef\rOVERWRITE"}, 8, 1)

	require.Len(t, lines, 1)
	require.Equal(t, 8, ansi.StringWidth(lines[0]))
	require.NotContains(t, lines[0], "\r")
	require.Equal(t, "0123456…", lines[0])
}

func TestFixedLinesDoesNotPreserveRtfControlSequencesAsLayoutCommands(t *testing.T) {
	lines := fixedLines([]string{"{\\rtf1\\ansi\\ansicpg1252\\cocoartf2868"}, 16, 1)

	require.Len(t, lines, 1)
	require.Equal(t, 16, ansi.StringWidth(lines[0]))
	require.NotContains(t, lines[0], "\n")
	require.NotContains(t, lines[0], "\r")
}

func TestFileBrowserKeepsSelectedEntryVisible(t *testing.T) {
	f := &FileBrowser{
		entries:  make([]fileBrowserEntry, 12),
		selected: 0,
	}

	f.selected = 8
	f.ensureSelectedVisible(4)
	require.Equal(t, 5, f.scroll)

	f.selected = 2
	f.ensureSelectedVisible(4)
	require.Equal(t, 2, f.scroll)

	f.selected = 11
	f.ensureSelectedVisible(4)
	require.Equal(t, 8, f.scroll)
}

func TestScrollbarLinesShowsMovingBlock(t *testing.T) {
	top := scrollbarLines(5, 20, 0)
	bottom := scrollbarLines(5, 20, 15)

	require.Len(t, top, 5)
	require.Len(t, bottom, 5)
	require.Equal(t, "█", top[0])
	require.Equal(t, "█", bottom[4])
	require.NotEqual(t, top, bottom)
}

func TestCloseLabelUsesUppercaseButton(t *testing.T) {
	require.Equal(t, "[X]", closeLabel(40))
	require.Equal(t, "X", closeLabel(8))
}

func TestFinderPaneWidthsLeavesRightGutter(t *testing.T) {
	left, right, list := finderPaneWidths(80)

	require.Equal(t, left-1, list)
	require.Equal(t, 77, left+1+right)
}

func TestFinderPaneWidthsHandlesNarrowContent(t *testing.T) {
	left, right, list := finderPaneWidths(2)

	require.GreaterOrEqual(t, left, 1)
	require.GreaterOrEqual(t, right, 1)
	require.GreaterOrEqual(t, list, 1)
}

func TestFileBrowserIgnoresStaleAsyncResults(t *testing.T) {
	f := &FileBrowser{
		dir:      "/project",
		loadSeq:  2,
		entries:  []fileBrowserEntry{{name: "new.txt", path: "/project/new.txt"}},
		selected: 0,
		preview:  "new preview",
	}

	f.HandleMsg(fileBrowserLoadMsg{
		seq:     1,
		dir:     "/project",
		entries: []fileBrowserEntry{{name: "old.txt", path: "/project/old.txt"}},
		preview: "old preview",
	})

	require.Equal(t, "new preview", f.preview)
	require.Equal(t, "new.txt", f.entries[0].name)

	f.HandleMsg(fileBrowserPreviewMsg{seq: 1, path: "/project/new.txt", preview: "old preview"})
	require.Equal(t, "new preview", f.preview)
}

func TestFileBrowserReloadCommandLoadsDirectoryOffUpdatePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "example.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0o600))
	f := &FileBrowser{dir: dir}

	cmd := f.requestReload()
	require.True(t, f.loading)
	msg := cmd()
	load, ok := msg.(fileBrowserLoadMsg)
	require.True(t, ok)
	require.NoError(t, load.err)
	require.Len(t, load.entries, 1)
	require.Equal(t, "hello", load.preview)
}

func TestFileBrowserLoadErrorIsVisible(t *testing.T) {
	f := &FileBrowser{dir: "/missing", loadSeq: 1}
	f.HandleMsg(fileBrowserLoadMsg{seq: 1, dir: "/missing", err: errors.New("directory unavailable")})

	require.False(t, f.loading)
	require.Equal(t, "directory unavailable", f.preview)
	require.Equal(t, "Metadata: unavailable", f.metadata)
}

func TestFileBrowserDrawStaysInsideScreenBounds(t *testing.T) {
	for _, size := range []image.Point{{X: 120, Y: 32}, {X: 42, Y: 10}, {X: 12, Y: 6}} {
		t.Run(fmt.Sprintf("%dx%d", size.X, size.Y), func(t *testing.T) {
			t.Parallel()
			theme := styles.DarkBvrTheme()
			f := &FileBrowser{
				com:      &common.Common{Styles: &theme},
				dir:      "/project/with/a/very/long/path",
				preview:  strings.Repeat("long preview token ", 100),
				metadata: "Metadata: a file",
				entries:  []fileBrowserEntry{{name: strings.Repeat("filename", 20), path: "/project/file.txt"}},
			}
			scr := uv.NewScreenBuffer(size.X, size.Y)
			f.Draw(scr, image.Rect(0, 0, size.X, size.Y))

			for _, line := range strings.Split(scr.Render(), "\n") {
				require.LessOrEqual(t, ansi.StringWidth(line), size.X)
			}
			require.True(t, f.closeRect.In(image.Rect(0, 0, size.X, size.Y)) || f.closeRect.Empty())
		})
	}
}
