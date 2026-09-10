package ggwave

import (
	"strings"
	"testing"
	"time"
)

func TestPairingRoundTrip(t *testing.T) {
	envelope := PairingEnvelope{NodeID: "node-a", Endpoint: "https://example.test", Token: "short-lived", ExpiresAt: time.Now().Add(time.Minute).Unix()}
	compact, err := CompactText(envelope)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := ParseCompactText(compact)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.NodeID != envelope.NodeID || decoded.Endpoint != envelope.Endpoint {
		t.Fatalf("decoded envelope mismatch: %#v", decoded)
	}
}

func TestWaveformWidthAndMotion(t *testing.T) {
	a := Waveform(20, 1, ModeIdle)
	b := Waveform(20, 2, ModeTransmitting)
	if len([]rune(a)) != 20 || len([]rune(b)) != 20 {
		t.Fatalf("unexpected waveform widths: %d, %d", len([]rune(a)), len([]rune(b)))
	}
	if strings.TrimSpace(a) == strings.TrimSpace(b) {
		t.Fatal("waveform should change with frame/mode")
	}
}
