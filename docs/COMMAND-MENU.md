# Command Menu — Architecture & Reference

This document catalogs every item in the BVR-CLI command palette and landing-screen
button stack, their keyboard shortcuts, the sound effect (SFX) assigned to each,
and the architecture flow from user input → action handler → dialog/UI mutation.

## Command Menu Sources

There are two entry points for commands:

1. **Landing Screen Buttons** — `internal/ui/model/landing.go` `landingView()`
   renders four clickable buttons below the CWD line and above the beaver hero.
2. **Command Palette** — `internal/ui/dialog/commands.go` `defaultCommands()` builds
   the list of system commands shown when `ctrl+p` (or the Command Palette button)
   is pressed. When a session is active, additional context-sensitive commands
   appear (e.g. Summarize Session, Toggle Pills).

```
                    ┌─────────────────────────────────────────┐
                    │         Landing Screen                  │
                    │                                         │
                    │  CWD line  ── ctrl+p ──→ Command Palette │
                    │  ┌──────────┐                          │
                    │  │OPEN CMD  │  click / ctrl+p           │
                    │  │OPEN DIR  │  click / ctrl+o            │
                    │  │CREATE    │  click / ctrl+n            │
                    │  │IMG RECOG │  click / ctrl+b            │
                    │  └──────────┘                          │
                    │        ┌───────┐                        │
                    │        │ BEAVER│  click → boop SFX      │
                    │        └───────┘                        │
                    │  MODEL / PROVIDER line                  │
                    │  GGWAVE waveform                         │
                    └─────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────────────────────────────┐
                    │  Command Palette (ctrl+p)               │
                    │  ┌────────────────────────────────────┐ │
                    │  │ System Commands                    │ │
                    │  │ User Commands (skills)             │ │
                    │  │ MCP Prompts                        │ │
                    │  └────────────────────────────────────┘ │
                    │  filter / fuzzy search                 │
                    │  shortcuts (ctrl+N, ctrl+L, …)         │
                    └─────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────────────────────────────┐
                    │  Dialog / Action Router                   │
                    │  internal/ui/model/ui.go                 │
                    │  case dialog.ActionXxx:                  │
                    │    → playAudio(...)                     │
                    │    → m.dialog.OpenDialog(...)           │
                    │    → state mutation / cmds              │
                    └─────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────────────────────────────┐
                    │  Sub-Dialogs                            │
                    │  • Models (ctrl+l)                      │
                    │  • Sessions (ctrl+s)                    │
                    │  • NODE Settings (ctrl+shift+n)         │
                    │  • Themes (ctrl+shift+t)                │
                    │  • Sounds                               │
                    │  • Quit (ctrl+c)                        │
                    │  • File Browser (ctrl+o)                │
                    │  • File Picker (ctrl+f)                 │
                    │  • Create File (ctrl+n)                 │
                    │  • Reasoning                           │
                    │  • Notifications                        │
                    │  • Other Models                         │
                    │  • Ollama How-To                        │
                    │  • Arguments (for custom commands)      │
                    │  • Permissions                          │
                    │  • OAuth (API key / device flow)        │
                    └─────────────────────────────────────────┘
```

## Landing Screen Buttons

| # | Button              | Shortcut  | SFX Type        | Action                                                      |
|---|---------------------|-----------|-----------------|-------------------------------------------------------------|
| 1 | OPEN COMMANDS       | `ctrl+p`  | `menu-open`     | Opens the command palette dialog (`openCommandsDialog`)     |
| 2 | OPEN FILE FINDER    | `ctrl+o`  | `menu-open`     | Opens the project file browser dialog (`openFileBrowserDialog`) |
| 3 | CREATE FILE         | `ctrl+n`  | `quick-notify-02` | Opens the file picker dialog (`NewCreateFile`)              |
| 4 | IMAGE RECOGNITION   | `ctrl+b`  | `quick-notify-03` | Opens file browser → selects image → runs `bvr lookup-image <path>` |
| 5 | BEAVER LOGO         | click     | `beaver`        | Triggers beaver boop animation + sound (`triggerBeaverBoop`)  |

## Command Palette — System Commands (in order)

> Rendered by `defaultCommands()` in `internal/ui/dialog/commands.go`.
> Conditional items are marked with ⚑ (only shown when their condition is met).

