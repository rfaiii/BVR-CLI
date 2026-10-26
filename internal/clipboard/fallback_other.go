//go:build !linux

package clipboard

// Every non-Linux platform uses the native backend only: pbcopy-style access on
// macOS, the Win32 clipboard on Windows, and X11 on the BSDs. The wl-clipboard
// hooks exist because Linux is the one place a session can be Wayland.
func initFallback() bool { return false }

func canWrite() bool { return false }

func canRead() bool { return false }

func writeFallback(string) {}

func readFallback(Format) ([]byte, error) { return nil, ErrUnsupported }
