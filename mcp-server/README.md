# BVR-CLI FastMCP Server

This directory contains the Python FastMCP server which acts as the centralized engine (Provider) for the BVR-CLI ecosystem, managing tools like `sync_gemini_command`. 

## Architecture

Based on our `rules.md`, this server is designed to act as a **local-first, modular provider**.
- It leverages the [`FastMCP`](https://github.com/jlowin/fastmcp) library to rapidly build Model Context Protocol endpoints in Python.
- It is managed by `uv`, an ultra-fast Python package installer and resolver.
- It exposes tools to both the Go CLI and the mobile companion apps.

## Network Transport (SSE)

To support local network testing with the iOS Simulator and the HTML Web Prototype, this server runs using **Server-Sent Events (SSE)** transport instead of standard `stdio`.

**Host:** `0.0.0.0`
**Port:** `8000`

Mobile clients connect to `http://localhost:8000/sse` to negotiate the JSON-RPC handshakes and fetch tool lists.

## Available Tools

- **`sync_gemini_command`**: A mock function currently designed to accept string commands and pass them to the Google GenAI SDK. This acts as the bridge for Gemini Spark integrations.

## How to Run

Ensure you have `uv` installed, then run the server:

```bash
cd mcp-server
uv run mcp-server
```

*(Press `Ctrl+C` to stop the server).*
