<p align="center">
  <img src="https://raw.githubusercontent.com/TRC-Loop/Pelton/13f56136136bc00b9c8721dc2042fc9c84e1b3a7/.github/pelton-large-bg.png" alt="Pelton Banner">
</p>

<p align="center">
  <a href="https://github.com/TRC-Loop/Pelton/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/TRC-Loop/Pelton?style=for-the-badge&color=blue" alt="License: GPL-3.0">
  </a>
  <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go" alt="Written in Go">
  <a href="https://arne.sh/discord">
    <img src="https://img.shields.io/badge/Discord-Join_Community-7289DA?style=for-the-badge&logo=discord&logoColor=white" alt="Discord">
  </a>
  <a href="https://github.com/TRC-Loop/Pelton/issues">
    <img src="https://img.shields.io/github/issues/TRC-Loop/Pelton?style=for-the-badge" alt="Issues">
  </a>
  <a href="https://github.com/TRC-Loop/Pelton/pulls">
    <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=for-the-badge" alt="PRs Welcome">
  </a>
</p>

<h3 align="center">An open-source email client built around your privacy, cross-platform with zero telemetry, fully yours.</h3>

***

## <img src="https://api.iconify.design/tabler/info-circle.svg?color=white" width="26" style="vertical-align: -4px;"> About Pelton

Pelton is a modern, Free and Open-Source Software (FOSS) email client written in Go and built using the Wails framework. It is designed from the ground up to respect your data. We believe your inbox belongs to you, which is why Pelton operates with absolute zero telemetry and full privacy. 

## <img src="https://api.iconify.design/tabler/sparkles.svg?color=white" width="26" style="vertical-align: -4px;"> Features

* <img src="https://api.iconify.design/tabler/shield-lock.svg?color=white" width="18" style="vertical-align: -4px;"> **Full Privacy:** Your data stays on your machine. Zero tracking, zero telemetry, and complete control over your inbox.
* <img src="https://api.iconify.design/tabler/bolt.svg?color=white" width="18" style="vertical-align: -4px;"> **Fast Search:** Find what you need instantly. The search engine is optimized for speed and handles large local mailboxes with ease.
* <img src="https://api.iconify.design/tabler/palette.svg?color=white" width="18" style="vertical-align: -4px;"> **Highly Customizable:** Tailor the client to fit your exact workflow and aesthetic preferences.
* <img src="https://api.iconify.design/tabler/file-export.svg?color=white" width="18" style="vertical-align: -4px;"> **Portable Configuration:** Export your entire setup, including accounts, preferences, and custom layouts, into a single easily transferable file.
* <img src="https://api.iconify.design/tabler/code.svg?color=white" width="18" style="vertical-align: -4px;"> **FOSS & Cross-Platform:** Truly open source and built to run beautifully across different operating systems.
* <img src="https://api.iconify.design/tabler/flag.svg?color=white" width="18" style="vertical-align: -4px;"> **Colored flags** with eight colors, kept local by default and optionally synced to the server as IMAP keywords so other clients see them.
* <img src="https://api.iconify.design/tabler/eye.svg?color=white" width="18" style="vertical-align: -4px;"> **In-app previewer** for PDFs, images, and text/code/markdown, so attachments open without leaving Pelton.
* <img src="https://api.iconify.design/tabler/download.svg?color=white" width="18" style="vertical-align: -4px;"> **Save all** attachments from a message to a folder in one click, each with a progress bar.
* <img src="https://api.iconify.design/tabler/clock.svg?color=white" width="18" style="vertical-align: -4px;"> **Snooze** a message from the right-click menu with a friendly date-time picker; it resurfaces marked unread, and can optionally hide from the inbox until then.
* <img src="https://api.iconify.design/tabler/cloud-off.svg?color=white" width="18" style="vertical-align: -4px;"> **Offline downloads**: pin a single message, or bulk-download every message since a chosen date (with a per-run choice to include attachments) for fast offline search, with a live progress bar and ETA in the status bar.
* <img src="https://api.iconify.design/tabler/hand-finger.svg?color=white" width="18" style="vertical-align: -4px;"> **Trackpad swipe gestures** on messages (configurable; left deletes, right marks unread by default).
* <img src="https://api.iconify.design/tabler/address-book.svg?color=white" width="18" style="vertical-align: -4px;"> **Address book autocomplete** in the composer, learned from mail you send and receive, manageable in settings.
* <img src="https://api.iconify.design/tabler/keyboard.svg?color=white" width="18" style="vertical-align: -4px;"> **Vim mode** in the compose editor for modal editing, plus **customizable keyboard shortcuts** throughout.

## <img src="https://api.iconify.design/tabler/help-circle.svg?color=white" width="26" style="vertical-align: -4px;"> Frequently Asked Questions

