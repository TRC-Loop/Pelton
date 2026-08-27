<p align="center">
  <img src="https://raw.githubusercontent.com/TRC-Loop/Pelton/13f56136136bc00b9c8721dc2042fc9c84e1b3a7/.github/pelton-large-bg.png" alt="Pelton Banner">
</p>


<p align="center">
  <a href="https://github.com/TRC-Loop/Pelton/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/TRC-Loop/Pelton?style=for-the-badge&color=blue" alt="License: GPL-3.0">
  </a>
  <a href="#-versioning">
    <img src="https://img.shields.io/badge/CalVer-YYYY.Q.INCR-22bfda?style=for-the-badge" alt="CalVer YYYY.Q.INCR">
  </a>
  <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go" alt="Written in Go">
  <a href="https://discord.gg/UzPNGZYy6V">
    <img src="https://img.shields.io/badge/Discord-Join_Community-7289DA?style=for-the-badge&logo=discord&logoColor=white" alt="Discord">
  </a>
  <a href="https://github.com/TRC-Loop/Pelton/issues">
    <img src="https://img.shields.io/github/issues/TRC-Loop/Pelton?style=for-the-badge" alt="Issues">
  </a>
  <a href="https://github.com/TRC-Loop/Pelton/pulls">
    <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=for-the-badge" alt="PRs Welcome">
  </a>
  <a href="https://github.com/TRC-Loop/Pelton/pulls">
    <img src="https://img.shields.io/github/downloads/TRC-Loop/Pelton/total?style=for-the-badge" alt="PRs Welcome">
  </a>
  <a href="https://github.com/TRC-Loop/Pelton/pulls">
    <img src="https://img.shields.io/github/v/release/TRC-Loop/Pelton?style=for-the-badge&label=Version" alt="PRs Welcome">
  </a>
</p>

<h3 align="center">An open-source email client built around your privacy, cross-platform with zero telemetry, fully yours.</h3>

<h3 align="center">Install Guides</h3>

<div align="center">

