# BVR-CLI Audio System

The audio directory contains BVR's embedded WAV resources, metadata, and
event-to-sound documentation. Playback is asynchronous, so feedback never
blocks the Bubble Tea update loop.

## Directory layout

```text
audio/
├── sounds/                  # Embedded WAV files
│   ├── *-FLAT.wav           # Current processed production cues
│   └── legacy cues          # Retained compatibility sounds
├── metadata/
│   └── sound-index.json     # Complete inventory and WAV metadata
└── formats/                 # Format notes and future processing guidance
```

## Current system

- Processed cues use 16-bit, 44.1 kHz WAV audio.
- macOS uses `afplay`; Linux uses `paplay`/`aplay`; Windows uses PowerShell
  `SoundPlayer`.
- The backend maps semantic event names such as `menu-open`, `chat`, and
  `quick-notify-03` to bundled filenames.
- Audio volume is controlled through the Sounds menu and
  `options.audio_volume`.

See [`sounds/README.md`](sounds/README.md) for the placement map and
[`metadata/sound-index.json`](metadata/sound-index.json) for the authoritative
asset inventory.
