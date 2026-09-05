// Package audio provides audio notification support for the UI.
//
// This package supports audio playback for various events and notifications:
//   - NativeAudioBackend: Uses the native OS audio system (macOS, Windows, Linux)
//   - OSSinkBackend: Uses OSC escape sequences for audio playback (SSH sessions)
//   - BellBackend: Uses terminal bell for basic audio feedback
//   - NoopBackend: A no-op backend that silently discards audio when disabled
//
// Audio backend selection is based on terminal capabilities, environment, and user config:
//   - Users can explicitly set audio in bvr.json (auto/native/osc/bell/disabled)
//   - Auto mode: SSH sessions use OSC backend
//   - Auto mode: Local sessions use native OS audio
//   - If focus events are not supported in local sessions, audio is disabled (NoopBackend)
package audio

import tea "charm.land/bubbletea/v2"

// Audio represents an audio notification request.
type Audio struct {
	Title   string
	Message string
	Type    string // "notification", "startup", "error", "exit", "chainsaw"
}

// Backend defines the interface for sending audio notifications.
// Implementations return a tea.Cmd that performs the audio playback, allowing
// each backend to choose between synchronous (native OS) and asynchronous
// (terminal escape sequences) delivery. Policy decisions (config checks,
// focus state) are handled by the caller.
type Backend interface {
	Play(a Audio) tea.Cmd
}
