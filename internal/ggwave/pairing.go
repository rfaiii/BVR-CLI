// Package ggwave contains the portable seam between BVR's node pairing UI and
// a GGWave transport. The codec is intentionally injected: the default BVR
// build is CGO-free, while a native GGWave adapter can be added later without
// coupling the TUI to microphone or speaker I/O.
package ggwave

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const protocolVersion = 1

// PairingEnvelope is the small, user-confirmed payload exchanged over sound.
// It deliberately carries a short-lived pairing token rather than commands.
type PairingEnvelope struct {
	Version   int    `json:"v"`
	Kind      string `json:"kind"`
	NodeID    string `json:"node"`
	Endpoint  string `json:"endpoint,omitempty"`
	Token     string `json:"token,omitempty"`
	ExpiresAt int64  `json:"expires_at"`
}

// EncodePairing serializes an envelope in a compact, transport-friendly form.
func EncodePairing(envelope PairingEnvelope) ([]byte, error) {
	if envelope.Kind == "" {
		envelope.Kind = "bvr-node-pair"
	}
	if envelope.Version == 0 {
		envelope.Version = protocolVersion
	}
	if envelope.NodeID == "" || envelope.ExpiresAt <= 0 {
		return nil, errors.New("node ID and expiry are required")
	}
	if time.Unix(envelope.ExpiresAt, 0).Before(time.Now()) {
		return nil, errors.New("pairing envelope has expired")
	}
	return json.Marshal(envelope)
}

// DecodePairing validates and decodes a received envelope. The caller must
// still show the details and ask the user to confirm before adding a node.
func DecodePairing(payload []byte) (PairingEnvelope, error) {
	var envelope PairingEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return PairingEnvelope{}, fmt.Errorf("decode pairing envelope: %w", err)
	}
	if envelope.Version != protocolVersion || envelope.Kind != "bvr-node-pair" {
		return PairingEnvelope{}, errors.New("unsupported pairing envelope")
	}
	if strings.TrimSpace(envelope.NodeID) == "" {
		return PairingEnvelope{}, errors.New("pairing envelope has no node ID")
	}
	if envelope.ExpiresAt <= time.Now().Unix() {
		return PairingEnvelope{}, errors.New("pairing envelope has expired")
	}
	return envelope, nil
}

// CompactText is useful for transports that prefer an ASCII-safe payload.
func CompactText(envelope PairingEnvelope) (string, error) {
	payload, err := EncodePairing(envelope)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

// ParseCompactText reverses CompactText.
func ParseCompactText(value string) (PairingEnvelope, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return PairingEnvelope{}, fmt.Errorf("decode compact pairing payload: %w", err)
	}
	return DecodePairing(payload)
}
