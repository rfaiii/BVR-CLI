# NODE WebSocket Protocol

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