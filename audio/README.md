# BVR-CLI Audio System

This directory contains the audio resources and configuration for the BVR-CLI audio/notification system.

## Overview

The audio system provides comprehensive sound support for BVR-CLI notifications, including:
- General notification alerts
- Application startup jingles
- Error condition notifications
- Application exit confirmations
- Special effects for git operations (chainsaw sound)

## Directory Structure

```
audio/
├── sounds/           # Audio files
│   ├── notification.wav
│   ├── startup.wav
│   ├── error.wav
│   ├── exit.wav
│   └── chainsaw.wav
├── metadata/         # Sound specifications and metadata
│   └── sound-index.json
└── formats/          # Audio format documentation
    └── audio-formats.md
```