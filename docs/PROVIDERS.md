# BVR-CLI Providers

This document is the consolidated provider reference: local Ollama models and
the hosted Cline gateway. It merges the former `OLLAMA_HOW_TO.md` and
`CLINE_PROVIDER.md`.

## Ollama (local models)

BVR can discover and use local Ollama models without manually maintaining a
model JSON entry.

### 1. Install Ollama

Install Ollama from [ollama.com](https://ollama.com), then verify the command.

**macOS** — Install the Ollama macOS application. It normally starts the local
service in the background. If the command is not on `PATH`, reopen your
terminal after installing or add the Ollama CLI location to your shell `PATH`.

**Windows** — Install the Ollama Windows application. It normally runs the
service in the background. Use PowerShell for the commands below;
`where.exe ollama` checks whether the CLI is available on `PATH`.

**Linux** — Install the Linux package from Ollama's instructions. Start the
service with your system service manager when available, or run `ollama serve`
in a terminal. For a systemd installation (Arch, Omarchy, and other systemd
systems), the service is commonly managed with:

```sh
systemctl --user enable --now ollama
```

Verify the command on every operating system:

```sh
ollama --version
```

### 2. Start Ollama

If the desktop/service installation is not already running, start the local
daemon manually:

```sh
ollama serve
```

The default API endpoint is:

```text
http://127.0.0.1:11434
```

If Ollama is already running as a desktop/background service, BVR detects it.
The picker `s` action is a fallback and may report that Ollama is already
running on macOS or Windows. Selecting a model also warms it automatically.

### 3. Open the BVR model picker

Use `ctrl+l` or open the command palette and choose **Switch Model**. BVR
checks Ollama asynchronously and lists installed models under **Ollama
(Local)**.

Available model-picker keys:

| Key | Action |
| --- | --- |
| `enter` | Select the highlighted model |
| `r` | Refresh local runtimes and models |
| `p` | Pull the highlighted Ollama model |
| `s` | Start Ollama if it is offline |
| `c` | Cancel an active model pull |
| `esc` | Close the picker |

### 4. Pull a model from BVR

Highlight a model in the Ollama section and press `p`. Pull progress appears in
the picker with status, percentage, and digest. Press `c` to cancel. When the
pull completes, BVR refreshes the local model list.

You can also pull models from a shell:

```sh
ollama pull qwen2.5:7b
```

### 5. Select and use a model

Press `enter` on an installed model. BVR creates or updates the managed
`ollama-local` provider, saves the selected model, refreshes the agent, and
warms the model in the background. The home-screen diamond turns green when it
is ready and red if loading fails.
After that, prompts, MCP tools, LSP context, and normal agent coordination use
the selected local model through Ollama's OpenAI-compatible endpoint.

### 6. Use the native small-model coder

For Qwen coder models that are slow or unreliable with the full BVR tool
surface, use the fantasy-free native path:

```sh
GOWORK=off go build -o "$HOME/bin/codetool" ./cmd/codetool
GOWORK=off go build -o "$HOME/bin/bvr-cli" .

BVR_CODETOOL="$HOME/bin/codetool" \
  "$HOME/bin/bvr-cli" code --native \
  --model qwen2.5-coder:3b-instruct --stream \
  "Read note.txt and tell me exactly what it contains."
```

The native path uses the current directory. Use `--cwd /path/to/project` when
the files are elsewhere. `BVR_CODETOOL` may be omitted when `codetool` is on
`PATH`. Streaming is on by default; raw JSON tool-call dumps are buffered and
hidden after recovery, while normal assistant text remains live.

### 7. Environment and advanced configuration

Set `OLLAMA_HOST` when the daemon listens somewhere other than the default.
The syntax is the same across macOS, Linux, and PowerShell on Windows:

```sh
export OLLAMA_HOST=http://127.0.0.1:11434
```

PowerShell:

```powershell
$env:OLLAMA_HOST = "http://127.0.0.1:11434"
```

Manual provider configuration remains available for remote Ollama hosts, TLS,
proxies, custom headers, and other advanced setups. See
[`config/README.md`](config/README.md).

### Ollama troubleshooting

**Ollama is not listed** — check that `ollama` is on `PATH`:

```sh
command -v ollama
```

**Ollama is installed but offline** — start the daemon and refresh the picker:

```sh
ollama serve
```

Then press `r` in the model picker.

**The model is missing** — press `p` on the model entry if it is shown, or
pull it directly:

```sh
ollama pull qwen2.5:7b
```

**Context size looks wrong** — BVR reads architecture-specific context
metadata from Ollama's `/api/show`. Refresh after changing a model or its
runtime settings.

**Native coder cannot find a file** — the native agent operates in its working
directory. Confirm the file exists there, or provide the project explicitly:

```sh
ls -l /path/to/project/note.txt
bvr-cli code --native --cwd /path/to/project \
  --model qwen2.5-coder:3b-instruct "Read note.txt"
```

**Native coder appears stuck loading** — run with diagnostics:

```sh
CODETOOL_DEBUG=1 bvr-cli code --native --stream \
  --model qwen2.5-coder:3b-instruct "Inspect this project"
```

Each Ollama turn is bounded and incomplete streams produce an error. Verify the
daemon directly with `curl http://127.0.0.1:11434/api/tags` and rebuild
`codetool` after updating the source.

> Note: the in-app **Ollama How To** dialog renders an embedded copy of this
> guide at `internal/ui/dialog/ollama_how_to.md`. Update both when changing
> instructions.

## Cline provider (api.cline.bot)

Cline's hosted API gateway is integrated into BVR-CLI as a first-class
provider. It is OpenAI-compatible and fronts a curated set of frontier models
(Claude, GPT, Gemini, Qwen) using a Cline account token.

### Setup

1. Create a Cline API token at https://account.cline.bot (or via the Cline
   extension's account page).
2. Either export it:

   ```sh
   export CLINE_API_KEY=your-token-here
   ```

   or run `bvr-cli`, open the model picker, choose the **Cline** provider,
   and paste the token when prompted (stored under
   `providers.cline.api_key` in the global config).

3. Select a model from the Cline group (e.g. `anthropic/claude-sonnet-4.5`).

### Usage notes

- The Cline catalog is reachable two ways:
  1. The `/` command palette → **Other Models** (aliases: `other`, `cline`,
     `external`). This opens the "Other Models" dialog
     (`internal/ui/dialog/other_models.go`) listing the Cline models; picking
     one prompts for the API key when none is stored.
  2. The main model picker (`ctrl+l`) when the provider is configured (i.e. a
     key was set via env or config).
- **"ADD CLINE API KEY" flow:** when the Cline provider is not yet configured,
  the Other Models dialog shows an "ADD CLINE API KEY" row. Selecting it opens
  the API key input; once the key is saved (verified with the `X-Api-Key`
  header), bvr fetches the full live catalog from the gateway — including all
  of Cline's **free models** — and shows it without a restart.
- Selecting a Cline model sets the large model; the small model defaults to
  the provider's `DefaultSmallModelID`.

### Details

- Base URL: `https://api.cline.bot/v1`
- Auth header: `X-Api-Key: <token>` (set automatically; the gateway does not
  use the standard `Authorization: Bearer` scheme)
- Provider type: `openai-compat`
- Default models: defined in `internal/config/cline.go` (`ClineModels()`).
  Add or override models per-user via the `providers.cline.models` config
  entry; user models are merged with the defaults.
- Endpoint override: set `providers.cline.base_url` in config if you need to
  proxy the gateway (honored by the known-provider config merge).

### Implementation notes

- `internal/config/cline.go` — provider + default model definitions.
- `internal/config/provider.go` — Cline is appended to the known provider
  list returned by `Providers()`.
- `internal/config/load.go` — `configureProviders` case resolves the API key
  (`$CLINE_API_KEY` or stored config) and attaches the `X-Api-Key` header;
  the header flows into the agent coordinator's OpenAI-compatible provider
  build (`internal/agent/coordinator.go`).

### Validation

`go build ./...`, `go vet ./internal/config/`, and
`go test ./internal/config/` pass. Live API validation requires a Cline token.