| #  | ID                    | Title                        | Shortcut      | Action Type            | SFX Type         |
|----|-----------------------|------------------------------|---------------|------------------------|------------------|
| 1  | `new_session`         | New Session                  | `ctrl+n`      | ActionNewSession       | `reload`         |
| 2  | `switch_session`      | Sessions                     | `ctrl+s`      | ActionOpenDialog       | `chat`           |
| 3  | `switch_model`        | Switch Model                 | `ctrl+l`      | ActionOpenDialog       | `menu-open`      |
| 4  | `file_browser`        | Open File Finder             | `ctrl+o`      | ActionOpenDialog       | `menu-open`      |
| 5  | `change_project`      | Change Project               | —             | ActionChangeProject    | `quick-notify-02`|
| 6  | `node_settings`       | NODE Connections             | `ctrl+shift+n`| ActionOpenDialog       | `menu-open`      |
| 7  | `ollama_models`       | Ollama Models                | —             | ActionOpenDialog       | `menu-open`      |
| 8  | `other_models`        | Other Models                 | —             | ActionOpenDialog       | `menu-open`      |
| 9  | `ollama_how_to`       | Ollama How To                | —             | ActionOpenDialog       | `menu-open`      |
| 10 | `themes`              | Themes                       | `ctrl+shift+t`| ActionOpenDialog       | `menu-open`      |
| 11 | `sounds`              | Sounds                       | —             | ActionOpenDialog       | `menu-open`      |
| 12 | `open_website`        | Open Website                 | —             | ActionOpenDialog (BrowserID) | `quick-notify-03` |
| 13 | `select_notifications`| Notification Style           | —             | ActionSelectNotificationStyle | `menu-close`  |
| 14 | `toggle_beastmode`    | Toggle Beast Mode            | `ctrl+y`      | ActionToggleBeastmodeMode | `chainsaw`     |
| 15 | `toggle_code_mode`    | Toggle Code Mode             | `ctrl+shift+c`| ActionToggleCodeMode   | `quick-notify-01`|
| 16 | `toggle_help`         | Toggle Help                  | `ctrl+g`      | ActionToggleHelp       | `quick-notify-02`|
| 17 | `init`                | Initialize Project           | —             | ActionInitializeProject | `quick-notify-01`|
| 18 | `toggle_transparent`  | Toggle Transparent BG        | —             | ActionToggleTransparentBackground | `quick-notify-01`|
| 19 | `quit`                | Quit                         | `ctrl+c`      | ActionQuit             | `exit`           |

### Conditional Commands (shown only when conditions are met)

| #  | ID                    | Title                        | Condition        | Action Type            | SFX Type         |
|----|-----------------------|------------------------------|-------------------|------------------------|------------------|
| 20 | `summarize` ⚑         | Summarize Session            | has active session | ActionSummarize       | `long-load`      |
| 21 | `toggle_thinking` ⚑   | Enable/Disable Thinking Mode | Anthropic model that supports thinking | ActionToggleThinking | `quick-notify-03` |
| 22 | `select_reasoning_effort` ⚑ | Select Reasoning Effort | OpenAI model with reasoning levels | ActionSelectReasoningEffort | `menu-open` |
| 23 | `toggle_sidebar` ⚑    | Toggle Sidebar               | active session + width ≥ 120 | ActionToggleCompactMode | `quick-notify-01` |
| 24 | `file_picker` ⚑       | Open File Picker             | active session + model supports images | ActionOpenDialog (FilePickerID) | `menu-open` |
| 25 | `open_external_editor` ⚑ | Open External Editor       | `$EDITOR` env var set | ActionExternalEditor  | `quick-notify-03` |
| 26 | `enable_docker_mcp` ⚑  | Enable Docker MCP Catalog    | Docker MCP available but not enabled | ActionEnableDockerMCP | `menu-open` |
| 27 | `disable_docker_mcp` ⚑ | Disable Docker MCP Catalog   | Docker MCP currently enabled | ActionDisableDockerMCP | `menu-close` |
| 28 | `toggle_pills` ⚑       | Toggle To-Dos/Queue          | has todos or queue items | ActionTogglePills   | `quick-notify-02` |

## Action Handler Flow

