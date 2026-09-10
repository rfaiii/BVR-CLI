# GGWave node pairing

BVR now has the portable first slice of its GGWave integration:

- `internal/ggwave` owns the short-lived, user-confirmed pairing envelope.
- The homescreen renders a live `GGWAVE NODE LINK` waveform driven by TUI animation ticks.
- Pairing data is limited to node identity, endpoint, token, and expiry; it does not carry commands.

The actual GGWave codec is not enabled in the default binary yet. GGWave is a native C/C++ library that generates and decodes audio waveforms, while BVR's release pipeline intentionally builds with `CGO_ENABLED=0`. The next adapter should therefore be optional and build-tagged, with audio capture/playback kept outside the TUI update loop.

Recommended follow-up:

1. Add a `ggwave`/CGO adapter that turns `PairingEnvelope` bytes into samples and decodes samples back into bytes.
2. Add platform audio capture/output implementations for macOS, Linux, Windows, and mobile clients.
3. Require explicit confirmation after decoding before registering a NODE.
4. Use a normal authenticated network transport for ongoing node control; use sound only for bootstrap/pairing.

Reference implementation: <https://github.com/ggerganov/ggwave>
