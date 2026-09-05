# NODE HTTP Protocol

The NODE agent exposes a small authenticated JSON API for device execution.
The same `bvr node serve` listener also exposes the WebSocket streaming
endpoint documented in `NODE_WEBSOCKET_PROTOCOL.md`.

## Start an agent

```sh
bvr node serve --host 0.0.0.0:7777 --token "$BVR_NODE_TOKEN"
```

Tokens are required when binding outside localhost. Keep the agent behind a
trusted network or TLS/reverse proxy; the current agent provides bearer-token
authentication but does not terminate TLS itself.

## Endpoints

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