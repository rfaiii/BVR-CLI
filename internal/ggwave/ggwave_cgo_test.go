//go:build ggwave
// +build ggwave

package ggwave

import (
	"encoding/base64"
	"testing"
	"time"
)

func TestCGOAdapterEncodeDecode(t *testing.T) {
	adapter, err := NewCGOAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	envelope := PairingEnvelope{
		NodeID:    "test-node-1",
		Endpoint:  "https://example.com",
		Token:     "test-token-123",
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}

	samples, err := adapter.Encode(envelope, int(DefaultProtocol), DefaultVolume)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if len(samples) == 0 {
		t.Fatal("expected non-empty waveform")
	}

	decoded, err := adapter.Decode(samples)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.NodeID != envelope.NodeID {
		t.Fatalf("node ID mismatch: got %q, want %q", decoded.NodeID, envelope.NodeID)
	}
	if decoded.Endpoint != envelope.Endpoint {
		t.Fatalf("endpoint mismatch: got %q, want %q", decoded.Endpoint, envelope.Endpoint)
	}
	if decoded.Token != envelope.Token {
		t.Fatalf("token mismatch: got %q, want %q", decoded.Token, envelope.Token)
	}
}

func TestCGOAdapterEncodeDecodeFloat32(t *testing.T) {
	adapter, err := NewCGOAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	envelope := PairingEnvelope{
		NodeID:    "test-node-float",
		Endpoint:  "https://example.com",
		Token:     "test-token-float",
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}

	samples, err := adapter.EncodeFloat32(envelope, int(DefaultProtocol), DefaultVolume)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if len(samples) == 0 {
		t.Fatal("expected non-empty waveform")
	}

	decoded, err := adapter.DecodeFloat32(samples)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.NodeID != envelope.NodeID {
		t.Fatalf("node ID mismatch: got %q, want %q", decoded.NodeID, envelope.NodeID)
	}
}

func TestCGOAdapterEncodeCompact(t *testing.T) {
	adapter, err := NewCGOAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	envelope := PairingEnvelope{
		NodeID:    "test-node-2",
		Endpoint:  "https://example.com",
		Token:     "test-token-456",
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}

	compact, err := adapter.EncodeCompact(envelope, int(DefaultProtocol), DefaultVolume)
	if err != nil {
		t.Fatalf("encode compact failed: %v", err)
	}

	if compact == "" {
		t.Fatal("expected non-empty compact encoding")
	}

	// Verify we can decode it back by decoding the base64 and then decoding the waveform
	payload, err := base64.RawURLEncoding.DecodeString(compact)
	if err != nil {
		t.Fatalf("compact encoding is not valid base64: %v", err)
	}

	// Convert bytes to int16 samples
	samples := make([]int16, len(payload)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(payload[i*2]) | int16(payload[i*2+1])<<8
	}

	decoded, err := adapter.Decode(samples)
	if err != nil {
		t.Fatalf("decode compact failed: %v", err)
	}

	if decoded.NodeID != envelope.NodeID {
		t.Fatalf("node ID mismatch: got %q, want %q", decoded.NodeID, envelope.NodeID)
	}
}

func TestCGOAdapterMultipleProtocols(t *testing.T) {
	adapter, err := NewCGOAdapter()
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	envelope := PairingEnvelope{
		NodeID:    "test-node-3",
		Endpoint:  "https://example.com",
		Token:     "test-token-789",
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}

	protocols := []int{
		int(ProtocolAudibleNormal),
		int(ProtocolAudibleFast),
		int(ProtocolAudibleFastest),
		int(ProtocolDTNormal),
		int(ProtocolDTFast),
		int(ProtocolDTFastest),
		int(ProtocolMTNormal),
		int(ProtocolMTFast),
		int(ProtocolMTFastest),
	}

	for _, protocol := range protocols {
		samples, err := adapter.Encode(envelope, protocol, DefaultVolume)
		if err != nil {
			t.Logf("protocol %d encode failed (may not be available): %v", protocol, err)
			continue
		}

		decoded, err := adapter.Decode(samples)
		if err != nil {
			t.Logf("protocol %d decode failed: %v", protocol, err)
			continue
		}

		if decoded.NodeID != envelope.NodeID {
			t.Errorf("protocol %d: node ID mismatch", protocol)
		}
	}
}
