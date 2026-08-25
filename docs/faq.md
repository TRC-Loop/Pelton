# FAQ

## macOS says the app could not be verified

Pelton builds are not notarized (notarization requires a paid Apple developer account). Press **Done** in the dialog, then allow the app under **System Settings, Privacy & Security, Open Anyway**. The exact steps are in [Install](install.md); `xattr -cr /Applications/Pelton.app` in a terminal does the same in one step.

## Does Pelton phone home?

No. There is no telemetry and nothing to opt out of. The complete list of network connections Pelton can make:

- your own IMAP and SMTP servers, always
- the GitHub releases API, only if you enable update checks (off by default)
- BIMI and Gravatar lookups for sender avatars, unless you switch the avatar source to generated placeholders
- remote images inside mails, only after you allow them (blocked by default)
- VirusTotal, only if you enable the integration and supply your own API key (off by default, see [Scanning links and attachments](virustotal.md))

Nothing else, ever. Themes with remote CSS references are flagged at import for exactly this reason.

## A mail shows as gibberish, and it looks fine in another client

Some mail does not say which encoding its text uses, and some names one no
table knows. Pelton works it out from the text itself, the same way other
clients do, and marks the message with a small badge in the reading pane so you
know it was a guess.

If your mail keeps coming from one system that gets this wrong in the same way,
you can pick a fixed encoding under **Settings, Display, Text encoding
fallback** instead of leaving it on detection.

Mail that was already in your cache before this existed was stored broken and
cannot be repaired locally, so Pelton fetches those messages from your server
again, a few per sync.

## Can I rename a mailbox without telling everyone I write to?

Yes. **Settings, Mailboxes** has two names. The **From name** is your outward identity and goes in the From header of every message you send. Turning on **Use a different name in the sidebar** adds a **sidebar label**, which is what Pelton calls the mailbox everywhere on screen. The label stays on your machine, so calling a mailbox "work junk" is between you and your sidebar.

Switching the toggle back off keeps the label you typed, in case you want it again.

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
