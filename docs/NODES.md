# BVR NODES — Multi-Device Connections

This document is the consolidated reference for BVR's NODE layer: the device
registry, the HTTP/WebSocket/SSH transport protocols, and GGWave audio
pairing. It merges the former `NODE_TRANSPORTS.md`, `NODE_HTTP_PROTOCOL.md`,
`NODE_WEBSOCKET_PROTOCOL.md`, `NODE_SSH_TRANSPORT.md`, and `GGWAVE.md`.

## Transport overview

BVR supports one device registry with four transport adapters:

| Scheme | Adapter | Best for |
| --- | --- | --- |
| `http://`, `https://` | HTTP/JSON | Request/response execution and health checks |
| `ws://`, `wss://` | WebSocket/JSON | Persistent connections and streaming output |
| SSH configuration | SSH | Existing computers reachable through SSH |
| `ggwave://` | GGWave Audio | Proximity-based pairing and bootstrap |

### Quick start

```sh
bvr node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN" --advertise
```

The listener exposes `/v1/node/health`, `/v1/node/capabilities`,
`/v1/node/execute`, and the `/v1/node/ws` WebSocket endpoint. Register HTTP or
WebSocket endpoints with `RegisterNodeTransport`. SSH uses `NewSSHTransport` or
`RegisterSSHTransport`. GGWave pairing uses `OpenPairingConfirm` to bootstrap
node registration.

Find the agents on the network with:

```sh
bvr node scan --wait 4s
Richards-MacBook-Air     172.20.20.20:7799      proto=1  id=Richards-MacBook-Air
```

### Local-network discovery

`internal/node/discovery.go` publishes and resolves a DNS-SD record so a
companion never has to type an IP. The service type is `_bvr._tcp.local`. Three
places spell it and are kept in sync by hand: `ServiceType` here,
`NSBonjourServices` in `mobile-cli/ios/BvrCompanion/Info.plist`, and
`NodeDiscovery.serviceType` in the app. Changing one means changing all three.

- `Advertise` runs for the lifetime of the returned `Advertisement` and closes
  it to withdraw the record. It publishes `proto`, `id`, and `name` in the TXT
  record and never the bearer token.
- `Scan`/`Browse` collect answers for a window (default `BrowseWindow`, 3 s)
  because mDNS replies arrive as separate PTR, SRV, and A records.
- `bvr node serve --advertise` refuses to start without `--token`: the agent
  executes arbitrary commands for any client that authenticates, and a
  published agent is reachable by everyone on the LAN.

Registration goes through `zeroconf.RegisterProxy` rather than
`zeroconf.Register`. The library derives the host name from `os.Hostname()`,
which on macOS returns `machine.local`; it then treats the name as already
qualified, omits the trailing dot, and every probe and announcement fails with
`dns: domain must be fully qualified`. Passing our own FQDN and IPv4 addresses
sidesteps it, and advertising explicit addresses also keeps link-local and
loopback addresses out of the A record.

The dependency costs ~0.8 MB on a stripped release binary.

`internal/node/ios_wire_test.go` pins the frame contract to byte-exact payloads
captured from Swift's encoder, so renaming `working_dir` or `exit_code` on
either side fails a test instead of silently sending an empty command.

Credentials remain in transport objects and are not stored in `Device` or shown
in the UI. The home screen and NODE settings dialog show device name and state:
grey offline, red error, green online.

The transport layer, mDNS discovery, device registry, and UI status colors are
complete. Protocol details for each transport follow. Discovery results are not
yet fed into the registry, and the registry is per-process: nothing persists a
device across runs.

## HTTP protocol

The NODE agent exposes a small authenticated JSON API for device execution.
The same `bvr node serve` listener also exposes the WebSocket streaming
endpoint described in the next section.

Tokens are required when binding outside localhost. Keep the agent behind a
trusted network or TLS/reverse proxy; the current agent provides bearer-token
authentication but does not terminate TLS itself.

### Endpoints

- `GET /v1/node/health` — returns `{ "ok": true, "protocol": "1" }`.
- `GET /v1/node/capabilities` — reports supported command families.
- `POST /v1/node/execute` — executes a command request.
- `GET /v1/node/ws` — persistent WebSocket endpoint for streaming execution.

Execute request:

```json
{
  "command": "node",
  "args": ["script.js", "--watch"],
  "working_dir": "/workspace/project"
}
```

