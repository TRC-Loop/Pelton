# Backend (Go)

Read this before touching `internal/` or `cmd/`.

## Layout

One package per concern: `imap`, `smtp`, `storage`, `sync`, `crypto`,
`credentials`, `oauth`, `search`, `outbox`, `configsync`, `autoconfig`,
`mailview`. `internal/desktop` is the Wails bind layer exposed to the
frontend (`bind_*.go` files group bindings by feature). `cmd/` holds small
standalone test/debug binaries (imaptest, smtptest, storagetest, synctest),
not the main app.

## Docstrings

Every exported function, type, and package-level var/const gets a GoDoc
comment stating the behavior and edge cases plainly, matching the existing
style in `internal/storage`, e.g.:

```go
// GetAccount returns one account by id, or ErrAccountNotFound.

// DeleteAccount removes an account. Its folders, messages and attachment
// rows cascade away; attachment files on disk are the caller's concern.
```

Unexported helpers only need a comment if their purpose or a constraint
isn't obvious from the name and body.

## Testing

Table-driven tests alongside the code (`*_test.go`), see `internal/crypto`,
`internal/smtp`, `internal/sync`, `internal/outbox` for the existing style.
Add or update tests for any backend logic change, especially storage, sync,
crypto, and parsing code. Message/attachment handling in particular has had
several encoding/filename/race-condition bugs (see recent `fix:` commits) —
new code touching that path needs a test that would have caught the last bug
in the area, not just a happy-path test.

Don't skip or weaken an existing test to make a change land.

## Concurrency

Sync, downloads, and other heavy IO run off the main/UI thread via
goroutines — never block the Wails main thread with network or disk work.
