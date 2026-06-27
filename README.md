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

> [!WARNING]
> **Pelton is currently a Work In Progress (WIP).** Features, APIs, and the user interface are under heavy active development. It is not yet ready for production use or primary email management.

## <img src="https://api.iconify.design/tabler/info-circle.svg?color=white" width="26" style="vertical-align: -4px;"> About Pelton

Pelton is a modern, Free and Open-Source Software (FOSS) email client written in Go and built using the Wails framework. It is designed from the ground up to respect your data. We believe your inbox belongs to you, which is why Pelton operates with absolute zero telemetry and full privacy. 

## <img src="https://api.iconify.design/tabler/sparkles.svg?color=white" width="26" style="vertical-align: -4px;"> Features

* <img src="https://api.iconify.design/tabler/shield-lock.svg?color=white" width="18" style="vertical-align: -4px;"> **Full Privacy:** Your data stays on your machine. Zero tracking, zero telemetry, and complete control over your inbox.
* <img src="https://api.iconify.design/tabler/bolt.svg?color=white" width="18" style="vertical-align: -4px;"> **Fast Search:** Find what you need instantly. The search engine is optimized for speed and handles large local mailboxes with ease.
* <img src="https://api.iconify.design/tabler/palette.svg?color=white" width="18" style="vertical-align: -4px;"> **Highly Customizable:** Tailor the client to fit your exact workflow and aesthetic preferences.
* <img src="https://api.iconify.design/tabler/file-export.svg?color=white" width="18" style="vertical-align: -4px;"> **Portable Configuration:** Export your entire setup, including accounts, preferences, and custom layouts, into a single easily transferable file.
* <img src="https://api.iconify.design/tabler/code.svg?color=white" width="18" style="vertical-align: -4px;"> **FOSS & Cross-Platform:** Truly open source and built to run beautifully across different operating systems.

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
<p>Your data stays entirely in your control. It is stored in a lightning-fast SQLite database on your local machine, and on your original email provider's server. We do not host or route your data through any third-party servers.</p>
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

Installation instructions and executable binary packages will be documented here once the client reaches its alpha testing phase.

## <img src="https://api.iconify.design/tabler/messages.svg?color=white" width="26" style="vertical-align: -4px;"> Contact & Community

* **Discord:** Join the discussion at [arne.sh/discord](https://arne.sh/discord)
* **Email:** Reach out directly via [pelton@arne.sh](mailto:pelton@arne.sh)

## <img src="https://api.iconify.design/tabler/users.svg?color=white" width="26" style="vertical-align: -4px;"> Contributing

Contributions are welcome. Whether you are fixing bugs, refining the UI layout, polishing backend Go code, or enhancing documentation, feel free to open an issue or submit a Pull Request.

## <img src="https://api.iconify.design/tabler/file-certificate.svg?color=white" width="26" style="vertical-align: -4px;"> License

Pelton is distributed under the **[GPL-3.0 License](https://github.com/TRC-Loop/Pelton/blob/main/LICENSE)**. See `LICENSE` for details.