Execute response includes `stdout`, `stderr`, `exit_code`, and an optional
`error`. Requests are limited to 1 MiB by default and inherit HTTP request
cancellation.

## WebSocket protocol

Persistent NODE connections use `ws://` or `wss://` at:

```text
/v1/node/ws
```

The client sends JSON frames with `type` values `ping` or `execute`. Execute
requests carry the same `command`, `args`, and `working_dir` fields as the HTTP
protocol. The server streams frames before the final result:

```json
{"id":"42","type":"stdout","stream":"stdout","data":"building\n"}
{"id":"42","type":"stderr","stream":"stderr","data":"warning\n"}
{"id":"42","type":"result","result":{"exit_code":0}}
```

Errors use `type: "error"` or an `error` field on the final result. Clients
reuse the connection for sequential requests and reconnect once after a
disconnect. Bearer authentication uses the same header as HTTP:

```text
Authorization: Bearer <token>
```

Start the persistent agent with the same command as the HTTP agent. The CLI
serves both HTTP and WebSocket routes from the same listener and token:

```sh
bvr node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN"
```

## SSH transport

SSH is available for computers that already expose an SSH service. It uses
standard password or private-key authentication and streams stdout/stderr
through the same NODE transport interface as HTTP and WebSocket.

Host-key verification is mandatory by default:

```go
transport, err := node.NewSSHTransport(node.SSHConfig{
    Address:         "workstation.example:22",
    User:            "rich",
    PrivateKeyPath:  "/Users/rich/.ssh/id_ed25519",
    KnownHostsPath:  "/Users/rich/.ssh/known_hosts",
})
```

`InsecureSkipHostKey: true` is available only as an explicit opt-in for
controlled environments and tests. Do not use it on untrusted networks.

The transport executes commands with shell-safe quoting and optionally changes
to `WorkingDir` before execution. Credentials are held by the transport and
are not stored in the device registry or rendered in the UI.

## GGWave node pairing

### Overview

BVR implements GGWave-based node pairing for secure, proximity-based device
discovery and authentication. This section describes the complete
implementation including the CGO adapter, audio transport layer, confirmation
UI, and production hardening.

### Implementation status

All four recommended follow-up items are complete:

1. **CGO Adapter** (`internal/ggwave/ggwave_cgo.go`) — Build-tagged optional
   CGO adapter for GGWave codec
2. **Platform Audio Transport** (`internal/ggwave/ggaudio.go`,
   `ggaudio_portaudio.go`) — Cross-platform audio I/O via PortAudio
3. **Explicit User Confirmation** (`internal/ui/dialog/ggwave_pairing.go`) —
   Interactive pairing confirmation dialog
4. **Network Transport Integration** (`internal/ui/model/ui.go`) — Pairing
   results register nodes for authenticated transport

### Architecture

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

### Build configuration

BVR ships with `CGO_ENABLED=0`, so the GGWave feature is behind a build tag.
The tag is **`ggwave && cgo`** on both the native codec and the PortAudio
backend, with `!ggwave || !cgo` stubs supplying the same API surface. Four
build combinations are valid:

| `CGO_ENABLED` | `-tags=ggwave` | Result |
| --- | --- | --- |
| `0` | no | Stubs (default release build) |
| `1` | no | Stubs |
| `1` | yes | Full native codec + PortAudio |
| `0` | yes | Stubs — the tag alone is not enough |

#### Default build (CGO-free)

```bash
go build .
```

The GGWave codec is NOT included. The `ggwave` package provides portable types
(`PairingEnvelope`, `Waveform`) and stub implementations that return clear
errors when CGO features are used.

Both stubs are required. Gating only the codec leaves `NewTransport` with no
`newCaptureSource`/`newPlaybackSink` and the package fails to compile, and
gating neither one makes `CGO_ENABLED=0` builds fail outright because PortAudio
has no CGO-free implementation.

#### Stub and native API parity

The stubs must expose exactly the signatures of the native build, otherwise
caller code compiles in one configuration and breaks in the other. `Encode` and
`Decode` operate on `[]int16`; the float32 variants are `EncodeFloat32` and
`DecodeFloat32`. An earlier stub declared `Encode`/`Decode` over `[]float32`,
which silently guaranteed a build break the moment the native backend was
compiled in.

#### CGO-enabled build

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

### Pairing flow

For proximity-based device discovery without network pre-configuration:

1. **Build with GGWave support** (one-time):
   `./scripts/build_ggwave.sh`, then `CGO_ENABLED=1 go build -tags=ggwave .`
