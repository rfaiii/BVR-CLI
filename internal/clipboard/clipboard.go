// Package clipboard provides cross-platform clipboard access behind build-tag
// guards so that unsupported platforms (e.g., Android, iOS) compile without
// requiring CGO or platform-specific dependencies.
package clipboard

import (
	"errors"
	"sync"
)

// Format identifies the kind of data to read from the clipboard.
type Format int

const (
	// FormatText is plain UTF-8 text.
	FormatText Format = iota
	// FormatImage is binary image data (PNG).
	FormatImage
)

var (
	// ErrUnsupported is returned when clipboard access is not available on
	// the current platform.
	ErrUnsupported = errors.New("clipboard operations are not supported on this platform")
	// ErrEmpty is returned when the clipboard holds no data for the
	// requested format.
	ErrEmpty = errors.New("clipboard is empty or holds an unsupported format")
)

var (
	initOnce sync.Once
	initErr  error
)

// Init initializes the clipboard subsystem. On unsupported platforms it
// returns ErrUnsupported but is otherwise safe to call.
//
// It is called from a goroutine that every app.New starts, and initialization
// writes package state on Linux, so repeated calls are folded into one and the
// first result is what later callers see.
func Init() error {
	initOnce.Do(func() { initErr = initClipboard() })
	return initErr
}

// WriteText writes plain text to the system clipboard. On unsupported
// platforms it is a no-op.
func WriteText(text string) {
	writeText(text)
}

// Read returns the clipboard contents for the given format. It returns
// ErrEmpty when no matching data is present and ErrUnsupported on platforms
// without clipboard support.
func Read(f Format) ([]byte, error) {
	return read(f)
}
