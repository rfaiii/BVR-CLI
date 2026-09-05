# BVR installation and testing

BVR is a terminal-first AI workspace. This guide covers macOS Apple Silicon,
macOS Intel, Windows x64, and Linux x64/arm64.

## Downloaded release archives

1. Download the archive matching your operating system and CPU.
2. Extract it.
3. Put `bvr-cli` somewhere on your `PATH`.
4. Open a new terminal and verify:

```sh
bvr-cli --version
bvr-cli --help
```

### macOS

Use `darwin_arm64` for Apple Silicon and `darwin_x86_64` for Intel Macs:

```sh
chmod +x bvr-cli
sudo install -m 0755 bvr-cli /usr/local/bin/bvr-cli
```

If macOS blocks an unsigned beta binary, approve BVR in **System Settings →
Privacy & Security** and run it again.

### Linux

Use `linux_x86_64` for Intel/AMD 64-bit or `linux_arm64` for ARM64:

```sh
chmod +x bvr-cli
sudo install -m 0755 bvr-cli /usr/local/bin/bvr-cli
```

Without administrator access:

```sh
mkdir -p ~/.local/bin
install -m 0755 bvr-cli ~/.local/bin/bvr-cli
```

Ensure `~/.local/bin` is on `PATH`.

### Windows x64

Use the `windows_x86_64` ZIP archive. Extract `bvr-cli.exe` to a folder such
as `%LOCALAPPDATA%\BVR\bin`, add that folder to the user `PATH`, and open a
new PowerShell window:

```powershell
bvr-cli.exe --version
bvr-cli.exe --help
```

## Build from source

Requirements: Go 1.26 or newer.

```sh
cd /path/to/bvr-cli-go
go build -o ./bvr-cli .
bvr-cli --version
```

Always use `bvr-cli` when testing a repository build. A stale copy elsewhere on
`PATH` can report an older version even when the repo binary is newer.

## Choose a project

Start BVR in a project:

```sh
bvr-cli --cwd /path/to/project
```

Inside the app, use `/cd` or `/project`, then press `s` on a directory to switch.
You can also type `cd ~/Projects/my-project`; BVR restarts there without
sending the command to the AI agent.

## Ollama setup

Install Ollama from [ollama.com](https://ollama.com). BVR discovers every model
reported by Ollama’s local API in `/models` and `/ollama`. Selecting a model
automatically starts and warms it when possible.

```sh
ollama --version
ollama ls
```

The default endpoint is `http://127.0.0.1:11434`. Set `OLLAMA_HOST` for another
endpoint. See [`OLLAMA_HOW_TO.md`](OLLAMA_HOW_TO.md).

## Native small-model coder

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

## Agent Skills

BVR discovers skills from `~/.agents/skills`. Sync the master catalog with:

```sh
./scripts/install-master-skills.sh
```

See [`SKILLS.md`](SKILLS.md).

## Configuration and data

Use `bvr-cli dirs` to inspect platform-specific data locations. `bvr-cli
projects` lists recently tracked projects.

## Dependencies

### Required
- BVR binary — no separate runtime needed after build.
- Terminal emulator — Ghostty, Terminal.app, Windows Terminal, Alacritty, iTerm2, Kitty, or WezTerm.
- Network access — required for provider auth, model discovery, and updates.

### Optional
- Ollama — local models; see `OLLAMA_HOW_TO.md`.
- Provider API key — Anthropic, OpenAI, Gemini, etc.
- Git — project detection and workspace state when available.
- LSP servers — optional per language.

### Build requirements
- Go 1.26+
- Git
- Optional: `goreleaser`, `nsis`/`wixtoolset`, Xcode Command Line Tools

## Troubleshooting

- **The app looks unchanged:** rebuild and run `bvr-cli`, not an older copy found
  on `PATH`.
- **Ollama models are missing:** start Ollama, open `/models`, and press `r`.
- **The model is slow initially:** the first request loads model weights; BVR
  warms the selected model after selection.
- **macOS security warning:** approve the binary in Privacy & Security.
- **Windows command not found:** reopen PowerShell after changing `PATH`.

## Verification

```sh
bvr-cli --version
bvr-cli --help
bvr-cli dirs
bvr-cli projects
```

Then launch BVR, switch to a test project, open `/models`, refresh local
models, select one, and confirm the model indicator becomes ready.
