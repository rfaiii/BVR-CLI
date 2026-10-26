//go:build linux

package clipboard

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// wlTimeout bounds a clipboard call. A wedged compositor must not take the UI
// down with it, and no clipboard read should ever take seconds.
const wlTimeout = 2 * time.Second

// wayland is set by Init. It stays zero-valued until then, so an uninitialised
// process keeps using the X11 backend.
var wayland backend

func canWrite() bool { return wayland.canWrite() }
func canRead() bool  { return wayland.canRead() }

func initFallback() bool {
	wayland = detect(os.Getenv("WAYLAND_DISPLAY"), exec.LookPath)
	return wayland.canWrite() || wayland.canRead()
}

func writeFallback(text string) {
	if !wayland.canWrite() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), wlTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, wayland.copyBin)
	cmd.Stdin = strings.NewReader(text)
	// wl-copy forks into the background to hold the selection; the process this
	// command starts exits once stdin has drained.
	_ = cmd.Run()
}

func readFallback(f Format) ([]byte, error) {
	args, ok := pasteArgs(f)
	if !wayland.canRead() || !ok {
		return nil, ErrUnsupported
	}
	ctx, cancel := context.WithTimeout(context.Background(), wlTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, wayland.pasteBin, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil || out.Len() == 0 {
		return nil, ErrEmpty
	}
	return out.Bytes(), nil
}
