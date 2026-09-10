# BVR-CLI Audio Resources

BVR embeds WAV cues for startup, interaction feedback, model activity, errors,
permissions, and remote-workspace events. Audio playback is asynchronous so a
sound never blocks the Bubble Tea update loop.

## Sound files

Processed cues are 16-bit, 44.1 kHz WAV files. The filenames are intentionally
semantic: code refers to an event name and the audio backend resolves it to the
bundled filename.

| File | Placement |
| --- | --- |
| `STARTUP-SONG-FLAT.wav` | Application startup. |
| `INTRO-BOP-01-FLAT.wav` | Reserved for the animated-logo intro sequence. |
| `BROKEN-FLAT.wav` | Failed shell command or unknown/janky failure. |
| `QUICK-NOTIFY-01-FLAT.wav` | Prompt/command submission, generic confirmation, and File Finder open. |
| `QUICK-NOTIFY-02-FLAT.wav` | Help toggle, lighter UI confirmations, and Create File open. |
| `QUICK-NOTIFY-03-FLAT.wav` | Tab/focus changes and Web Browser open. |
| `CONNECTION-ISSUE-FLAT.wav` | MCP/model disconnects and Ollama load failures. |
| `MENU-OPEN-FLAT.wav` | Opening menus, dialogs, Finder, models, themes, NODE, and file picker. |
| `MENU-CLOSE-FLAT.wav` | Closing the command menu. |
| `CHAT-01-FLAT.wav` … `CHAT-04-FLAT.wav` | Randomized assistant/model response cues. |
| `CHAT-OPEN-FLAT.wav` / `CHAT-CLOSE-FLAT.wav` | Opening sessions and selecting a session. |
| `ERROR-01-FLAT.wav` … `ERROR-03-FLAT.wav` | Randomized application/agent error cues. |
| `CHAINSAW-01-FLAT.wav` / `CHAINSAW-02-FLAT.wav` | Reserved for high-energy sync or git actions. |
| `INCOMING-FLAT.wav` | Pending permission, question, or other request requiring attention. |
| `LONG-LOAD-01-FLAT.wav` / `LONG-LOAD-02-FLAT.wav` | Ollama/model loading and longer asynchronous work. |
| `LONG-UPLOADING-FLAT.wav` | Reserved for future upload/sync progress. |
| `DENIED-FLAT.wav` | Permission denial. |
| `WHAT-FLAT.wav` | User question/clarification cue. |
| `WHAT-FLAT.wav` + `OH-BEAV-01.wav` … `OH-BEAV-05.wav` | Randomized Beaver logo interaction pool. Clicking the animated landing or chat/sidebar Beaver chooses one of all six cues. |
| `RELOAD-FLAT.wav` | New session and reload/reset actions. |
## Current creative placements

These are deliberately playful additions to the original event list:

- Prompt submission gets a short confirmation cue.
- Tab and focus changes get a light navigation cue.
- Permission and question requests get an incoming cue.
- Denied tools get a distinct denial cue.
- Failed shell commands get the broken cue.
- Assistant responses randomly choose one of four chat cues.
- Opening and closing menus use separate cues.
- The homescreen File Finder, Create File, and Web Browser actions each have
  their own quick-notify cue so the four launch buttons do not sound identical.
- Clicking the animated Beaver logo on the homescreen or in the chat sidebar
  randomly selects one of six Beaver cues, including `WHAT-FLAT.wav`.

## Legacy / not-yet-processed cues

These remain embedded and available but are not part of the processed set:

- `exit-01.wav`, `exit-02.wav`, `exit-03.wav`: exit confirmation variations.
- `advert-01.wav`: future AFK/advertising cue.
- `sub-01.wav`, `sub-02.wav`: future low-frequency transition cues.
- `who-01.wav`: reserved for a future identity/attention cue.

## Technical notes

- Native playback uses `afplay` on macOS, `paplay`/`aplay` on Linux, and
  PowerShell WPF `MediaPlayer` on Windows when available.
- The audio package extracts embedded files to a temporary directory on first
  use and cleans them up at process exit.
- `25`, `50`, `75`, `100`, and `silent` volume modes are controlled from the
  Sounds menu or `options.audio_volume`. New installs default to 25% to avoid
  surprising headphone users; the older `high` and `low` values remain
  compatible as aliases for 100% and 50%.
- Long cues should not be assigned to high-frequency key events.
- All files must remain readable by the application; executable permissions are
  unnecessary.

## Validation

Before release, test native playback on macOS, Windows, and Linux; verify the
silent mode; test missing-player fallback behavior; and confirm that sounds do
not overlap excessively during rapid keyboard use.
