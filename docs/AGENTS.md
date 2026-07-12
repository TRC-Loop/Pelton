# Pelton, machine-readable documentation

Pelton is a free, open source, privacy-focused email client for macOS, Windows and Linux. Go + Wails backend, Svelte frontend. No telemetry. Website: https://pelton.app. Source: https://github.com/TRC-Loop/Pelton. Human docs: https://docs.pelton.app.

## Install

- macOS: .dmg per release at https://github.com/TRC-Loop/Pelton/releases (unsigned; right-click Open or `xattr -cr /Applications/Pelton.app`)
- Windows: `Pelton-<version>-windows-amd64-installer.exe` from releases (unsigned)
- Fedora: `sudo dnf copr enable arnek/Pelton && sudo dnf install pelton`
- Raw rpm: `Pelton-<version>-linux-fedora-x86_64.rpm` from releases
- Source: Go + pnpm + wails CLI v2.13.0; `wails build`; on Linux `-tags webkit2_41` with gtk3-devel and webkit2gtk4.1-devel

## Data locations

- macOS: `~/Library/Application Support/Pelton`
- Linux: `~/.config/Pelton`
- Windows: `%AppData%\Pelton`
- Contents: `pelton.db` (settings + mail cache), search index, `themes/<id>/` (installed themes)
- Credentials: OS keyring only (Keychain / Credential Manager / Secret Service)
- Backup: Settings > Import/Export, JSON export, credentials optionally included encrypted (AES-256-GCM, scrypt)

## Network behavior

Only: user's IMAP/SMTP servers; GitHub releases API if update check enabled (default off); BIMI/Gravatar avatar lookups unless avatar source set to generated; remote mail images only after user approval (default blocked). No telemetry.

## Mail providers

App passwords required by most large providers. iCloud: appleid.apple.com > App-Specific Passwords. Gmail: enable 2-Step Verification, then https://myaccount.google.com/apppasswords.

## Default shortcuts

mod = Cmd on macOS, Ctrl elsewhere. mod+N compose, mod+F search, mod+R sync, mod+M add mailbox, mod+, settings, mod+P export PDF, mod+Z undo send/delete/archive, Ctrl+Cmd+F fullscreen. Message-level actions (reply, reply-all, forward, read/unread, flag, snooze, archive, delete, offline download) are unbound by default; user binds them in Settings > Shortcuts.

## Theme format (.peltontheme)

Zip container. `manifest.json` at archive root is the only fixed name; it references everything else.

manifest.json fields:
- manifestVersion (int, required): format version, currently 1; newer than the app understands = import refused
- id (slug, optional): install folder name and update-detection key; defaults to slug of name
- name (required), author, version, description, homepage, license
- base (required): "light" | "dark", the built-in token set that fills unoverridden tokens
- pelton (optional): { "min": "x.y.z", "max": "x.y" } app version range; outside range = warning badge, never a block
- tokens (optional): list of JSON file paths merged in order (later wins), or one inline object
- css (optional): list of CSS files concatenated in order
- preview (optional): screenshot path for the gallery
- icons (optional): map of icon name (tabler name, lowercase kebab, no Icon prefix) to SVG path

Token files: flat maps of token name (no `--` prefix needed) to CSS value. Allowlisted tokens only:
surfaces (surface-base/raised/overlay/sunken/hover, selection-bg, selection-bg-strong), text (text-primary/secondary/tertiary/inverse, link), borders (border-subtle/default/strong, hairline), accent (accent, accent-fg), semantic (success/-bg, warning/-bg, danger/-bg), radii (radius-control/card/none), fonts (font-ui, font-mono), type (fz-meta/label/list/body/heading/title, fw-regular/medium/semibold/bold), elevation (shadow-overlay). Spacing/density tokens are not themeable. Values reject `; { } @ url(` and control chars.

CSS rules: relative url("assets/...") refs are inlined as data: URIs at apply time (bundle fonts/images this way); remote url()/@import are listed at import with a keep-or-strip choice (strip default), the choice is baked into installed files. Caps: 20 MB container, 1 MB CSS total, 5 MB per inlined asset, 256 KB per icon SVG.

Icon SVGs: must use currentColor, are sanitized at import and load (no scripts, no event handlers, no href, no foreignObject, no url()); unknown icon names are ignored for forward compatibility.

Install location: `<data dir>/themes/<id>/`, extracted and hand-editable; Settings > Themes > Reload re-validates and re-reads. Export zips the folder back to a shareable file.

## Legal

- Website: https://pelton.app
- Imprint / Impressum: https://pelton.app/imprint/en/
- Privacy: https://pelton.app/privacy/
