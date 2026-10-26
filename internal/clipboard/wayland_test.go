package clipboard

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectKeepsX11SessionsOnTheNativeBackend(t *testing.T) {
	present := func(name string) (string, error) { return "/usr/bin/" + name, nil }
	missing := func(string) (string, error) { return "", errors.New("not found") }

	require.Equal(t, backend{}, detect("", present), "no WAYLAND_DISPLAY means do not touch the X11 path")
	require.Equal(t, backend{}, detect("wayland-1", missing), "no wl-clipboard installed")

	full := detect("wayland-1", present)
	require.Equal(t, "/usr/bin/wl-copy", full.copyBin)
	require.Equal(t, "/usr/bin/wl-paste", full.pasteBin)
}

// A partial toolset is still worth using: writes through the CLI, reads falling
// back to whatever else is available.
func TestDetectHandlesHalfAToolset(t *testing.T) {
	onlyCopy := detect("wayland-1", func(name string) (string, error) {
		if name == "wl-copy" {
			return "/usr/bin/wl-copy", nil
		}
		return "", errors.New("not found")
	})
	require.True(t, onlyCopy.canWrite())
	require.False(t, onlyCopy.canRead())

	onlyPaste := detect("wayland-1", func(name string) (string, error) {
		if name == "wl-paste" {
			return "/usr/bin/wl-paste", nil
		}
		return "", errors.New("not found")
	})
	require.False(t, onlyPaste.canWrite())
	require.True(t, onlyPaste.canRead())
}

func TestPasteArgs(t *testing.T) {
	// Without -n a newline is appended, which would corrupt a pasted path.
	args, ok := pasteArgs(FormatText)
	require.True(t, ok)
	require.Equal(t, []string{"-n"}, args)

	args, ok = pasteArgs(FormatImage)
	require.True(t, ok)
	require.Equal(t, []string{"--type", "image/png"}, args)

	_, ok = pasteArgs(Format(99))
	require.False(t, ok)
}
