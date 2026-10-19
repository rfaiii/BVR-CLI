//go:build !ggwave || !cgo

package ggwave

import "errors"

// CGOAdapter is a no-op stub used by the default CGO-free release build.
//
// Every method mirrors the signature of the real adapter in ggwave_cgo.go so
// that caller code compiles unchanged against either build. Keep the two files
// in sync: an earlier version of this stub declared Encode/Decode over
// []float32 while the real adapter uses []int16, which meant any caller written
// against the stub failed to build once the native backend was compiled in.
//
// The float32 helpers mirror the real adapter's EncodeFloat32/DecodeFloat32
// rather than redefining Encode/Decode, keeping one canonical API.

// errNotCompiledIn is returned by every stub operation.
var errNotCompiledIn = errors.New("GGWave support not compiled in: build with -tags=ggwave and CGO enabled, and ensure libggwave is installed")

// CGOAdapter is a no-op stub when CGO is disabled.
type CGOAdapter struct{}

// NewCGOAdapter returns an error when GGWave support is not compiled in.
func NewCGOAdapter() (*CGOAdapter, error) {
	return nil, errNotCompiledIn
}

// Close is a no-op.
func (a *CGOAdapter) Close() {}

// Encode returns an error.
func (a *CGOAdapter) Encode(_ PairingEnvelope, _, _ int) ([]int16, error) {
	return nil, errNotCompiledIn
}

// Decode returns an error.
func (a *CGOAdapter) Decode(_ []int16) (PairingEnvelope, error) {
	return PairingEnvelope{}, errNotCompiledIn
}

// EncodeFloat32 returns an error.
func (a *CGOAdapter) EncodeFloat32(_ PairingEnvelope, _, _ int) ([]float32, error) {
	return nil, errNotCompiledIn
}

// DecodeFloat32 returns an error.
func (a *CGOAdapter) DecodeFloat32(_ []float32) (PairingEnvelope, error) {
	return PairingEnvelope{}, errNotCompiledIn
}

// EncodeCompact returns an error.
func (a *CGOAdapter) EncodeCompact(_ PairingEnvelope, _, _ int) (string, error) {
	return "", errNotCompiledIn
}
