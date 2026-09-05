# Beta Testing Program

## Overview

This document outlines the BVR-CLI beta testing program, including invite management, distribution channels, and testing workflows.

## Beta Channels

### 1. Direct DMG Distribution (Primary)

**Platform:** macOS (Apple Silicon / Intel)
**Format:** Signed DMG with terminal-aware launcher
**Distribution:** Direct download / email attachment

```bash
# Build DMG for distribution
cd /Users/richavery/Projects/bvr-cli-main
./scripts/package/macos-dmg.sh v1.2.1-beta.1
```

**Installation:**
1. Download `bvr-cli_v1.2.1-beta.1_darwin_arm64.dmg`
2. Open DMG and drag BVR.app to Applications
3. First launch may require: Right-click → Open (to bypass Gatekeeper)
4. Launcher auto-detects terminal preference

### 2. Homebrew Tap (Secondary)

**Platform:** macOS / Linux
**Formula:** `packaging/homebrew-bvr-cli.rb`

```bash
# Install via Homebrew
brew install richavery/tap/bvr-cli

# Upgrade to beta
brew upgrade bvr-cli
```

### 3. NPM Package (Cross-platform)

**Platform:** macOS / Windows / Linux
**Package:** `packaging/npm-package.json`

```bash
# Install via NPM
npm install -g @bvr-cli/cli

# Upgrade to beta
npm update -g @bvr-cli/cli
```

### 4. Go Install (Developers)

**Platform:** Any with Go 1.23+
**Command:**

```bash
go install github.com/richavery/bvr-cli-main@latest
```

### 5. Windows (Future)

**Platform:** Windows 10/11
**Format:** EXE installer / MSI
**Status:** Planned for v1.2.1

---

## Beta Invite List

### Invite Management

Beta testers are tracked in this file. Each entry includes:
- Name / GitHub handle
- Platform (macOS/Windows/Linux)
- Terminal preference
- Testing focus area
- Status (pending/active/completed)

### Current Beta Testers

| Name | Platform | Terminal | Focus | Status |
|------|----------|----------|-------|--------|
| Richard Avery | macOS | Ghostty | Full suite | Active |
| [Add testers here] | | | | |

### Invite Workflow

1. **Add to list:** Update this file with new tester info
2. **Send invite:** Share DMG link or install instructions
3. **Track status:** Mark as pending → active → completed
4. **Collect feedback:** Use GitHub Issues or dedicated feedback form

---

## Testing Checklist

### Installation & Launch

- [ ] DMG opens without corruption
- [ ] Drag-and-drop to Applications works
- [ ] First launch bypasses Gatekeeper correctly
- [ ] Launcher opens correct terminal (Ghostty/Alacritty/etc.)
- [ ] BVR-CLI starts without errors
- [ ] `bvr-cli --version` returns correct version

### Onboarding Flow

- [ ] First-run onboarding appears automatically
- [ ] **OPT OUT** button works (skips tour)
- [ ] **CONTINUE** button advances through steps
- [ ] Each step displays correct content:
  - [ ] Welcome screen
  - [ ] Home view with ASCII preview
  - [ ] Models menu with ASCII preview
  - [ ] File Finder with ASCII preview
  - [ ] Notifications with ASCII preview
  - [ ] Themes with ASCII previews (pink/purple/green)
  - [ ] Complete screen
- [ ] Pressing Enter advances steps
- [ ] Pressing O/Esc opts out at any step
- [ ] Onboarding completes and shows main UI

### Web Browser

- [ ] `ctrl+b` opens the browser dialog
- [ ] Landing screen browser icon opens the browser
- [ ] File Finder `b` key opens the browser
- [ ] Typing a URL and pressing `Enter` fetches the article
- [ ] Loading spinner animates during fetch
- [ ] Article content renders as Markdown
- [ ] Error message displays for invalid URLs or non-HTML content
- [ ] `Esc` or `[X]` closes the browser
- [ ] Clicking outside the browser closes it
- [ ] Help text at bottom shows shortcuts

### Create File Dialog

- [ ] `ctrl+n` opens the New File dialog
- [ ] Landing screen paper icon opens the New File dialog
- [ ] File navigation works (keyboard and mouse)
- [ ] `[X]` closes the dialog
- [ ] `Esc` closes the dialog
- [ ] Clicking outside the dialog closes it

