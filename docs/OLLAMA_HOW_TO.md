# Ollama How-To

BVR can discover and use local Ollama models without manually maintaining a
model JSON entry.

## 1. Install Ollama

Install Ollama from [ollama.com](https://ollama.com), then verify the command.

### macOS

Install the Ollama macOS application. It normally starts the local service in
the background. If the command is not on `PATH`, reopen your terminal after
installing or add the Ollama CLI location to your shell `PATH`.

### Windows

Install the Ollama Windows application. It normally runs the service in the
background. Use PowerShell for the commands below; `where.exe ollama` checks
whether the CLI is available on `PATH`.

### Linux

Install the Linux package from Ollama's instructions. Start the service with
your system service manager when available, or run `ollama serve` in a terminal.
For a systemd installation, the service is commonly managed with:

```sh
systemctl --user enable --now ollama
```

Verify the command on every operating system:

```sh
ollama --version
```

## 2. Start Ollama

If the desktop/service installation is not already running, start the local
daemon manually:

```sh
ollama serve
```

The default API endpoint is:

```text
http://127.0.0.1:11434
```

If Ollama is already running as a desktop/background service, BVR detects it
The picker `s` action is a fallback and may report that Ollama is already
running on macOS or Windows. Selecting a model also warms it automatically.

## 3. Open the BVR model picker

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

## 4. Pull a model from BVR

Highlight a model in the Ollama section and press `p`. Pull progress appears in
the picker with status, percentage, and digest. Press `c` to cancel. When the
pull completes, BVR refreshes the local model list.

You can also pull models from a shell:

```sh
ollama pull qwen2.5:7b
```

## 5. Select and use a model

Press `enter` on an installed model. BVR creates or updates the managed
`ollama-local` provider, saves the selected model, refreshes the agent, and
warms the model in the background. The home-screen diamond turns green when it
is ready and red if loading fails.
After that, prompts, MCP tools, LSP context, and normal agent coordination use
the selected local model through Ollama's OpenAI-compatible endpoint.

## 5a. Use the native small-model coder

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

## 6. Environment and advanced configuration

Set `OLLAMA_HOST` when the daemon listens somewhere other than the default. The
syntax is the same across macOS, Linux, and PowerShell on Windows:

```sh
export OLLAMA_HOST=http://127.0.0.1:11434
```

PowerShell:

```powershell
$env:OLLAMA_HOST = "http://127.0.0.1:11434"
```

Manual provider configuration remains available for remote Ollama hosts, TLS,
proxies, custom headers, and other advanced setups. See
`docs/config/README.md`.

## Troubleshooting

### Ollama is not listed

Check that `ollama` is on `PATH`:

```sh
command -v ollama
```

### Ollama is installed but offline

Start the daemon and refresh the picker:

```sh
ollama serve
```

Then press `r` in the model picker.

### The model is missing

Press `p` on the model entry if it is shown, or pull it directly:

```sh
ollama pull qwen2.5:7b
```

### Context size looks wrong

BVR reads architecture-specific context metadata from Ollama's `/api/show`.
Refresh after changing a model or its runtime settings.

### Native coder cannot find a file

The native agent operates in its working directory. Confirm the file exists
there, or provide the project explicitly:

```sh
ls -l /path/to/project/note.txt
bvr-cli code --native --cwd /path/to/project \
  --model qwen2.5-coder:3b-instruct "Read note.txt"
```

### Native coder appears stuck loading

Run with diagnostics:

```sh
CODETOOL_DEBUG=1 bvr-cli code --native --stream \
  --model qwen2.5-coder:3b-instruct "Inspect this project"
```

Each Ollama turn is bounded and incomplete streams produce an error. Verify the
daemon directly with `curl http://127.0.0.1:11434/api/tags` and rebuild
`codetool` after updating the source.