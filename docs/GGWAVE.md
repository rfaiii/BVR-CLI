# GGWave Node Pairing

## Overview

BVR implements GGWave-based node pairing for secure, proximity-based device discovery and authentication. This document describes the complete implementation including the CGO adapter, audio transport layer, confirmation UI, and production hardening.

## Implementation Status

✅ **All four recommended follow-up items completed:**

1. **CGO Adapter** (`internal/ggwave/ggwave_cgo.go`) - Build-tagged optional CGO adapter for GGWave codec
2. **Platform Audio Transport** (`internal/ggwave/ggaudio.go`, `ggaudio_portaudio.go`) - Cross-platform audio I/O via PortAudio
3. **Explicit User Confirmation** (`internal/ui/dialog/ggwave_pairing.go`) - Interactive pairing confirmation dialog
4. **Network Transport Integration** (`internal/ui/model/ui.go`) - Pairing results register nodes for authenticated transport

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        BVR Application                          │
├─────────────────────────────────────────────────────────────────┤
│  internal/ggwave/                                               │
│  ├── pairing.go         → PairingEnvelope (portable, no CGO)   │
│  ├── waveform.go        → Visual waveform animation            │
│  ├── ggaudio.go         → Audio transport interfaces (portable)│
│  ├── ggaudio_metrics.go → Metrics collection                   │
│  ├── ggwave_cgo.go      → CGO adapter         (ggwave && cgo)  │
│  ├── ggwave_stub.go     → CGO adapter stub   (!ggwave || !cgo) │
│  ├── ggaudio_portaudio.go → PortAudio backend (ggwave && cgo)   │
│  └── ggaudio_stub.go    → Audio backend stub  (!ggwave || !cgo) │
├─────────────────────────────────────────────────────────────────┤
│  internal/ui/dialog/                                            │
│  └── ggwave_pairing.go  → Confirmation dialog                  │
├─────────────────────────────────────────────────────────────────┤
│  internal/ui/model/ui.go                                        │
│  └── PairingConfirmMsg handler → Node registration             │
└─────────────────────────────────────────────────────────────────┘
```

## Build Configuration

BVR ships with `CGO_ENABLED=0`, so the GGWave feature is behind a build tag. The
tag is **`ggwave && cgo`** on both the native codec and the PortAudio backend,
with `!ggwave || !cgo` stubs supplying the same API surface. Four build
combinations are valid:

| `CGO_ENABLED` | `-tags=ggwave` | Result |
| --- | --- | --- |
| `0` | no | Stubs (default release build) |
| `1` | no | Stubs |
| `1` | yes | Full native codec + PortAudio |
| `0` | yes | Stubs — the tag alone is not enough |

### Default Build (CGO-free)
```bash
go build .
```
The GGWave codec is NOT included. The `ggwave` package provides portable types
(`PairingEnvelope`, `Waveform`) and stub implementations that return clear errors
when CGO features are used.

Both stubs are required. Gating only the codec leaves `NewTransport` with no
`newCaptureSource`/`newPlaybackSink` and the package fails to compile, and
gating neither one makes `CGO_ENABLED=0` builds fail outright because PortAudio
has no CGO-free implementation.

### Stub and native API parity

The stubs must expose exactly the signatures of the native build, otherwise
caller code compiles in one configuration and breaks in the other. `Encode` and
`Decode` operate on `[]int16`; the float32 variants are `EncodeFloat32` and
`DecodeFloat32`. An earlier stub declared `Encode`/`Decode` over `[]float32`,
which silently guaranteed a build break the moment the native backend was
compiled in.

### CGO-Enabled Build
```bash
# 1. Build GGWave C library (one-time)
./scripts/build_ggwave.sh

# 2. Build with GGWave support
CGO_ENABLED=1 go build -tags=ggwave .

# 3. Test
CGO_ENABLED=1 go test -tags=ggwave ./internal/ggwave/...
```

`scripts/build_ggwave.sh` clones the upstream GGWave repository into
`internal/ggwave/ggwave/`, copies headers to `internal/ggwave/include/`, and
builds `libggwave.a` into `internal/ggwave/build/`. All three are generated and
gitignored.

PortAudio is initialized lazily via `sync.Once` on first use rather than in an
`init()` function, so an initialization failure surfaces as a normal error from
`NewTransport` instead of being swallowed during package import.

## Core Components

### 1. Pairing Envelope (`internal/ggwave/pairing.go`)

Portable, CGO-free data structure for node pairing:

```go
type PairingEnvelope struct {
    Version   int    `json:"v"`        // Protocol version (1)
    Kind      string `json:"kind"`     // "bvr-node-pair"
    NodeID    string `json:"node"`     // Unique node identifier
    Endpoint  string `json:"endpoint"` // Network endpoint (e.g., "https://host:port")
    Token     string `json:"token"`    // Short-lived auth token
    ExpiresAt int64  `json:"expires_at"` // Unix timestamp
}
```

**Functions:**
- `EncodePairing(envelope)` - Serialize to JSON bytes
- `DecodePairing(payload)` - Validate and deserialize
- `CompactText(envelope)` - Base64url encoding for non-audio transport
- `ParseCompactText(text)` - Reverse of CompactText

### 2. CGO Adapter (`internal/ggwave/ggwave_cgo.go`)

**Build tag:** `ggwave` (requires `CGO_ENABLED=1`)

**Types:**
```go
type CGOAdapter struct{}

