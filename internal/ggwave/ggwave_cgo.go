//go:build ggwave && cgo

package ggwave

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo LDFLAGS: -L${SRCDIR}/build -lggwave -lc++

#include <stdlib.h>
#include <stdio.h>
#include <ggwave/ggwave.h>

// Helper to get default parameters
ggwave_Parameters getDefaultParams(void) {
    return ggwave_getDefaultParameters();
}

// Create a GGWave instance with default parameters
ggwave_Instance createInstance(void) {
    ggwave_Parameters params = ggwave_getDefaultParameters();
    // Use I16 format like the test does
    params.sampleFormatInp = GGWAVE_SAMPLE_FORMAT_I16;
    params.sampleFormatOut = GGWAVE_SAMPLE_FORMAT_I16;
    fprintf(stderr, "C: payloadLength=%d\n", params.payloadLength);
    fprintf(stderr, "C: sampleRateInp=%f\n", params.sampleRateInp);
    fprintf(stderr, "C: sampleRateOut=%f\n", params.sampleRateOut);
    fprintf(stderr, "C: sampleRate=%f\n", params.sampleRate);
    fprintf(stderr, "C: samplesPerFrame=%d\n", params.samplesPerFrame);
    fprintf(stderr, "C: soundMarkerThreshold=%f\n", params.soundMarkerThreshold);
    fprintf(stderr, "C: sampleFormatInp=%d\n", params.sampleFormatInp);
    fprintf(stderr, "C: sampleFormatOut=%d\n", params.sampleFormatOut);
    fprintf(stderr, "C: operatingMode=%d\n", params.operatingMode);
    ggwave_Instance instance = ggwave_init(params);
    fprintf(stderr, "C: instance=%d\n", instance);
    return instance;
}

// Encode with query to get size
int encodeQuerySize(ggwave_Instance instance, const void *payload, int payloadSize, ggwave_ProtocolId protocolId, int volume) {
    return ggwave_encode(instance, payload, payloadSize, protocolId, volume, NULL, 1);
}

// Encode for real
int encodeReal(ggwave_Instance instance, const void *payload, int payloadSize, ggwave_ProtocolId protocolId, int volume, void *buffer) {
    return ggwave_encode(instance, payload, payloadSize, protocolId, volume, buffer, 0);
}

// Decode
int decode(ggwave_Instance instance, const void *waveform, int waveformSize, void *payload) {
    return ggwave_decode(instance, waveform, waveformSize, payload);
}
*/
import "C"

import (
	"encoding/base64"
	"errors"
	"unsafe"
)

// CGOAdapter wraps the native GGWave C library for encoding/decoding.
type CGOAdapter struct {
	instance C.ggwave_Instance
}

// NewCGOAdapter creates a new GGWave adapter with default parameters.
// Uses audible fast protocol for good balance of speed and robustness.
func NewCGOAdapter() (*CGOAdapter, error) {
	println("Go: Calling C.createInstance()")
	instance := C.createInstance()
	println("Go: C.createInstance() returned:", instance)
	if instance < 0 {
		return nil, errors.New("failed to initialize GGWave instance")
	}

	return &CGOAdapter{
		instance: instance,
	}, nil
}

// Close frees the GGWave instance.
func (a *CGOAdapter) Close() {
	if a.instance != 0 {
		C.ggwave_free(a.instance)
		a.instance = 0
	}
}

// Protocol constants matching ggwave_ProtocolId
const (
	ProtocolAudibleNormal     = C.GGWAVE_PROTOCOL_AUDIBLE_NORMAL
	ProtocolAudibleFast       = C.GGWAVE_PROTOCOL_AUDIBLE_FAST
	ProtocolAudibleFastest    = C.GGWAVE_PROTOCOL_AUDIBLE_FASTEST
	ProtocolUltrasoundNormal  = C.GGWAVE_PROTOCOL_ULTRASOUND_NORMAL
	ProtocolUltrasoundFast    = C.GGWAVE_PROTOCOL_ULTRASOUND_FAST
	ProtocolUltrasoundFastest = C.GGWAVE_PROTOCOL_ULTRASOUND_FASTEST
	ProtocolDTNormal          = C.GGWAVE_PROTOCOL_DT_NORMAL
	ProtocolDTFast            = C.GGWAVE_PROTOCOL_DT_FAST
	ProtocolDTFastest         = C.GGWAVE_PROTOCOL_DT_FASTEST
	ProtocolMTNormal          = C.GGWAVE_PROTOCOL_MT_NORMAL
	ProtocolMTFast            = C.GGWAVE_PROTOCOL_MT_FAST
	ProtocolMTFastest         = C.GGWAVE_PROTOCOL_MT_FASTEST
)

