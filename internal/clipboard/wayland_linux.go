//go:build linux

package clipboard

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

// wlTimeout bounds a clipboard call. A wedged compositor must not take the UI
// down with it, and no clipboard read should ever take seconds.
const wlTimeout = 2 * time.Second

// wayland is set by Init and stays nil until then, so an uninitialised process
// keeps using the X11 backend. It is a pointer rather than package state because
// Init runs on a goroutine started per workspace while other goroutines read it.
var wayland atomic.Pointer[backend]

func canWrite() bool {
	b := wayland.Load()
	return b != nil && b.canWrite()
}

func canRead() bool {
	b := wayland.Load()
	return b != nil && b.canRead()
}

func initFallback() bool {
	found := detect(os.Getenv("WAYLAND_DISPLAY"), exec.LookPath)
	wayland.Store(&found)
	return found.canWrite() || found.canRead()
}

func writeFallback(text string) {
	b := wayland.Load()
	if b == nil || !b.canWrite() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), wlTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, b.copyBin)
	cmd.Stdin = strings.NewReader(text)
	// wl-copy forks into the background to hold the selection; the process this
	// command starts exits once stdin has drained.
	_ = cmd.Run()
}

func readFallback(f Format) ([]byte, error) {
	b := wayland.Load()
	args, ok := pasteArgs(f)
	if b == nil || !b.canRead() || !ok {
		return nil, ErrUnsupported
	}
	ctx, cancel := context.WithTimeout(context.Background(), wlTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, b.pasteBin, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil || out.Len() == 0 {
		return nil, ErrEmpty
	}
	return out.Bytes(), nil
}
