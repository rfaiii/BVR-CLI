// Package paintbrush renders an image as colored terminal cells by picking,
// for each cell, the glyph whose coverage best matches the pixels underneath
// it.
//
// This is github.com/jordanella/go-ansi-paintbrush
// v0.0.0-20240728195301-b7ad996ecf3d (MIT, see LICENSE.md) with three changes,
// all in paint.go and results.go:
//
//  1. Canvas.Paint waited for its worker goroutines but not for the goroutine
//     that writes their results into taskResults, so processResults read the
//     slice concurrently with those writes. go test -race flags it; the visible
//     symptom is cells that render blank or with a stale glyph.
//  2. Progress was written from the collector goroutine while GetProgress read
//     it under the mutex.
//  3. Glyph warnings were printed to stdout, which is the screen a TUI is
//     drawing on. They go to slog now.
//
// Upstream has not been touched since 2024 and @latest is still the pinned
// pseudo-version above, so the fix lives here rather than in a go.mod replace.
package paintbrush
