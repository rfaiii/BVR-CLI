package audio

import "testing"

func TestQuickNotifyMappings(t *testing.T) {
	tests := map[string]string{
		"quick-notify-01": "QUICK-NOTIFY-01-FLAT.wav",
		"quick-notify-02": "QUICK-NOTIFY-02-FLAT.wav",
		"quick-notify-03": "QUICK-NOTIFY-03-FLAT.wav",
		"notification":    "QUICK-NOTIFY-01-FLAT.wav",
	}
	for event, want := range tests {
		if got := getAudioFilename(event); got != want {
			t.Errorf("getAudioFilename(%q) = %q, want %q", event, got, want)
		}
	}
}
