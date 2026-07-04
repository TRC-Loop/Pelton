# Privacy and configurability

Read this before touching networking, sync, credentials, or any settings
surface.

## Privacy is a hard constraint

- No telemetry, no crash reporting, no analytics, ever. If you're tempted to
  add a network call for anything other than the user's own mail server or a
  feature they explicitly configured (OAuth, config sync to their own
  storage), stop and ask first.
- Remote content in mail (tracking pixels, remote images) is blocked by
  default and opt-in per message. Don't loosen this default.
- Secrets (passwords, OAuth tokens) go through `internal/credentials`
  (OS keyring), never into the SQLite DB or plain config files.
- Data lives in SQLite locally (`internal/storage`) and on the user's own
  IMAP/SMTP server. No third-party servers in the data path, ever.

## Configurability

Pelton's pitch is that almost everything is user-configurable: theme accent
color, density (compact/medium/luxe), keyboard shortcuts, swipe gestures,
image blocking, offline caching windows, flag colors, config export/import.
When adding a new UI feature, default to making its behavior a setting rather
than hardcoding one opinionated behavior, unless there's a strong reason not
to (e.g. it's a correctness/security matter, not a preference).

Settings surfaces live in `internal/desktop/bind_settings.go` (backend) and
`frontend/src/components/settings/` (UI). Config export/import goes through
`internal/configsync`.