```
┌──────────────────────────────────────────────────────────────────────────┐
│  Command Palette selection (enter / click)                                │
│  dialog/commands.go: HandleMsg() → item.Action()                        │
└──────────────────────────────────────────────────────────────────────────┘
      │ returns Action (one of the types in dialog/actions.go)
      ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Dialog Action Router                                                    │
│  internal/ui/model/ui.go: handleDialogAction()                           │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 1. ActionOpenDialog │  →  openDialog(msg.DialogID)                   │
│  │    → plays SFX      │  →  switch case on DialogID                    │
│  │    → opens dialog   │     each case: playAudio() + open method      │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 2. ActionToggleXxx   │  →  toggle state field                       │
│  │    → plays SFX      │  →  playAudio()                                │
│  │    → closes cmds    │  →  CloseDialog(CommandsID)                    │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 3. ActionQuit      │  →  playAudio("exit")                          │
│  │                     │  →  tea.Quit                                   │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 4. ActionPermission │  →  switch msg.Action                         │
│  │    Response        │     Allow   → playAudio("quick-notify-01")      │
│  │                     │     Allow4S → playAudio("quick-notify-02")      │
│  │                     │     Deny    → playAudio("denied")               │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 5. ActionRunCustom  │  →  playAudio("quick-notify-01")               │
│  │    Command          │  →  sendMessage(content)                       │
│  │                     │  →  CloseFrontDialog                           │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 6. ActionAttachSkl  │  →  attachSkill(id, name)                      │
│  │                     │  →  playAudio("quick-notify-02")               │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 7. ActionRunMCPProm│  →  runMCPPrompt(clientID, promptID, args)      │
│  │    pt              │  →  playAudio("quick-notify-03")                │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 8. ActionFilePick  │  →  tea.Sequence(msg.Cmd(), ...)                │
│  │    erSelected       │  →  CloseDialog(FilePickerID)                  │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 9. ActionFileBrows │  →  if imageLookupPending:                      │
│  │    erSelected       │     runShellCommand("bvr lookup-image <path>")│
│  │                     │  →  else: attachFileFromPath(path)              │
│  │                     │  →  CloseDialog(FileBrowserID)                  │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 10. ActionChangePr │  →  exec.Command(bvr, --cwd, path)               │
│  │     oject           │  →  CloseDialog(FileBrowserID)                  │
│  └─────────────────────┘                                               │
│                                                                          │
│  ┌─────────────────────┐                                               │
│  │ 11. default         │  →  cmds = append(cmds, util.CmdHandler(msg))  │
│  │                     │     (passes through to Bubble Tea loop)        │
│  └─────────────────────┘                                               │
└──────────────────────────────────────────────────────────────────────────┘
      │
      ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Audio System                                                            │
│  internal/ui/audio/native.go                                             │
│                                                                          │
│  playAudio(title, message, audioType)                                    │
│    → getAudioFilename(audioType)                                         │
│    → maps type → .wav file in audio/sounds/                              │
│    → Native (afplay on macOS) / OSC (SSH) / Bell / Noop                  │
└──────────────────────────────────────────────────────────────────────────┘
```

## SFX Inventory (audio types used by menu items)

| Audio Type           | Sound File(s)                          | Used By                                      |
|----------------------|----------------------------------------|----------------------------------------------|
| `beaver`             | OH-BEAV-01..05                         | Beaver logo click                            |
| `menu-open`          | MENU-OPEN-FLAT.wav                     | Commands, Models, File Finder, Themes, Sounds, NODE, File Picker, Reasoning |
| `menu-close`         | MENU-CLOSE-FLAT.wav                    | Close commands dialog, Notification Style, Disable Docker MCP |
| `quick-notify-01`    | QUICK-NOTIFY-01-FLAT.wav               | File Finder (landing), Create File, Code Mode, Init, Transparent |
| `quick-notify-02`    | QUICK-NOTIFY-02-FLAT.wav               | Create File (palette), Help toggle, Pills, Permission (session allow) |
| `quick-notify-03`    | QUICK-NOTIFY-03-FLAT.wav               | Image Recognition, External Editor, Thinking toggle, MCP Prompt, Permission (session allow) |
| `quick-notify`       | QUICK-NOTIFY-01..03 (random)           | Volume Updated |
| `notification`       | QUICK-NOTIFY-01-FLAT.wav               | Volume Updated (notification type) |
| `chat`               | CHAT-01..04 (random)                   | Sessions dialog open, Session selected |
| `chat-open`          | CHAT-OPEN-FLAT.wav                     | (fallback for chat-related opens) |
| `chat-close`         | CHAT-CLOSE-FLAT.wav                     | Session selected |
| `reload`             | RELOAD-FLAT.wav                        | New Session |
| `long-load`          | LONG-LOAD-01/02 (alternating)          | Summarize Session |
| `chainsaw`           | CHAINSAW-01/02 (random)                | Toggle Beast Mode |
| `exit`               | exit-01/02/03 (random)                 | Quit, Open Quit Dialog |
| `error`              | ERROR-01..03 (random)                  | Various error states |
| `denied`             | DENIED-FLAT.wav                        | Permission denied |
| `connected`          | QUICK-NOTIFY-01-FLAT.wav                | Model selected |
| `sub`                | sub-01.wav                             | Theme selection |
| `incoming`           | INCOMING-FLAT.wav                      | Permission required, Question asked |
| `startup`            | STARTUP-SONG-FLAT.wav                  | Application launch |

