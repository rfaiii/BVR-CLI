# NODE Transports

BVR supports one device registry with four transport adapters:

| Scheme | Adapter | Best for |
| --- | --- | --- |
| `http://`, `https://` | HTTP/JSON | Request/response execution and health checks |
| `ws://`, `wss://` | WebSocket/JSON | Persistent connections and streaming output |
| SSH configuration | SSH | Existing computers reachable through SSH |
| `ggwave://` | GGWave Audio | Proximity-based pairing and bootstrap |

## Quick start

```sh
bvr node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN"
```

The listener exposes `/v1/node/health`, `/v1/node/capabilities`,
`/v1/node/execute`, and the `/v1/node/ws` WebSocket endpoint. Register HTTP or
WebSocket endpoints with `RegisterNodeTransport`. SSH uses `NewSSHTransport` or
`RegisterSSHTransport`. GGWave pairing uses `OpenPairingConfirm` to bootstrap
node registration.

Credentials remain in transport objects and are not stored in `Device` or shown
in the UI. The home screen and NODE settings dialog show device name and state:
grey offline, red error, green online.

## GGWave Audio Pairing

For proximity-based device discovery without network pre-configuration:

1. **Build with GGWave support** (one-time):
   ```sh
   ./scripts/build_ggwave.sh
   CGO_ENABLED=1 go build -tags=ggwave .
   ```

2. **Start both devices** with BVR running

3. **Initiate pairing** from the home screen - the `GGWAVE NODE LINK` waveform
   transmits a `PairingEnvelope` containing node identity, endpoint, and token

4. **Confirm on receiving device** - a dialog shows decoded details (Node ID,
   Endpoint, Token, Expiry) with Accept/Reject buttons

5. **Node registers** - accepted pairings create a `Device` with connection
   type `ggwave` and the `Endpoint` as `TransportURL` for ongoing
   authenticated communication

The GGWave codec **and** the audio transport are both optional. They require
the `ggwave` build tag together with `CGO_ENABLED=1`; the tag alone is not
enough. The default binary keeps the portable pairing types, the pairing
confirmation dialog, and the node registration handler, but the codec and the
PortAudio backend both compile down to stubs that return a clear
"not compiled in" error.

Run `./scripts/build_ggwave.sh` once to fetch the native library, then build with
`CGO_ENABLED=1 go build -tags=ggwave .`. See `GGWAVE.md` for the full build
matrix and the stub/native API parity rules.
