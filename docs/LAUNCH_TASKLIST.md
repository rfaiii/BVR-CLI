# Phase 4 Launch Tasklist

This document tracks our progress towards launching BVR-CLI, covering essential testing phases and marketplace distribution accounts.

## 1. Feature & Sound Verification
Before launching, run through this manual checklist on your primary machine (macOS/Windows) to ensure polish.

### Audio & UI Polish
- [x] **Startup Chime:** Verify the startup sound triggers when BVR-CLI launches.
- [ ] **Action Sounds:** Verify sounds for `Enter` key presses, success notifications, and error notifications.
- [ ] **Animations:** Check the beaver mascot pulse animation (4s) and hover debounce (400ms) on the homescreen.
- [ ] **Resource Bars:** Verify the CPU/RAM gradient bars animate smoothly and accurately reflect system load.

### Core Workflows
- [ ] **Authentication:** Test `/login` or `bvr auth` workflow and ensure the key is saved correctly.
- [ ] **Ollama Discovery:** Ensure `/models` correctly loads and lists available models from `localhost:11434`.
- [ ] **File Finder:** Test `Ctrl+Shift+F` across a large project. Verify previews, metadata, hidden files, and clipboard.
- [ ] **Project Switching:** Use `/cd` or `/project` to switch workspaces and confirm BVR updates the working directory.
- [ ] **NODE Transports:** Test connection via HTTP/JSON, WebSocket, and SSH.

## 2. Storefronts & Developer Accounts
To sell and distribute BVR-CLI effectively, sign up for or configure the following accounts:

### Distribution Platforms
- [ ] **Apple Developer Program:** Required to sign and notarize the macOS `.dmg` and `.pkg` installers so macOS Gatekeeper doesn't block them. ($99/year).
- [ ] **Microsoft Partner Center:** Required to sign the Windows `.exe` and distribute via the Microsoft Store or Winget without SmartScreen warnings.
- [ ] **Snapcraft Developer Account:** Register the `bvr-cli` name on the Snap store for easy Linux distribution.
- [ ] **Homebrew Tap:** Ensure `rfaiii/homebrew-bvr` (or similar) is ready and accessible.

### Payment & Sales Stores
- [ ] **Gumroad / Lemon Squeezy / Stripe:** Set up the main storefront to handle license key generation and secure payments.
- [ ] **Setapp / MacPaw:** Apply for Setapp distribution as an alternative subscription revenue stream for macOS users.

### Launch Day Platforms
- [ ] **Product Hunt:** Draft the upcoming product page, tag makers, and prep the launch day assets.
- [ ] **Hacker News (Show HN):** Draft the introductory post explaining the "keyboard-first AI workspace" pitch.