### Terminal Compatibility

Test on each terminal:
- [ ] Ghostty
- [ ] Alacritty
- [ ] Kitty
- [ ] WezTerm
- [ ] iTerm2
- [ ] Terminal.app

Verify:
- [ ] ASCII art renders correctly
- [ ] Colors display properly
- [ ] Layout is not broken
- [ ] No rendering glitches

### Core Features

- [ ] Home screen loads
- [ ] Model selection works
- [ ] File finder opens and browses files
- [ ] Notifications appear
- [ ] Theme switching works
- [ ] Bottom resource bar shows CPU/RAM
- [ ] Keyboard shortcuts function

### Performance

- [ ] App launches in < 3 seconds
- [ ] No memory leaks after 10 min usage
- [ ] CPU usage is reasonable (< 5% idle)
- [ ] No crashes during normal usage

---

## Feedback Collection

### GitHub Issues

**Template:**
```
**Version:** v1.2.1-beta.X
**Platform:** macOS/Windows/Linux
**Terminal:** Ghostty/Alacritty/etc.
**Go Version:** (if applicable)

**Issue:**
[Description]

**Steps to reproduce:**
1.
2.
3.

**Expected:**
[What should happen]

**Actual:**
[What actually happened]

**Screenshots:**
[If applicable]
```

### Feedback Form (Optional)

For testers without GitHub, use this form:
- Email: averydevz@outlook.com
- Subject: `[BVR Beta] Feedback - v1.2.1-beta.X`

---

## Bug Severity Levels

### Critical (P0)
- App crashes on launch
- Data loss
- Security vulnerability
- **Action:** Fix within 24 hours, hotfix release

### High (P1)
- Core feature broken
- Workaround exists but is painful
- **Action:** Fix within 3 days, patch release

### Medium (P2)
- Minor feature broken
- Cosmetic issues
- **Action:** Fix in next regular release

### Low (P3)
- Nice-to-have improvements
- Documentation fixes
- **Action:** Backlog for future release

---

## Release Process

### Beta Phase

1. **Build beta DMG:**
   ```bash
   ./scripts/package/macos-dmg.sh v1.2.1-beta.1
   ```

2. **Upload to distribution:**
   - GitHub Releases (draft)
   - Google Drive / Dropbox (if needed)
   - Email to beta list

3. **Send invites:**
   - Include install instructions
   - Link to this document
   - Request feedback within 7 days

4. **Monitor feedback:**
   - Check GitHub Issues daily
   - Triage bugs by severity
   - Prioritize fixes

### Release Candidate (RC)

After 2 weeks of beta testing:

1. **Fix all P0/P1 bugs**
2. **Build RC:**
   ```bash
   ./scripts/package/macos-dmg.sh v1.2.1-rc.1
   ```
3. **Send to full beta list**
4. **Wait 3 days for showstoppers**

### General Availability (GA)

If no showstoppers:

1. **Tag release:**
   ```bash
   git tag v1.2.1
   git push origin v1.2.1
   ```

2. **Build final DMG:**
   ```bash
   ./scripts/package/macos-dmg.sh v1.2.1
   ```

3. **Publish:**
   - GitHub Release
   - Homebrew tap
   - NPM package
   - Update website/docs

---

## Communication

### Beta Mailing List

Maintain a list of beta testers with:
- Email address
- GitHub handle (if applicable)
- Platform/terminal
- Testing focus

### Status Updates

Weekly email to beta list:
- What was fixed
- What's being worked on
- Next steps

### Changelog

Maintain `CHANGELOG.md` with:
- Beta-specific changes
- Known issues
- Upgrade instructions

---

## Next Steps

1. **Finalize DMG packaging** for beta distribution
2. **Create beta invite list** with 5-10 initial testers
3. **Set up feedback channel** (GitHub Issues or form)
4. **Build beta DMG** and test locally
5. **Send invites** and begin beta phase
6. **Monitor feedback** for 2 weeks
7. **Release RC** if stable
8. **GA release** after RC validation

---

## Resources

- **Beta Issues:** https://github.com/richavery/bvr-cli-main/issues
- **Email:** averydevz@outlook.com
- **Discord/Slack:** [Add if applicable]
- **Documentation:** See `docs/` directory
