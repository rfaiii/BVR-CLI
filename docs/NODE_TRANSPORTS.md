# NODE Transports

BVR supports one device registry with three transport adapters:

| Scheme | Adapter | Best for |
| --- | --- | --- |
| `http://`, `https://` | HTTP/JSON | Request/response execution and health checks |
| `ws://`, `wss://` | WebSocket/JSON | Persistent connections and streaming output |
| SSH configuration | SSH | Existing computers reachable through SSH |

## Quick start

```sh
bvr node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN"
```

The listener exposes `/v1/node/health`, `/v1/node/capabilities`,
`/v1/node/execute`, and the `/v1/node/ws` WebSocket endpoint. Register HTTP or
WebSocket endpoints with `RegisterNodeTransport`. SSH uses `NewSSHTransport` or
`RegisterSSHTransport`.

Credentials remain in transport objects and are not stored in `Device` or shown
in the UI. The home screen and NODE settings dialog show device name and state:
grey offline, red error, green online.

Open **NODE Connections** from the command palette (`/node`) or press
`ctrl+shift+n`. Use `r` in the dialog to refresh discovery. Configured
transport endpoints are shown for each device so you can verify which node is
selected before running Node/NPM commands.

## Pairing your iPhone

1. Start the agent on your laptop/desktop:
   ```sh
   bvr node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN"
   ```
2. Make sure your iPhone is on the same local network as the host.
3. Open the BVR companion app on iPhone.
4. Enter the host IP and port, then the same bearer token.
5. The device appears in **NODE Connections** as `online` when the health
   check succeeds.

If the device stays `offline`, verify the host firewall allows port `7777`
and that the token matches exactly.

## Commands from iPhone

When connected, prompts you send from the companion chat are forwarded to the
host and executed with the same workspace context. Output streams back to the
phone chat in real time over WebSocket, or as request/response results over
HTTP.

See `NODE_HTTP_PROTOCOL.md`, `NODE_WEBSOCKET_PROTOCOL.md`, and
`NODE_SSH_TRANSPORT.md` for protocol details.
