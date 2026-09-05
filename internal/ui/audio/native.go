package audio

import (
	"fmt"
	"log/slog"
	"math/rand"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// NativeAudioBackend sends audio notifications using the native OS audio system.
// The actual delivery function is supplied per-platform via defaultAudioFunc;
// on unsupported platforms it is a no-op. Selection logic avoids this backend there
// and uses a terminal-based backend instead, so this is only a safety net.
// See NativeSupported.
type NativeAudioBackend struct {
	// audioFunc is the function used to send audio (swappable for testing).
	audioFunc func(title, message, audioType string) error
}

// NewNativeAudioBackend creates a new native audio backend.
func NewNativeAudioBackend() *NativeAudioBackend {
	return &NativeAudioBackend{
		audioFunc: defaultAudioFunc,
	}
}

// Play returns a command that sends audio using the native OS audio system.
func (b *NativeAudioBackend) Play(a Audio) tea.Cmd {
	return func() tea.Msg {
		slog.Debug("Sending native audio", "title", a.Title, "message", a.Message, "type", a.Type)

		if err := b.audioFunc(a.Title, a.Message, a.Type); err != nil {
			slog.Error("Failed to send audio", "error", err)
		} else {
			slog.Debug("Audio sent successfully")
		}

		return nil
	}
}

// SetAudioFunc allows replacing the audio function for testing.
func (b *NativeAudioBackend) SetAudioFunc(fn func(title, message, audioType string) error) {
	b.audioFunc = fn
}

// ResetAudioFunc resets the audio function to the default.
func (b *NativeAudioBackend) ResetAudioFunc() {
	b.audioFunc = defaultAudioFunc
}

// defaultAudioFunc is a no-op fallback for unsupported platforms.
// Actual implementations will override this in their init() functions.
var defaultAudioFunc = func(title, message, audioType string) error { return nil }

// getAudioFilename maps a notification type to its corresponding .wav file.
func getAudioFilename(audioType string) string {
	// If the audioType specifies a variation (e.g. "chat-02" or "drill-01")
	if strings.Contains(audioType, "-") && !strings.HasSuffix(audioType, ".wav") {
		return audioType + ".wav"
	}

	switch audioType {
	case "startup":
		return "startup-song-01.wav"
	case "chainsaw":
		// Could randomize between 01, 02, 03, but let's stick to 01 for now
		return "chainsaw-01.wav"
	case "chat":
		return "chat-01.wav"
	case "incoming":
		return "incoming-01.wav"
	case "loading":
		return "loading-01.wav"
	case "error":
		return fmt.Sprintf("error-%02d.wav", rand.Intn(3)+1)
	case "exit":
		return "exit-01.wav"
	case "notification":
		return "notification-01.wav"
	case "question":
		return "question-01.wav"
	case "reload":
		return "reload-01.wav"
	case "drill":
		return "drill-01.wav"
	case "sub":
		return "sub-01.wav"
	case "oops":
		return "oops-01.wav"
	case "connected":
		return "connected-01.wav"
	case "advert":
		return "advert-01.wav"
	case "broken":
		return "broken-01.wav"
	default:
		return "notification-01.wav"
	}
}