| <img src="https://api.iconify.design/tabler/brand-apple.svg?color=white" width="18" style="vertical-align: -4px;"> [macOS](#-macos) | <img src="https://api.iconify.design/tabler/brand-windows.svg?color=white" width="18" style="vertical-align: -4px;"> [Windows](#-windows) | <img src="https://api.iconify.design/tabler/terminal-2.svg?color=white" width="18" style="vertical-align: -4px;"> [Linux](#-linux) |
|:---:|:---:|:---:|

</div>

<p align="center">
  <a href="https://discord.gg/UzPNGZYy6V">
    <img src="https://discord.com/api/guilds/1535688892689162260/widget.png?style=banner2" alt="Discord Banner">
  </a>
</p>

***

## <img src="https://api.iconify.design/tabler/info-circle.svg?color=white" width="26" style="vertical-align: -4px;"> About Pelton

Pelton is a modern, Free and Open-Source Software (FOSS) email client written in Go and built using the Wails framework. It is designed from the ground up to respect your data. We believe your inbox belongs to you, which is why Pelton operates with absolute zero telemetry and full privacy. 

## <img src="https://api.iconify.design/tabler/sparkles.svg?color=white" width="26" style="vertical-align: -4px;"> Features

* <img src="https://api.iconify.design/tabler/shield-lock.svg?color=white" width="18" style="vertical-align: -4px;"> **Full Privacy:** Your data stays on your machine. Zero tracking, zero telemetry, and complete control over your inbox.
* <img src="https://api.iconify.design/tabler/bolt.svg?color=white" width="18" style="vertical-align: -4px;"> **Fast Search:** Find what you need instantly. The search engine is optimized for speed and handles large local mailboxes with ease.
* <img src="https://api.iconify.design/tabler/palette.svg?color=white" width="18" style="vertical-align: -4px;"> **Highly Customizable:** Tailor the client to fit your exact workflow and aesthetic preferences. See https://themes.pelton.app
* <img src="https://api.iconify.design/tabler/file-export.svg?color=white" width="18" style="vertical-align: -4px;"> **Portable Configuration:** Export your entire setup, including accounts, preferences, and custom layouts, into a single easily transferable file.
* <img src="https://api.iconify.design/tabler/code.svg?color=white" width="18" style="vertical-align: -4px;"> **FOSS & Cross-Platform:** Truly open source and built to run beautifully across different operating systems.
* <img src="https://api.iconify.design/tabler/flag.svg?color=white" width="18" style="vertical-align: -4px;"> **Colored flags** with eight colors, kept local by default and optionally synced to the server as IMAP keywords so other clients see them.
* <img src="https://api.iconify.design/tabler/eye.svg?color=white" width="18" style="vertical-align: -4px;"> **In-app previewer** for PDFs, images, and text/code/markdown, so attachments open without leaving Pelton.
* <img src="https://api.iconify.design/tabler/clock.svg?color=white" width="18" style="vertical-align: -4px;"> **Snooze** a message from the right-click menu with a friendly date-time picker; it resurfaces marked unread, and can optionally hide from the inbox until then.
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

<img src="docs/assets/screenshots/screenshot-inbox-dark.png" alt="The unified inbox in dark mode, with a message open in the reading pane" width="100%">

<table>
<tr>
<td width="50%"><img src="docs/assets/screenshots/screenshot-inbox-light.png" alt="The same unified inbox in light mode, showing per-account mailboxes and saved views"></td>
<td width="50%"><img src="docs/assets/screenshots/screenshot-message-actions.png" alt="The message context menu with reply, colour tags, snooze and move actions"></td>
</tr>
<tr>
<td colspan="2" align="center"><em>Light mode with saved views &middot; the message context menu</em></td>
</tr>
</table>

<img src="docs/assets/screenshots/screenshot-compose.png" alt="The compose window in markdown mode with the send later menu open" width="100%">

<p align="center"><em>Composing in plain text, markdown or rich text, with send later</em></p>

## <img src="https://api.iconify.design/tabler/download.svg?color=white" width="26" style="vertical-align: -4px;"> Installation

| <img src="https://api.iconify.design/tabler/brand-apple.svg?color=white" width="18" style="vertical-align: -4px;"> [macOS](#-macos) | <img src="https://api.iconify.design/tabler/brand-windows.svg?color=white" width="18" style="vertical-align: -4px;"> [Windows](#-windows) | <img src="https://api.iconify.design/tabler/terminal-2.svg?color=white" width="18" style="vertical-align: -4px;"> [Linux](#-linux) |
|:---:|:---:|:---:|


### <img src="https://api.iconify.design/tabler/brand-apple.svg?color=white" width="26" style="vertical-align: -4px;"> macOS

1. Download `Pelton-<version>-macos-applesilicon.dmg` (Apple Silicon / M-series Macs).
2. Open the `.dmg` and drag `Pelton.app` into `Applications`.
3. Since the build is unsigned, macOS Gatekeeper blocks the first launch with an "unidentified developer" warning. Right-click (or Control-click) `Pelton.app` in Applications and choose **Open**, then confirm in the dialog that appears. You only need to do this once; after that it opens normally, including from Spotlight or the Dock.

A `.zip` of the raw `.app` (`Pelton-<version>-macos-<arch>-app.zip`) is also attached to each release if you'd rather not use the `.dmg`.

### <img src="https://api.iconify.design/tabler/brand-windows.svg?color=white" width="26" style="vertical-align: -4px;"> Windows

1. Download the latest installer from https://github.com/TRC-Loop/Pelton/releases/latest.
2. Run it. Since the build is unsigned, Windows SmartScreen may show an "unrecognized app" warning the first time. Click **More info** then **Run anyway**. This warning fades on its own as the file builds up download reputation.
3. The installer walks you through:
   - the GPL-3.0 license
   - the warranty and liability terms
   - **install for all users** (needs admin, installs to Program Files) or **just me** (installs to AppData) (no admin needed)
   - an optional desktop shortcut (Start Menu shortcut is always created)
   - a "Launch Pelton" checkbox on the last page

Windows on ARM is not supported yet.

### <img src="https://api.iconify.design/tabler/terminal-2.svg?color=white" width="26" style="vertical-align: -4px;"> Linux

**Currently supported:** `fedora (copr/dnf)`, `Arch (AUR)` – Thanks to [leeteral (AUR Maintainer)](https://leeism.com) [(Github)](https://github.com/leeteral)

[![Packaging status](https://repology.org/badge/vertical-allrepos/pelton.svg)](https://repology.org/project/pelton/versions)

#### fedora

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

### Other Linux distributions

On other distributions, build from source (needs Go, Node/pnpm, and the Wails CLI. See `make run` / `make build-linux` in the `Makefile`):

```sh
git clone https://github.com/TRC-Loop/Pelton.git
cd Pelton
make build-linux
```

This produces a binary plus a `.desktop` launcher in `build/bin/`; copy the binary somewhere on your `PATH`, install the `.desktop` file to `~/.local/share/applications/`, and give it an icon named `pelton` (see `build/icons/`).

## <img src="https://api.iconify.design/tabler/tag.svg?color=white" width="26" style="vertical-align: -4px;"> Versioning

Up to and including **1.0.9**, Pelton used semantic versioning. From the next release onward it uses **Calendar Versioning** in the form:

```
YYYY.Q.INCR
```

* **YYYY** — the full calendar year (e.g. `2026`).
* **Q** — the calendar quarter, `1`–`4` (Q1 = Jan–Mar, Q2 = Apr–Jun, Q3 = Jul–Sep, Q4 = Oct–Dec).
* **INCR** — the release counter within that quarter. It starts at `0` and resets to `0` at the start of each quarter.

The first release of a quarter (INCR `0`) drops the trailing `.0`, so it reads as just `YYYY.Q`. Examples:

| Release | Version |
| --- | --- |
| First release of Q4 2026 (flagship) | `2026.4` |
| Next release that quarter | `2026.4.1` |
| The one after | `2026.4.2` |
| First release of Q1 2027 | `2027.1` |

## <img src="https://api.iconify.design/tabler/messages.svg?color=white" width="26" style="vertical-align: -4px;"> Contact & Community

* **Discord:** Join the discussion at [discord.gg/UzPNGZYy6V](https://discord.gg/UzPNGZYy6V)
* **Email:** Reach out directly via [pelton@arne.sh](mailto:pelton@arne.sh)

## <img src="https://api.iconify.design/tabler/users.svg?color=white" width="26" style="vertical-align: -4px;"> Contributing

Contributions are welcome. Whether you are fixing bugs, refining the UI layout, polishing backend Go code, or enhancing documentation, feel free to open an issue or submit a Pull Request. See [AUTHORS.md](AUTHORS.md) if you'd like to be credited by name.

## <img src="https://api.iconify.design/tabler/file-certificate.svg?color=white" width="26" style="vertical-align: -4px;"> License

Pelton is distributed under the **[GPL-3.0 License](https://github.com/TRC-Loop/Pelton/blob/main/LICENSE)**. See `LICENSE` for details.

Warranty and liability are set out in **[DISCLAIMER.md](DISCLAIMER.md)**, an additional term under GPLv3 section 7(a), published as well at [pelton.app/terms](https://pelton.app/terms). Pelton comes without warranty and you use it at your own risk: it connects to your real mailboxes, deletions can be permanent, and it is not a backup tool. Keep your own backup of anything you cannot afford to lose.

Security issues belong in [SECURITY.md](SECURITY.md), not in a public issue.

## Other Info

Arne/TRC-Loop's timezone is `CET/CEST` or `UTC+2` or `Europe/Berlin`
<br>
![Time Badge](https://readmeme.eu.cc/api/time.svg?timezone=Europe%2FBerlin&theme=classic&timeFormat=24h&showSeconds=1&showDate=1&showDay=1&label=Arne%27s+Local+Time)
