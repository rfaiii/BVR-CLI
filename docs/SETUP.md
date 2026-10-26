# BVR-CLI Setup and Installation

This document is the consolidated setup guide for BVR-CLI: installing
released binaries, building from source, platform requirements, dependencies,
and troubleshooting. It merges the former `installation.md` and
`ONBOARDING.md`. For first-run provider and model configuration, see
[`PROVIDERS.md`](PROVIDERS.md); for in-app controls, see
[`UI.md`](UI.md).

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

Use `darwin_arm64` for Apple Silicon and `darwin_amd64`/`darwin_x86_64`
for Intel Macs:

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

### Arch Linux / Manjaro / Omarchy

There is **no AUR package yet**. The release pipeline that would publish one
has not run for this repository, so `yay -S bvr-cli` and `yay -S bvr-cli-bin`
both fail with "package not found" until that packaging lands. Install the
built `.pkg.tar.zst` instead:

```sh
./scripts/package/build-installers.sh 1.2.8
sudo pacman -U ./dist/installers/bvr-cli-1.2.8-1-x86_64.pkg.tar.zst
```

Use the `aarch64` file on Arm machines. Manjaro and Omarchy use the same
package. After install:

```sh
bvr-cli --version
bvr-cli --help
```

`goreleaser release --snapshot` also produces Arch artifacts, but
`.goreleaser.yml` requires a GoReleaser **PRO** license key and Charm's
runner group, so it fails in this repository; use the script above.

### Windows x64

Use the `windows_amd64` ZIP archive on standard Intel/AMD PCs, or
`windows_arm64` on Windows ARM. Extract `bvr-cli.exe` to a folder such
as `%LOCALAPPDATA%\BVR\bin`, add that folder to the user `PATH`, and open a
new PowerShell window:

```powershell
bvr-cli.exe --version
bvr-cli.exe --help
```

For local beta package generation across all six targets:

```sh
./scripts/package/build-beta-packages.sh 1.2.8
```

## Build from source

Requirements: Go 1.27.3 or newer, matching the `go` directive in `go.mod`.

```sh
git clone https://github.com/rfaiii/BVR-CLI.git
cd BVR-CLI
go build -o ./bvr-cli .
./bvr-cli --version
```

Or install the local build:

```sh
./scripts/install-bvr-cli.sh
bvr-cli --version
```

Always use `bvr-cli` when testing a repository build. A stale copy elsewhere on
`PATH` can report an older version even when the repo binary is newer.

### Go workspace note

This repo may expect a shared `go.work` that includes sibling Charm module
paths. If `go build ./...` fails from this repo alone, use one of:

- `go build .` from `BVR-CLI` only
- Add this repo to the shared workspace
- Replace missing `charm.land/*` modules with local paths in `go.mod`

BVR depends on Charm ecosystem libraries:
`charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`,
`charm.land/glamour/v2`, `charm.land/catwalk`, `charm.land/fantasy`,
`charm.land/fang/v2`, `charm.land/log/v2`, `charm.land/x/vcr`,
`github.com/charmbracelet/ultraviolet`, and `github.com/charmbracelet/x/*`.

## Choose a project

Start BVR in a project:

```sh
bvr-cli --cwd /path/to/project
```

Inside the app, use `/cd` or `/project`, then press `s` on a directory to
switch. You can also type `cd ~/Projects/my-project`; BVR restarts there
without sending the command to the AI agent.

## Platform requirements

### macOS

- macOS 13.0+ for DMG builds. The locally produced DMGs are unsigned.
- Rosetta 2 on Apple Silicon if using Intel-only terminal plugins.
- Internet connection for first-run model/provider setup.
- Developer/build: Go 1.27.3+, Xcode Command Line Tools
  (`xcode-select --install`), optional `goreleaser` for release builds.

### Arch Linux / Linux

- Any modern terminal (Ghostty, Alacritty, Kitty, WezTerm, foot).
- Arch-based systems: the AUR packages above, or install the `nfpms`
  deb/rpm/apk artifacts on other distributions.
- Optional system service for Ollama (`systemctl --user enable --now ollama`).
- Developer/build: Go 1.27.3+, Git, optional `goreleaser`.

### Windows

- Windows 10/11 x64 or ARM64.
- Windows Terminal or another modern terminal host.
- Internet connection for first-run setup.
- Developer/build: Go 1.27.3+; MinGW/MSYS2 or Visual Studio Build Tools only
  if CGO dependencies require it; optional `goreleaser`, `nsis` or
  `wixtoolset` for installer packaging.

## Dependencies

### Required

| Dependency | Purpose | Notes |
| --- | --- | --- |
| BVR binary | Terminal UI runtime | No Go runtime required after build |
| Terminal emulator | Bubble Tea rendering surface | Ghostty, Terminal.app, Windows Terminal, Alacritty, iTerm2, Kitty, WezTerm |
| Network access | Provider auth, model discovery, updates | Required during first run and `/login` |

### Optional

| Dependency | Purpose | Auto-install | Notes |
| --- | --- | --- | --- |
| Ollama | Local models | Guided | Recommended for offline-capable setups; see [`PROVIDERS.md`](PROVIDERS.md) |
| Provider API key | Cloud providers | Guided | Anthropic, OpenAI, Gemini, etc. |
| Git | Project detection and workspace state | No | Used when available |
| LSP servers | Code intelligence | No | Optional per language |

### Build requirements

- Go 1.27.3+
- Git
- Optional: `goreleaser`, `nsis`/`wixtoolset`, Xcode Command Line Tools

## Ollama quick setup

Install Ollama from [ollama.com](https://ollama.com). BVR discovers every
model reported by Ollama's local API in `/models` and `/ollama`. Selecting a
model automatically starts and warms it when possible.

```sh
ollama --version
ollama ls
```

The default endpoint is `http://127.0.0.1:11434`. Set `OLLAMA_HOST` for another
endpoint. The full guide — pull, start, warm-up, native small-model coder, and
troubleshooting — lives in [`PROVIDERS.md`](PROVIDERS.md).

## Provider setup

BVR supports multiple providers. Required items per provider:

| Provider | Required | Auto-detect |
| --- | --- | --- |
| Ollama | `ollama` binary + pulled model | Yes |
| Anthropic | API key | No |
| OpenAI | API key | No |
| Gemini | API key or OAuth | No |
| Cline | API token | No (see [`PROVIDERS.md`](PROVIDERS.md)) |
| Custom/local | Endpoint + key/model | No |

Use `/login` or the models dialog (`ctrl+l`) to complete setup.

## Agent Skills

BVR discovers skills from `~/.agents/skills`. Sync the master catalog with:

```sh
./scripts/install-master-skills.sh
```

See [`SKILLS.md`](SKILLS.md).

## Configuration and data

Use `bvr-cli dirs` to inspect platform-specific data locations.
`bvr-cli projects` lists recently tracked projects.

Configuration lives in `bvrrc` (Bash builtins) with legacy `bvr.json`
support. Full reference: [`config/README.md`](config/README.md).

## Troubleshooting

- **The app looks unchanged:** rebuild and run `bvr-cli`, not an older copy
  found on `PATH`.
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
