# Pelton
An open-source email client built around your privacy, cross-platform with zero telemetry, fully yours.

## Features

- **Colored flags** with eight colors, kept local by default and optionally synced to the server as IMAP keywords so other clients see them.
- **In-app previewer** for PDFs, images, and text/code/markdown, so attachments open without leaving Pelton.
- **Save all** attachments from a message to a folder in one click, each with a progress bar.
- **Snooze** a message from the right-click menu with a friendly date-time picker; it resurfaces marked unread, and can optionally hide from the inbox until then.
- **Mark read/unread** and every other row action from the context menu.
- **Offline downloads**: pin a single message, or bulk-download every message since a chosen date (with a per-run choice to include attachments) for fast offline search, with a live progress bar and ETA in the status bar.
- **Downloaded indicator** on messages kept offline, which can be hidden.
- **Trackpad swipe gestures** on messages (configurable; left deletes, right marks unread by default).
- **Address book autocomplete** in the composer, learned from mail you send and receive, manageable in settings.
- **Vim mode** in the compose editor for modal editing.
- **Dynamic window title** that reflects the open message or current folder.
- **Customizable keyboard shortcuts**, including unbound-by-default keys for the right-click actions.

## Installation

Every [GitHub release](https://github.com/TRC-Loop/Pelton/releases) ships installers for macOS (Intel and Apple Silicon), Windows, and Fedora Linux. All builds are unsigned (no Apple notarization, no Windows code signing), so first launch needs one extra step on macOS and Windows - see below.

### macOS

1. Download `Pelton-<version>-macos-intel.dmg` (Intel Macs) or `Pelton-<version>-macos-applesilicon.dmg` (Apple Silicon / M-series Macs).
2. Open the `.dmg` and drag `Pelton.app` into `Applications`.
3. Since the build is unsigned, macOS Gatekeeper blocks the first launch with an "unidentified developer" warning. Right-click (or Control-click) `Pelton.app` in Applications and choose **Open**, then confirm in the dialog that appears. You only need to do this once; after that it opens normally, including from Spotlight or the Dock.

A `.zip` of the raw `.app` (`Pelton-<version>-macos-<arch>-app.zip`) is also attached to each release if you'd rather not use the `.dmg`.

### Windows

1. Download `Pelton-<version>-windows-amd64-installer.exe`.
2. Run it. Since the build is unsigned, Windows SmartScreen may show an "unrecognized app" warning the first time - click **More info** then **Run anyway**. This warning fades on its own as the file builds up download reputation.
3. The installer walks you through:
   - an installer-language picker (matches Pelton's own UI languages: English, German, French, Dutch, Spanish)
   - the GPL-3.0 license
   - **install for all users** (needs admin) or **just me** (no admin needed)
   - an optional desktop shortcut (Start Menu shortcut is always created)
   - a "Launch Pelton" checkbox on the last page

Windows on ARM is not built yet.

### Linux (Fedora)

Two ways to install, pick one:

**Via `dnf` (recommended - gets updates automatically once you upgrade):**

```sh
sudo dnf copr enable arnek/pelton
sudo dnf install pelton
```

(Fedora's Copr is a community repo host; enabling it adds Pelton's repo to `dnf` so future releases show up as normal updates.)

**Via the standalone `.rpm`:**

Download `Pelton-<version>-linux-fedora-x86_64.rpm` from the release and install it directly:

```sh
sudo dnf install ./Pelton-<version>-linux-fedora-x86_64.rpm
```

This won't auto-update; you'll need to download and install each new version's `.rpm` by hand.

### Other Linux distributions

Only a Fedora package is published right now. On other distributions, build from source (needs Go, Node/pnpm, and the Wails CLI - see `make run` / `make build-linux` in the `Makefile`):

```sh
git clone https://github.com/TRC-Loop/Pelton.git
cd Pelton
make build-linux
```

This produces a binary plus a `.desktop` launcher in `build/bin/`; copy the binary somewhere on your `PATH`, install the `.desktop` file to `~/.local/share/applications/`, and give it an icon named `pelton` (see `build/icons/`).