## Data Flow Diagram

```
User Input (keyboard / mouse)
      │
      ├─── Mouse Click (landing screen) ───────────────────────────────┐
      │   landing.go: landingView() renders button rects               │
      │   ui.go: MouseMsg handler checks image.Pt.In(rect)             │
      │   → playAudio(...) + openDialog / openFileBrowserDialog / ...  │
      │                                                                 │
      └─── Key Press ───────────────────────────────────────────────┐
          keys.go: KeyMap defines all shortcuts                      │
          ui.go: handleKeyPressMsg()                                 │
            → handleGlobalKeys() (global shortcuts)               │
              → openCommandsDialog() (ctrl+p)                      │
              → openFileBrowserDialog() (ctrl+o)                   │
              → openFileBrowserDialog + imageLookupPending (ctrl+b) │
              → openModelsDialog() (ctrl+l)                        │
              → etc.                                               │
            → state-specific handlers (editor, chat, sidebar)      │
          │                                                        │
          ▼                                                        │
   Command Palette (commands.go) ←───────────────────────────────┘
   defaultCommands() returns []*CommandItem
   Each CommandItem has: id, title, shortcut, action
     │
     ├── Enter / Click → item.Action() returns dialog.Action
     │
     ▼
   dialog/action_router (ui.go: handleDialogAction)
     │
     ├── ActionOpenDialog → openDialog(id)
     │     → playAudio() + dialog-specific open method
     │
     ├── ActionToggleXxx → state mutation + playAudio()
     │
     ├── ActionQuit → playAudio("exit") + tea.Quit
     │
     ├── ActionPermissionResponse → switch + playAudio()
     │
     ├── ActionRunCustomCommand → sendMessage() + playAudio()
     │
     ├── ActionAttachSkill → attachSkill() + playAudio()
     │
     ├── ActionRunMCPPrompt → runMCPPrompt() + playAudio()
     │
     ├── ActionFilePickerSelected → tea.Sequence(...)
     │
     ├── ActionFileBrowserSelected → attach or lookup-image
     │
     ├── ActionChangeProject → exec process restart
     │
     └── default → util.CmdHandler(msg)
```

## Key Source Files

| File | Role |
|------|------|
| `internal/ui/model/landing.go` | Landing screen view with beaver hero + 4 button stack |
| `internal/ui/model/ui.go` | Main UI model: mouse handling, key handling, `handleDialogAction`, `openDialog`, `playAudio` |
| `internal/ui/model/keys.go` | KeyMap struct + DefaultKeyMap with all keyboard shortcuts |
| `internal/ui/dialog/commands.go` | Command palette: `defaultCommands()`, `CommandItem`, search/filter |
| `internal/ui/dialog/commands_item.go` | `CommandItem` type: rendering, aliases, info column |
| `internal/ui/dialog/actions.go` | All `Action` type definitions (ActionOpenDialog, ActionToggleXxx, etc.) |
| `internal/ui/audio/native.go` | `getAudioFilename()` — maps audio types to .wav files |
| `internal/ui/audio/native_darwin.go` | macOS `afplay` backend |
| `internal/ui/audio/native_linux.go` | Linux backend |
| `internal/ui/audio/native_windows.go` | Windows backend |
| `internal/ui/anim/beaver.go` | `BeaverFrame` (dense 13×5 mascot) + `LargeMascotFrame` (sidebar) |
| `internal/ui/anim/beaver.go` | `LargeMascotFrame` (8-row sidebar beaver) |
| `internal/ui/boot/boot.go` | `BeaverFramesDenseAlpha` + `BeaverFramesDenseBeta` (ASCII frames) |
| `internal/cmd/lookup_image.go` | `bvr lookup-image [path]` CLI command for image recognition |
