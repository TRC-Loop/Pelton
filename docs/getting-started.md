# Getting started

## Add your first mailbox

On first launch Pelton walks you through a short tour and asks for your first mailbox. You can always add more with ++cmd+m++ (++ctrl+m++ on Windows and Linux) or via **Mailbox, Add Mailbox** in the menu.

Enter your email address and Pelton discovers the IMAP and SMTP servers for most providers automatically. If discovery comes up empty, the hosts and ports can be entered by hand; your provider's help pages list them, usually under "IMAP settings".

Passwords are stored in the operating system keyring (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux), never in plain files.

## App passwords

Most large providers refuse your normal account password for IMAP and want an app password instead:

- **iCloud**: sign in at [appleid.apple.com](https://appleid.apple.com), open **Sign-In and Security, App-Specific Passwords**, generate one and use it as the password in Pelton. Pelton also shows this guidance inline when it detects an iCloud address.
- **Gmail**: turn on 2-Step Verification, then create an app password at [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords) and use it as the password. Making Gmail less painful is on the roadmap.
- **Others**: look for "app password" or "application password" in your provider's security settings. If the provider offers plain IMAP with your normal password, that works too.

## How sync works

Pelton keeps an IMAP IDLE connection open per account, so new mail is pushed to you as it arrives. On top of that, a full sync pass runs periodically (every 15 minutes by default, configurable down to off in **Settings, Power**). ++cmd+r++ syncs on demand.

Heavy work like sync and bulk downloads runs in the background; the interface stays responsive throughout.

## Where your data lives

Everything Pelton stores locally sits in one folder:

| Platform | Path |
| -------- | ---- |
| macOS | `~/Library/Application Support/Pelton` |
| Linux | `~/.config/Pelton` |
| Windows | `%AppData%\Pelton` |

Inside you will find the local mail cache and settings database (`pelton.db`), the search index, and installed themes under `themes/`. Credentials are not in there; they live in the OS keyring.

For backups, use **Settings, Import / Export**. It exports settings, mailboxes and signatures as a JSON file, and can include your mail credentials encrypted with a password you choose (AES-256-GCM, scrypt-derived key).

## Privacy defaults

- Remote images in mail are blocked by default. Allow them per message, per sender domain, or globally in **Settings, Privacy**.
- Update checks are off by default. When enabled, Pelton only compares version tags against the public GitHub releases API.
- Sender avatars use BIMI and Gravatar lookups by default. Set the avatar source to generated placeholders in **Settings, Avatars** if you want zero avatar-related network requests.
- There is no telemetry of any kind. Nothing to opt out of.