func NewCGOAdapter() (*CGOAdapter, error)
func (a *CGOAdapter) Close()
func (a *CGOAdapter) Encode(envelope PairingEnvelope, protocol int, volume int) ([]int16, error)
func (a *CGOAdapter) EncodeFloat32(envelope PairingEnvelope, protocol int, volume int) ([]float32, error)
func (a *CGOAdapter) Decode(waveform []int16) (PairingEnvelope, error)
func (a *CGOAdapter) DecodeFloat32(waveform []float32) (PairingEnvelope, error)
func (a *CGOAdapter) EncodeCompact(envelope PairingEnvelope, protocol int, volume int) (string, error)
```

**Protocols:**
```go
const (
    ProtocolAudibleNormal     = C.GGWAVE_PROTOCOL_AUDIBLE_NORMAL
    ProtocolAudibleFast       = C.GGWAVE_PROTOCOL_AUDIBLE_FAST     // Default
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
```

### 3. Audio Transport (`internal/ggwave/ggaudio.go`)

**Interfaces:**
```go
type CaptureSource interface {
    io.Reader
    io.Closer
    Start() error
    Stop() error
    Config() AudioConfig
    Device() *AudioDevice
    Err() error
    AudioLevel() float64          // RMS 0.0-1.0
    SetVAD(enabled bool, threshold float64) error
}

type PlaybackSink interface {
    io.Writer
    io.Closer
    Start() error
    Stop() error
    Config() AudioConfig
    Device() *AudioDevice
    Err() error
    AudioLevel() float64          // RMS 0.0-1.0
}
```

**Transport:**
```go
type Transport struct {
    // ...
}

func NewTransport(config AudioConfig) (*Transport, error)
func (t *Transport) Capture() CaptureSource
func (t *Transport) Playback() PlaybackSink
func (t *Transport) Start() error           // With retry logic
func (t *Transport) Stop() error
func (t *Transport) Close() error
func (t *Transport) IsClosed() bool
func (t *Transport) Config() AudioConfig
```

**AudioConfig:**
```go
type AudioConfig struct {
    SampleRate      int           // Default: 48000
    Channels        int           // Default: 1 (mono)
    Format          SampleFormat  // Default: Int16
    BufferSize      int           // Default: 1024 frames
    DeviceID        string        // Empty = default device
    AutoReconnect   bool          // Default: true
    ReconnectDelay  time.Duration // Default: 500ms
    MaxReconnects   int           // Default: 3
    EnableVAD       bool          // Default: false
    VADThreshold    float64       // Default: 0.01
    Logger          *slog.Logger  // Optional
}
```

### 4. Pairing Session (`internal/ggwave/ggaudio.go`)

```go
type PairingSessionConfig struct {
    Timeout      time.Duration // Total pairing timeout
    Deadline     time.Time     // Absolute deadline
    MinFrames    int           // Min frames before decode attempt (default: 1024)
    MaxFrames    int           // Max frames before giving up (default: 10240)
    EnableMetrics bool         // Default: true
}

func NewPairingSession(transport *Transport) *PairingSession
func NewPairingSessionWithConfig(transport *Transport, config PairingSessionConfig) *PairingSession

func (s *PairingSession) Transmit(waveform []int16) error
func (s *PairingSession) Receive(timeout time.Duration) ([]int16, error)
func (s *PairingSession) Cancel()
func (s *PairingSession) IsClosed() bool
```

### 5. Confirmation Dialog (`internal/ui/dialog/ggwave_pairing.go`)

Interactive dialog showing decoded envelope details:

```
🔗  GGWAVE NODE PAIRING REQUEST

Node ID:   node-abc123
Endpoint:  https://192.168.1.100:8080
Token:     short-lived-xyz
Expires:   15:04:05 (in 45s)

  [   ]  ACCEPT    [ ✓ ]  REJECT
```

**Keys:** `y`/`Y` = accept, `n`/`N` = reject, `←`/`→` = switch, `Enter` = confirm, `Esc` = reject

**Opens via:**
```go
dialog.OpenPairingConfirm(com, envelope)
```

**Message:**
```go
type PairingConfirmMsg struct {
    Envelope *ggwave.PairingEnvelope
    Accepted bool
}
```

### 6. Node Integration (`internal/ui/model/ui.go`)

When `PairingConfirmMsg.Accepted == true`:
1. Creates `node.Device` with connection type "ggwave"
2. Registers via `node.UpsertDevice()`
3. Uses `Endpoint` as `TransportURL` for authenticated transport
4. Shows info message: "Node paired via GGWave: {nodeID}"

## Device Enumeration

```go
// List all available audio devices
devices, err := ggwave.ListAudioDevices()

// Get default devices
input, err := ggwave.GetDefaultInputDevice()
output, err := ggwave.GetDefaultOutputDevice()

// AudioDevice struct
type AudioDevice struct {
    ID            string
    Name          string
    IsInput       bool
    IsOutput      bool
    IsDefaultIn   bool
    IsDefaultOut  bool
    MaxChannels   int
    SampleRates   []int
}
```

## Metrics (`internal/ggwave/ggaudio_metrics.go`)

```go
metrics := ggwave.GlobalMetrics()

// Counters
metrics.RecordTransportStart()
metrics.RecordTransportError()
metrics.RecordTransportReconnect()
metrics.RecordSessionCreate()
metrics.RecordSessionComplete()
metrics.RecordSessionError()
metrics.RecordBytesTransmitted(n)
metrics.RecordBytesReceived(n)
metrics.RecordFramesTransmitted(n)
metrics.RecordFramesReceived(n)
metrics.RecordLatency(duration)
metrics.RecordVADActive()
metrics.RecordVADInactive()

// Snapshot
snapshot := metrics.Snapshot()
fmt.Println(snapshot.String())

// Reset
metrics.Reset()
```

**Snapshot fields:**
- Transport: starts, stops, errors, reconnects
- Sessions: created, completed, errors, cancelled
- Audio: tx/rx bytes, tx/rx frames
- Issues: underruns, overruns
- Latency: avg µs, sample count
- VAD: active, inactive counts
- Uptime

## Visual Waveform (`internal/ggwave/waveform.go`)

```go
type Mode uint8
const (
    ModeIdle Mode = iota
    ModeListening
    ModeTransmitting
    ModePaired
)

func Waveform(width int, frame uint64, mode Mode, speed ...float64) string
```

Renders animated terminal waveform (▁▂▃▄▅▆▇█) on homescreen.

## Production Hardening Features

| Feature | Implementation |
|---------|---------------|
| **Device Enumeration** | `ListAudioDevices()`, `GetDefaultInput/OutputDevice()` |
| **Retry Logic** | `AutoReconnect`, `ReconnectDelay`, `MaxReconnects` in `AudioConfig` |
| **Panic Recovery** | `safeExecute()` wrapper in PortAudio implementation |
| **Config Validation** | `AudioConfig.Validate()` |
| **Audio Level Monitoring** | `AudioLevel()` (RMS), `SetVAD()` |
| **Structured Logging** | `slog.Logger` integration via `AudioConfig.Logger` |
| **Metrics** | `MetricsCollector` with atomic counters |
| **Timeout Handling** | `PairingSessionConfig` with `Timeout`, `Deadline`, `MinFrames`, `MaxFrames` |
| **Context Cancellation** | Full `context.Context` propagation |

## Testing

```bash
# Unit tests (no CGO)
go test ./internal/ggwave/...

# Integration tests (with CGO)
CGO_ENABLED=1 go test -tags=ggwave ./internal/ggwave/...

# All tests
go test ./...
```

## Scripts

### `scripts/build_ggwave.sh`
```bash
#!/bin/bash
# Builds GGWave C library from source
# 1. Clones ggwave repo (with submodules)
# 2. Builds static library (libggwave.a)
# 3. Copies headers to internal/ggwave/include/
# 4. Copies library to internal/ggwave/build/
```

## Security Considerations

1. **Pairing envelope** contains only bootstrap credentials (node ID, endpoint, short-lived token)
2. **Token expiry** enforced - rejected if `ExpiresAt <= now`
3. **Explicit user confirmation** required before node registration
4. **Sound only for bootstrap** - ongoing communication uses authenticated network transport (HTTPS/WebSocket)
5. **No commands in envelope** - only identity and connection info

## Future Work

- [ ] macOS CoreAudio implementation (`ggaudio_darwin.go`)
- [ ] Windows WASAPI implementation
- [ ] Linux ALSA/PulseAudio/PipeWire optimization
- [ ] Mobile client audio backends (iOS AudioToolbox, Android OpenSL ES)
- [ ] Encrypted pairing envelope (age/NaCl)
- [ ] Multi-device simultaneous pairing
- [ ] Pairing history and revocation

## Reference

- GGWave C library: https://github.com/ggerganov/ggwave
- PortAudio: http://www.portaudio.com/
- Protocol: Frequency-Shift Keying (FSK) with Reed-Solomon error correction