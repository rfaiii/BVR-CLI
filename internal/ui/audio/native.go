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
	audioFunc func(title, message, audioType, volume string) error
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
		slog.Debug("Sending native audio", "title", a.Title, "message", a.Message, "type", a.Type, "volume", a.Volume)

		if err := b.audioFunc(a.Title, a.Message, a.Type, a.Volume); err != nil {
			slog.Error("Failed to send audio", "error", err)
		} else {
			slog.Debug("Audio sent successfully")
		}

		return nil
	}
}

// SetAudioFunc allows replacing the audio function for testing.
func (b *NativeAudioBackend) SetAudioFunc(fn func(title, message, audioType, volume string) error) {
	b.audioFunc = fn
}

// ResetAudioFunc resets the audio function to the default.
func (b *NativeAudioBackend) ResetAudioFunc() {
	b.audioFunc = defaultAudioFunc
}

// defaultAudioFunc is a no-op fallback for unsupported platforms.
// Actual implementations will override this in their init() functions.
var defaultAudioFunc = func(title, message, audioType, volume string) error { return nil }

var loadingState int

// getAudioFilename maps a notification type to its corresponding .wav file.
func getAudioFilename(audioType string) string {
	typeName := strings.ToLower(strings.TrimSuffix(audioType, ".wav"))

	// Explicit variations are semantic event names; callers do not need to
	// know the bundled filename casing or the processed suffix.
	if strings.HasPrefix(typeName, "chat-") {
		return "CHAT-" + strings.TrimPrefix(typeName, "chat-") + "-FLAT.wav"
	}
	if strings.HasPrefix(typeName, "quick-notify-") {
		return "QUICK-NOTIFY-" + strings.TrimPrefix(typeName, "quick-notify-") + "-FLAT.wav"
	}
	if strings.HasPrefix(typeName, "error-") {
		return "ERROR-" + strings.TrimPrefix(typeName, "error-") + "-FLAT.wav"
	}
	if strings.HasPrefix(typeName, "chainsaw-") {
		return "CHAINSAW-" + strings.TrimPrefix(typeName, "chainsaw-") + "-FLAT.wav"
	}
	if strings.HasPrefix(typeName, "drill-") {
		if strings.HasSuffix(typeName, "-02") {
			return "MENU-CLOSE-FLAT.wav"
		}
		return "MENU-OPEN-FLAT.wav"
	}

	switch typeName {
	case "startup":
		return "STARTUP-SONG-FLAT.wav"
	case "chainsaw":
		return fmt.Sprintf("CHAINSAW-%02d-FLAT.wav", rand.Intn(2)+1)
	case "chat":
		return fmt.Sprintf("CHAT-%02d-FLAT.wav", rand.Intn(4)+1)
	case "chat-open":
		return "CHAT-OPEN-FLAT.wav"
	case "chat-close":
		return "CHAT-CLOSE-FLAT.wav"
	case "incoming":
		return "INCOMING-FLAT.wav"
	case "loading":
		loadingState = (loadingState + 1) % 2
		if loadingState == 1 {
			return "LONG-LOAD-02-FLAT.wav"
		}
		return "LONG-LOAD-01-FLAT.wav"
	case "long-load":
		return fmt.Sprintf("LONG-LOAD-%02d-FLAT.wav", rand.Intn(2)+1)
	case "long-uploading":
		return "LONG-UPLOADING-FLAT.wav"
	case "connection", "connection-issue":
		return "CONNECTION-ISSUE-FLAT.wav"
	case "menu-open":
		return "MENU-OPEN-FLAT.wav"
	case "menu-close":
		return "MENU-CLOSE-FLAT.wav"
	case "quick-notify":
		return fmt.Sprintf("QUICK-NOTIFY-%02d-FLAT.wav", rand.Intn(3)+1)
	case "error":
		return fmt.Sprintf("ERROR-%02d-FLAT.wav", rand.Intn(3)+1)
	case "exit":
		return fmt.Sprintf("exit-%02d.wav", rand.Intn(3)+1)
	case "notification":
		return "QUICK-NOTIFY-01-FLAT.wav"
	case "question":
		return "WHAT-FLAT.wav"
	case "reload":
		return "RELOAD-FLAT.wav"
	case "drill":
		return "MENU-OPEN-FLAT.wav"
	case "sub":
		return "sub-01.wav"
	case "oops":
		return "BROKEN-FLAT.wav"
	case "connected":
		return "QUICK-NOTIFY-01-FLAT.wav"
	case "advert":
		return "advert-01.wav"
	case "broken":
		return "BROKEN-FLAT.wav"
	case "denied":
		return "DENIED-FLAT.wav"
	case "what":
		return "WHAT-FLAT.wav"
	default:
		return "QUICK-NOTIFY-01-FLAT.wav"
	}
}