<details>
<summary><b>Which email providers does Pelton support?</b></summary>
<p>Pelton supports standard IMAP and SMTP, which means it works with the vast majority of email providers out of the box. We also have built-in OAuth2 support for Gmail. Because Pelton is FOSS, the community can easily contribute to add custom support for even more providers.</p>
</details>

<details>
<summary><b>Does Pelton work offline?</b></summary>
<p>Yes. You can configure Pelton to cache your newest emails locally (you choose the timeframe), and you can explicitly mark specific emails to be kept offline permanently.</p>
</details>

<details>
<summary><b>Does Pelton support PGP/GPG encryption?</b></summary>
<p>Yes, of course!</p>
</details>

<details>
<summary><b>Where is my data actually stored?</b></summary>
<p>Your data stays entirely in your control. It is stored in a SQLite database on your local machine, and on your original email provider's server. We do not host or route your data through any third-party servers.</p>
</details>

<details>
<summary><b>Why use Wails and Go instead of Electron?</b></summary>
<p>Speed and memory efficiency. By leveraging Go and Wails, Pelton uses significantly less RAM compared to heavy Electron wrappers, while still giving you a snappy, cross-platform UI.</p>
</details>

<details>
<summary><b>How does Pelton handle custom HTML tracking pixels or remote images?</b></summary>
<p>This is entirely configurable. By default, remote images and tracking pixels are blocked (similarly to Thunderbird). A small banner will appear letting you know images were blocked, and you can choose to allow them for that email if you wish.</p>
</details>

<details>
<summary><b>Can I self-host or use my own custom database path for sync?</b></summary>
<p>Using a custom database path (like pointing SQLite to a network share) is not yet recommended or fully tested, but this functionality will be coming soon.</p>
</details>

<details>
<summary><b>Is there any telemetry or automated crash reporting?</b></summary>
<p>No. Pelton has absolutely zero telemetry. If you experience a crash or bug, please help us out by manually opening an issue on GitHub.</p>
</details>

<br>

## <img src="https://api.iconify.design/tabler/photo.svg?color=white" width="26" style="vertical-align: -4px;"> Gallery

*(Screenshots and UI previews will be placed here soon!)*

## <img src="https://api.iconify.design/tabler/download.svg?color=white" width="26" style="vertical-align: -4px;"> Installation

Every [GitHub release](https://github.com/TRC-Loop/Pelton/releases) ships installers for macOS (Apple Silicon), Windows, and Fedora Linux. All builds are unsigned (no Apple notarization, no Windows code signing), so first launch needs one extra step on macOS and Windows - see below.

### macOS

1. Download `Pelton-<version>-macos-applesilicon.dmg` (Apple Silicon / M-series Macs).
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
sudo dnf copr enable arnek/Pelton
sudo dnf install pelton
```

(Fedora's Copr is a community repo host; enabling it adds Pelton's repo to `dnf` so future releases show up as normal updates.)

> [!NOTE]  
> **Fedora <=42 is not supported**, even if you pin the chroot.
> 
> You will get an error like package not available.
> 
> Update your System folks! (You have to do it anyway so ig)

**Via the standalone `.rpm`:**

Download `Pelton-<version>-linux-fedora-x86_64.rpm` from the release and install it directly:

```sh
sudo dnf install ./Pelton-<version>-linux-fedora-x86_64.rpm
```

This won't auto-update; you'll need to download and install each new version's `.rpm` by hand.

To uninstall either way, run `sudo dnf remove pelton` (or use the Remove button in GNOME Software / KDE Discover).

### Other Linux distributions

Only a Fedora package is published right now. On other distributions, build from source (needs Go, Node/pnpm, and the Wails CLI - see `make run` / `make build-linux` in the `Makefile`):

```sh
git clone https://github.com/TRC-Loop/Pelton.git
cd Pelton
make build-linux
```

This produces a binary plus a `.desktop` launcher in `build/bin/`; copy the binary somewhere on your `PATH`, install the `.desktop` file to `~/.local/share/applications/`, and give it an icon named `pelton` (see `build/icons/`).

## <img src="https://api.iconify.design/tabler/messages.svg?color=white" width="26" style="vertical-align: -4px;"> Contact & Community

* **Discord:** Join the discussion at [arne.sh/discord](https://arne.sh/discord)
* **Email:** Reach out directly via [pelton@arne.sh](mailto:pelton@arne.sh)

## <img src="https://api.iconify.design/tabler/users.svg?color=white" width="26" style="vertical-align: -4px;"> Contributing

Contributions are welcome. Whether you are fixing bugs, refining the UI layout, polishing backend Go code, or enhancing documentation, feel free to open an issue or submit a Pull Request. See [AUTHORS.md](AUTHORS.md) if you'd like to be credited by name.

## <img src="https://api.iconify.design/tabler/file-certificate.svg?color=white" width="26" style="vertical-align: -4px;"> License

Pelton is distributed under the **[GPL-3.0 License](https://github.com/TRC-Loop/Pelton/blob/main/LICENSE)**. See `LICENSE` for details.
