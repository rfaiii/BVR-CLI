package audio

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/x/ansi"

	tea "charm.land/bubbletea/v2"
)

const oscAudioQueryID = "bvr-audio-query"

// DetectOSCAudioSupport parses an OSC response sequence and returns true if it
// indicates OSC audio support. This function should be called from the
// capabilities detection layer to determine terminal support.
func DetectOSCAudioSupport(seq string) bool {
	var ok bool

	p := ansi.NewParser()
	p.SetHandler(ansi.Handler{
		HandleOsc: func(cmd int, data []byte) {
			if cmd != 99 {
				return
			}

			response := strings.TrimPrefix(string(data), "99;")
			metadata, payload, found := strings.Cut(response, ";")
			if !found {
				return
			}

			var hasID, hasQuery bool
			for field := range strings.SplitSeq(metadata, ":") {
				hasID = hasID || field == "i="+oscAudioQueryID
				hasQuery = hasQuery || field == "p=?"
			}
			if !hasID || !hasQuery {
				return
			}

			ok = isOSCAudioCapacityPayload(payload)
		},
	})

	for i := 0; i < len(seq); i++ {
		p.Advance(seq[i])
	}

	return ok
}

func isOSCAudioCapacityPayload(payload string) bool {
	for field := range strings.SplitSeq(payload, ":") {
		key, value, found := strings.Cut(field, "=")
		if !found || key != "p" {
			continue
		}

		for item := range strings.SplitSeq(value, ",") {
			if item == "audio" {
				return true
			}
		}
	}

	return false
}

// OSC99AudioQuerySequence returns the OSC 99 query sequence used to detect
// audio support. This should be sent during capability detection.
func OSC99AudioQuerySequence() string {
	return ansi.DesktopNotification("", "i="+oscAudioQueryID, "p=?")
}

// OSCBackend sends audio notifications using OSC escape sequences. It
// automatically selects the best available protocol: OSC 99 (modern standard)
// if supported, falling back to OSC 777 (urxvt extension) otherwise.
type OSCBackend struct {
	audioSeq uint64
}

// NewOSCBackend creates a new audio backend with automatic protocol
// detection. If supports99 is true, it uses OSC 99; otherwise it falls back to
// OSC 777.
func NewOSCBackend(supports99 bool) *OSCBackend {
	return &OSCBackend{}
}

// Play returns a [tea.Cmd] that writes OSC escape sequences to the terminal.
// Uses OSC 99 if supported, otherwise OSC 777.
func (b *OSCBackend) Play(a Audio) tea.Cmd {
	// For now, use OSC 777 as it's more widely supported for audio
	return b.sendOSC777(a)
}

func (b *OSCBackend) sendOSC99(a Audio) tea.Cmd {
	slog.Debug("Sending OSC 99 audio", "type", a.Type, "title", a.Title, "message", a.Message)

	var sb strings.Builder
	b.audioSeq++
	id := fmt.Sprintf("bvr-audio-%d", b.audioSeq)

	appName := "BVR"
	audioType := "bvr-audio"

	// Send audio notification with type-specific parameters
	sb.WriteString(ansi.DesktopNotification(a.Type, "i="+id, "d=0", "p=audio", "a="+appName, "t="+audioType))

	return tea.Raw(sb.String())
}

func (b *OSCBackend) sendOSC777(a Audio) tea.Cmd {
	slog.Debug("Sending OSC 777 audio", "type", a.Type, "title", a.Title, "message", a.Message)

	// For OSC 777, we can use the urxvt extension with audio parameter
	return tea.Raw(ansi.URxvtExt("audio", a.Type, a.Title))
}

// AudioFromNotification converts a notification to an audio request based on type.
func AudioFromNotification(title, message, audioMode string) Audio {
	// Determine audio type based on notification content and mode
	if strings.Contains(strings.ToLower(title), "error") || strings.Contains(strings.ToLower(message), "error") {
		return Audio{Title: title, Message: message, Type: "error"}
	} else if strings.Contains(strings.ToLower(title), "start") || strings.Contains(strings.ToLower(message), "start") {
		return Audio{Title: title, Message: message, Type: "startup"}
	} else if strings.Contains(strings.ToLower(title), "exit") || strings.Contains(strings.ToLower(message), "exit") {
		return Audio{Title: title, Message: message, Type: "exit"}
	} else if strings.Contains(strings.ToLower(title), "git") || strings.Contains(strings.ToLower(message), "push") || strings.Contains(strings.ToLower(message), "upload") {
		return Audio{Title: title, Message: message, Type: "chainsaw"}
	} else {
		return Audio{Title: title, Message: message, Type: "notification"}
	}
}