2. **Start both devices** with BVR running
3. **Initiate pairing** from the home screen — the `GGWAVE NODE LINK` waveform
   transmits a `PairingEnvelope` containing node identity, endpoint, and token
4. **Confirm on receiving device** — a dialog shows decoded details (Node ID,
   Endpoint, Token, Expiry) with Accept/Reject buttons
5. **Node registers** — accepted pairings create a `Device` with connection
   type `ggwave` and the `Endpoint` as `TransportURL` for ongoing
   authenticated communication

The GGWave codec **and** the audio transport are both optional. They require
the `ggwave` build tag together with `CGO_ENABLED=1`; the tag alone is not
enough. The default binary keeps the portable pairing types, the pairing
confirmation dialog, and the node registration handler, but the codec and the
PortAudio backend both compile down to stubs that return a clear
"not compiled in" error.

### Core components

#### 1. Pairing envelope (`internal/ggwave/pairing.go`)

Portable, CGO-free data structure for node pairing:

```go
type PairingEnvelope struct {
    Version   int    `json:"v"`          // Protocol version (1)
    Kind      string `json:"kind"`       // "bvr-node-pair"
    NodeID    string `json:"node"`       // Unique node identifier
    Endpoint  string `json:"endpoint"`   // Network endpoint (e.g., "https://host:port")
    Token     string `json:"token"`      // Short-lived auth token
    ExpiresAt int64  `json:"expires_at"` // Unix timestamp
}
```

**Functions:**

- `EncodePairing(envelope)` — Serialize to JSON bytes
- `DecodePairing(payload)` — Validate and deserialize
- `CompactText(envelope)` — Base64url encoding for non-audio transport
- `ParseCompactText(text)` — Reverse of CompactText

#### 2. CGO adapter (`internal/ggwave/ggwave_cgo.go`)

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

#### 3. Audio transport (`internal/ggwave/ggaudio.go`)

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

#### 4. Pairing session (`internal/ggwave/ggaudio.go`)

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

#### 5. Confirmation dialog (`internal/ui/dialog/ggwave_pairing.go`)

Interactive dialog showing decoded envelope details:

```
🔗  GGWAVE NODE PAIRING REQUEST

Node ID:   node-abc123
Endpoint:  https://192.168.1.100:8080
Token:     short-lived-xyz
Expires:   15:04:05 (in 45s)

  [   ]  ACCEPT    [ ✓ ]  REJECT
```

**Keys:** `y`/`Y` = accept, `n`/`N` = reject, `←`/`→` = switch, `Enter` =
confirm, `Esc` = reject

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

#### 6. Node integration (`internal/ui/model/ui.go`)

When `PairingConfirmMsg.Accepted == true`:

1. Creates `node.Device` with connection type "ggwave"
2. Registers via `node.UpsertDevice()`
3. Uses `Endpoint` as `TransportURL` for authenticated transport
4. Shows info message: "Node paired via GGWave: {nodeID}"

### Device enumeration

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

### Metrics (`internal/ggwave/ggaudio_metrics.go`)

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

### Visual waveform (`internal/ggwave/waveform.go`)

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

### Production hardening features

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

### Testing

```bash
# Unit tests (no CGO)
go test ./internal/ggwave/...

# Integration tests (with CGO)
CGO_ENABLED=1 go test -tags=ggwave ./internal/ggwave/...

# All tests
go test ./...
```

### Security considerations

1. **Pairing envelope** contains only bootstrap credentials (node ID, endpoint,
   short-lived token)
2. **Token expiry** enforced — rejected if `ExpiresAt <= now`
3. **Explicit user confirmation** required before node registration
4. **Sound only for bootstrap** — ongoing communication uses authenticated
   network transport (HTTPS/WebSocket)
5. **No commands in envelope** — only identity and connection info

### Future work

- [ ] macOS CoreAudio implementation (`ggaudio_darwin.go`)
- [ ] Windows WASAPI implementation
- [ ] Linux ALSA/PulseAudio/PipeWire optimization
- [ ] Mobile client audio backends (iOS AudioToolbox, Android OpenSL ES)
- [ ] Encrypted pairing envelope (age/NaCl)
- [ ] Multi-device simultaneous pairing
- [ ] Pairing history and revocation

### Reference

- GGWave C library: https://github.com/ggerganov/ggwave
- PortAudio: http://www.portaudio.com/
- Protocol: Frequency-Shift Keying (FSK) with Reed-Solomon error correction
