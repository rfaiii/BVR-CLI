# Audio Resources for BVR-CLI

This directory contains audio files for the BVR-CLI audio/notification system.

## Audio File Structure

### Sound Files
- **startup-song-01.wav**: Musical jingle for application startup
- **advert-01.wav**: Used for future advertisement space or smoke break
- **broken-01.wav**: Used to indicate that model is broken or too jank to work
- **notification-01.wav**: Short alert sound for general notifications variation 1
- **notification-02.wav**: Short alert sound for general notifications variation 2
- **notification-03.wav**: Short alert sound for general notifications variation 3
- **connected-01.wav**: Indication for successful connection attempt
- **drill-01.wav**: Used for annoying your users variation 1
- **drill-02.wav**: Used for annoying your users variation 2
- **chat-01.wav**: Used for incoming chat from models variation 1
- **chat-02.wav**: Used for incoming chat from models variation 2
- **chat-03.wav**: Used for incoming chat from models variation 3
- **error-01.wav**: Distinctive sound for error conditions
- **error-02.wav**: Distinctive sound for error conditions
- **error-03.wav**: Distinctive sound for error conditions
- **exit-01.wav**: Confirmation sound for application termination version 1
 **exit-02.wav**: Confirmation sound for application termination version 2
  **exit-03.wav**: Confirmation sound for application termination version 3
- **chainsaw-01.wav**: Unique beaver chainsaw effect for git operations version 1
- **chainsaw-02.wav**: Unique beaver chainsaw effect for git operations version 2
- **chainsaw-03.wav**: Unique beaver chainsaw effect for git operations version 3
- **incoming-01.wav**: Unique way to show up to a party or chat
**loading-01.wav**: Medium alert sound for slowed operations variation 1
**loading-02.wav**: Medium alert sound for slowed operations variation 2
- **oops-01.wav**: Distinctive sound for crapping your pants when trying to fart too hard
- **question-01.wav**: Did someone say something? I couldn't hear you very well... variation 1
- **question-02.wav**: Did someone say something? I couldn't hear you very well... variation 2
- **reload-01.wav**: Did someone spray something? I could smell you from very far away...
- **sub-01.wav**: When you need more bass than treble... variation 1
- **sub-02.wav**: When you need more bass than treble... variation 2
- **sub-03.wav**: When you need more bass than treble... variation 3


## Technical Specifications

### Format Requirements
- **Primary Format**: WAV (uncompressed, high quality)
- **Sample Rate**: 44.1kHz
- **Bit Depth**: 16-bit
- **Channels**: Mono for notifications, stereo for effects (COMING SOON)
- **File Size**: Under 500KB per file 
- **Duration**: 0.5-5 seconds based on usage 

### Audio Properties
| Sound File | Duration | Size | Sample Rate | Bit Depth | Channels |
|------------|----------|------|-------------|-----------|----------|
| notification-01.wav | 0.8s | 245KB | 44.1kHz | 16-bit | Mono |
| startup-song-01.wav | 3.2s | 512KB | 44.1kHz | 16-bit | Stereo |
| error-01.wav | 0.6s | 198KB | 44.1kHz | 16-bit | Mono |
| question-01.wav | 1.2s | 312KB | 44.1kHz | 16-bit | Mono |
| chainsaw-01.wav | 2.3s | 412KB | 44.1kHz | 16-bit | Stereo |

## Usage Instructions

### Audio System Integration
1. Place audio files in this directory
2. Reference using relative paths
3. Audio files are automatically loaded by the audio notification system
4. System falls back to terminal bell if audio playback fails

### File Permissions
- All audio files should be readable by the application
- No executable permissions required
- Maintain proper directory structure for compatibility

## Audio Source Information

### Sound Acquisition
- **Notification Sound**: Source from Freesound.org (CC BY 3.0)
- **Startup Sound**: Source from Pixabay (Free for commercial use)
- **Error Sound**: Source from CC Mixter (Creative Commons)
- **Exit Sound**: Source from Pixabay (Free for commercial use)
- **Chainsaw Sound**: Custom production for unique BVR branding

### Sound Licensing
- All sounds are licensed for commercial use
- No attribution required for usage in BVR-CLI
- Files are self-contained with licensing information

## Testing and Validation

### Audio Testing
1. Test on all target platforms (macOS, Windows, Linux)
2. Verify audio playback compatibility
3. Test fallback mechanisms when audio unavailable
4. Performance testing for resource usage

### Quality Assurance
1. Verify all audio files play without errors
2. Test audio synchronization with notifications
3. Validate audio levels are consistent
4. Ensure no audio artifacts or distortions

## File Organization

### Current Structure
```
audio/
├── sounds/           # Audio files
│   ├── advert-01.wav
│   ├── broken-01.wav
│   ├── chainsaw-01.wav
│   ├── chainsaw-02.wav
│   ├── chainsaw-03.wav
│   ├── chat-01.wav
│   ├── chat-02.wav
│   ├── chat-03.wav
│   ├── connected-01.wav
│   ├── drill-01.wav
│   ├── drill-02.wav
│   ├── error-01.wav
│   ├── error-02.wav
│   ├── error-03.wav
│   ├── exit-01.wav
│   ├── exit-02.wav
│   ├── exit-03.wav
│   ├── incoming-01.wav
│   ├── loading-01.wav
│   ├── loading-02.wav
│   ├── notification-01.wav
│   ├── notification-02.wav
│   ├── notification-03.wav
│   ├── oops-01.wav
│   ├── question-01.wav
│   ├── question-02.wav
│   ├── reload-01.wav
│   ├── startup-song-01.wav
│   ├── sub-01.wav
│   ├── sub-02.wav
│   └── sub-03.wav
├── metadata/         # Sound specifications
│   └── sound-index.json
└── formats/          # Format documentation
    └── audio-formats.md
```

### Sound Index
```json
{
  "sounds": [
    {
      "name": "notification-01",
      "file": "notification-01.wav",
      "duration": 0.8,
      "size": 245024,
      "format": "WAV",
      "sampleRate": 44100,
      "bitDepth": 16,
      "channels": "mono",
      "description": "Short notification alert sound",
      "usage": "General notifications and system events"
    },
    {
      "name": "startup-song-01",
      "file": "startup-song-01.wav",
      "duration": 3.2,
      "size": 524288,
      "format": "WAV",
      "sampleRate": 44100,
      "bitDepth": 16,
      "channels": "stereo",
      "description": "Musical jingle for application startup",
      "usage": "Application startup confirmation"
    },
    {
      "name": "error-01",
      "file": "error-01.wav",
      "duration": 0.6,
      "size": 202432,
      "format": "WAV",
      "sampleRate": 44100,
      "bitDepth": 16,
      "channels": "mono",
      "description": "Distinctive alert sound for error conditions",
      "usage": "Error notifications and failure confirmations"
    },
    {
      "name": "question-01",
      "file": "question-01.wav",
      "duration": 1.2,
      "size": 312000,
      "format": "WAV",
      "sampleRate": 44100,
      "bitDepth": 16,
      "channels": "mono",
      "description": "Sound for when a model disconnects or a question mark is typed",
      "usage": "Model disconnections or chat questions"
    },
    {
      "name": "chainsaw-01",
      "file": "chainsaw-01.wav",
      "duration": 2.3,
      "size": 422112,
      "format": "WAV",
      "sampleRate": 44100,
      "bitDepth": 16,
      "channels": "stereo",
      "description": "Unique beaver chainsaw effect for git operations",
      "usage": "Git uploads, pushes, and file synchronization events"
    }
  ]
}
```