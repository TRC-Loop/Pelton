-- How each account's last sync went (#322).
--
-- A sync run reported success as long as one account got through, so an
-- account that had been failing for weeks left nothing behind but a log line,
-- and logging is off by default. This is where the outcome goes so the ui can
-- mark the account, and so "when did this mailbox last actually sync" has an
-- answer that survives a restart.

CREATE TABLE IF NOT EXISTS account_sync_state (
    -- one row per account, written on every sync attempt.
    account_id INTEGER PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    -- when this account last synced through to the end. Kept across a later
    -- failure: how long it has been broken is the useful part.
    last_ok_at TEXT NOT NULL DEFAULT '',
    -- when the last failure was, empty once a sync succeeds again.
    failed_at TEXT NOT NULL DEFAULT '',
    -- what kind of failure it was: 'auth', 'network', 'credentials' or 'other'.
    -- The ui turns this into a sentence; detail carries the server's own words
    -- for the reader who wants them.
    reason TEXT NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT ''
);
