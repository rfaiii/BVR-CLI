# NODE SSH Transport

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