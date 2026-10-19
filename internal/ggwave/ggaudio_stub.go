//go:build !ggwave || !cgo

package ggwave

// This file is the default audio backend for BVR's standard release build,
// which ships with CGO_ENABLED=0 (see AGENTS.md). Without it the package fails
// to compile, because the PortAudio backend in ggaudio_portaudio.go is gated
// behind the `ggwave` + `cgo` tags and therefore defines no backend at all.
//
// The transport seam in ggaudio.go stays compiled either way, so callers keep
// the same types and get a clear "not compiled in" error instead of a build
// failure or a silent no-op device.

// newCaptureSource reports that no capture backend is compiled in. The error is
// the sentinel ErrNoCaptureSource so callers can branch on it with errors.Is
// and offer the native build as the remedy.
func newCaptureSource(AudioConfig) (CaptureSource, error) {
	return nil, ErrNoCaptureSource
}

// newPlaybackSink reports that no playback backend is compiled in. See
// newCaptureSource; the sentinel is ErrNoPlaybackSink.
func newPlaybackSink(AudioConfig) (PlaybackSink, error) {
	return nil, ErrNoPlaybackSink
}
