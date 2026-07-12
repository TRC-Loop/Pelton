# FAQ

## macOS says the app is damaged or from an unidentified developer

Pelton builds are not notarized (notarization requires a paid Apple developer account). Right-click the app and choose **Open**, or run `xattr -cr /Applications/Pelton.app` once. See [Install](install.md).

## Does Pelton phone home?

No. There is no telemetry and nothing to opt out of. The complete list of network connections Pelton can make:

- your own IMAP and SMTP servers, always
- the GitHub releases API, only if you enable update checks (off by default)
- BIMI and Gravatar lookups for sender avatars, unless you switch the avatar source to generated placeholders
- remote images inside mails, only after you allow them (blocked by default)

Nothing else, ever. Themes with remote CSS references are flagged at import for exactly this reason.

## How do I back up my accounts and settings?

**Settings, Import / Export**. The export is a single JSON file with your settings, mailboxes and signatures. Mail credentials can be included too, encrypted with a password you choose. The cached mail itself is not part of the export; it re-syncs from your server after an import.

## Gmail keeps rejecting my password

Google refuses normal passwords for IMAP. Enable 2-Step Verification on your Google account, then create an app password at [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords) and use that in Pelton. A smoother Gmail story is on the roadmap.

## The interface was smaller than the window on Linux

That was a WebKitGTK viewport bug, fixed in Pelton 1.0.7. Update via `dnf` or grab the latest release. If you ever need to debug rendering, clicking the version number in **Settings, About** toggles a viewport readout you can screenshot.

## The menu bar language did not change until I restarted (Linux)

Known platform limitation: rebuilding the native GTK menu live crashes inside the toolkit, so on Linux the menu picks up a language change at the next launch. An in-app menu bar that live-updates (and follows your theme) is planned.

## Where does an update come from when I click "Check now"?

The check compares your version against the latest tag on the public GitHub releases API and links you to the release page. Nothing is downloaded or installed automatically.

## Can a theme mess with my mail or send data somewhere?

Not silently. Themes cannot run scripts, icon SVGs are sanitized, token values are validated, and CSS network references require your explicit approval at import, where the raw CSS is shown first. If you decline, they are stripped from the installed copy. See [Themes](themes/index.md).
