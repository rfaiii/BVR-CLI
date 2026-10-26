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
└── metadata/
    └── sound-index.json     # Complete inventory and WAV metadata
```

## Current system

- Cues are mono, 8-bit unsigned PCM at 11.025 kHz. That is the smallest shape
  every player on every target accepts without a codec, and UI blips do not
  need the headroom 44.1 kHz stereo carries. The whole embedded bundle is
  845 KB, down from 6.9 MB, which is 6.2 MB off the released binary.
- Re-export with `scripts/compress_sounds.sh`, which prefers `ffmpeg` and falls
  back to `afconvert`, and strips the metadata chunks editors leave behind.
- macOS uses `afplay`; Linux uses `paplay`/`aplay`; Windows uses PowerShell
  WPF `MediaPlayer` when available.
- The backend maps semantic event names such as `menu-open`, `chat`, and
  `quick-notify-03` to bundled filenames.
- `audio/embed.go` extracts only the file being played, on first use, so
  startup never pays for a cue that never sounds.
- The `beaver` event randomly selects `WHAT-FLAT.wav` or one of the five
  `OH-BEAV-##.wav` cues for interactive Beaver logos.
- Audio volume starts at a conservative 25% by default. The Sounds menu and
  `options.audio_volume` support `25`, `50`, `75`, `100`, and `silent`.

See [`sounds/README.md`](sounds/README.md) for the placement map and
[`metadata/sound-index.json`](metadata/sound-index.json) for the authoritative
asset inventory.