// DefaultProtocol is the recommended protocol for general use.
const DefaultProtocol = ProtocolAudibleFast

// DefaultVolume is the recommended volume level (0-100).
const DefaultVolume = 25

// Encode encodes a PairingEnvelope into audio samples (int16).
// Returns the audio samples as []int16.
func (a *CGOAdapter) Encode(envelope PairingEnvelope, protocol int, volume int) ([]int16, error) {
	payload, err := EncodePairing(envelope)
	if err != nil {
		return nil, err
	}

	// Query required size in bytes
	cPayload := unsafe.Pointer(&payload[0])
	size := C.encodeQuerySize(a.instance, cPayload, C.int(len(payload)), C.ggwave_ProtocolId(protocol), C.int(volume))
	if size < 0 {
		return nil, errors.New("failed to query encode size")
	}

	// Allocate buffer for int16 samples (2 bytes each)
	numSamples := size / 2
	buffer := make([]int16, numSamples)

	// Encode for real
	cBuffer := unsafe.Pointer(&buffer[0])
	written := C.encodeReal(a.instance, cPayload, C.int(len(payload)), C.ggwave_ProtocolId(protocol), C.int(volume), cBuffer)
	if written < 0 {
		return nil, errors.New("failed to encode waveform")
	}

	// Resize to actual written samples
	actualSamples := written / 2
	return buffer[:actualSamples], nil
}

// EncodeFloat32 encodes a PairingEnvelope into audio samples as float32 (normalized -1.0 to 1.0).
func (a *CGOAdapter) EncodeFloat32(envelope PairingEnvelope, protocol int, volume int) ([]float32, error) {
	samples, err := a.Encode(envelope, protocol, volume)
	if err != nil {
		return nil, err
	}

	floatSamples := make([]float32, len(samples))
	for i, s := range samples {
		floatSamples[i] = float32(s) / 32768.0
	}
	return floatSamples, nil
}

// Decode decodes audio samples (int16) into a PairingEnvelope.
// Returns the decoded envelope.
func (a *CGOAdapter) Decode(waveform []int16) (PairingEnvelope, error) {
	if len(waveform) == 0 {
		return PairingEnvelope{}, errors.New("empty waveform")
	}

	// Max payload size according to GGWave
	maxPayload := 256
	payload := make([]byte, maxPayload)

	cWaveform := unsafe.Pointer(&waveform[0])
	cPayload := unsafe.Pointer(&payload[0])

	decoded := C.decode(a.instance, cWaveform, C.int(len(waveform)*2), cPayload)
	if decoded < 0 {
		return PairingEnvelope{}, errors.New("failed to decode waveform")
	}
	if decoded == 0 {
		return PairingEnvelope{}, errors.New("no payload detected in waveform")
	}

	return DecodePairing(payload[:decoded])
}

// DecodeFloat32 decodes audio samples (float32 normalized -1.0 to 1.0) into a PairingEnvelope.
func (a *CGOAdapter) DecodeFloat32(waveform []float32) (PairingEnvelope, error) {
	intSamples := make([]int16, len(waveform))
	for i, s := range waveform {
		v := int32(s * 32767.0)
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		intSamples[i] = int16(v)
	}
	return a.Decode(intSamples)
}

// EncodeCompact encodes an envelope and returns base64-encoded waveform for transport.
// This is useful for testing or non-audio transports.
func (a *CGOAdapter) EncodeCompact(envelope PairingEnvelope, protocol int, volume int) (string, error) {
	samples, err := a.Encode(envelope, protocol, volume)
	if err != nil {
		return "", err
	}

	// Convert int16 to bytes (little-endian)
	bytes := make([]byte, len(samples)*2)
	for i, s := range samples {
		bytes[i*2] = byte(s)
		bytes[i*2+1] = byte(s >> 8)
	}

	// Use RawURLEncoding to match ParseCompactText
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
