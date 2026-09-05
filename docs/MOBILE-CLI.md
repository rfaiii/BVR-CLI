# Mobile CLI

Plans and outline for the iPhone and Android companion variants, plus the
host-side companion bridge that connects them to `bvr-cli`.

## Goals

- Provide a mobile-first control surface for BVR-CLI sessions.
- Keep the existing Go CLI intact; add a lightweight companion bridge instead of rewriting the core in another runtime.
- Support secure local/network transport between the host CLI and mobile clients.

## Source of Truth

This plan is based on:
`/Users/richavery/.hermes/attachments/bvr_cli_companion_architecture.html`

Companion server design details live in:
[`docs/BVR-SERVER.md`](BVR-SERVER.md)

## Topology

- Host machine runs the normal BVR-CLI.
- Optional companion server/bridge runs alongside the CLI to expose secure mobile-facing commands.
- Mobile clients connect via a private/link-local or authenticated channel.

## Essentials

1. Host-side bridge module
2. Mobile client variant(s)
3. Secure auth/transport model
4. Session and telemetry bridge
5. Installation and docs

## Proposed Structure

```
mobile-cli/
  README.md
  proto.md
  ios/
  android/
```

## Roadmap

1. Capture requirements and constraints.
2. Define the bridge protocol and data model.
3. Prototype the host-side entrypoint.
4. Prototype one mobile client variant (iOS companion).
5. Build a web-based HTML/CSS prototype for rapid UI iteration without Xcode.
6. Migrate desktop settings and commands into the mobile surface.
7. Add tests, docs, and packaging.

## Current status

- **iOS Companion App** scaffolded in `mobile-cli/ios/` with chat, command palette, theme switching, MCP client, and web preview surfaces.
- **Web UI Prototype** built in `mobile-cli/web-preview/` for instant zero-compile UI testing.
- **FastMCP Server** implemented in `mcp-server/`, exposing tools over SSE.
- **Networking** The iOS app connects to the local FastMCP/companion bridge via `URLSession` for tool discovery and JSON-RPC requests.
- Companion server design is documented in `docs/BVR-SERVER.md`.

## Current limitations

- Mobile companion is in **test mode**.
- Full iPhone ↔ laptop NODE sync flow is not yet implemented end to end.
- Android client is not yet scaffolded.

## Next Steps

- Refine command palette parity: `/themes`, `/node`, `/finder`, `/mcp`, `/cd`.
- Implement the primary host-side companion bridge in Go.
- Lock down the minimum viable command set.
- Add onboarding breadcrumbs so users can quickly connect their iPhone to their desktop node.
