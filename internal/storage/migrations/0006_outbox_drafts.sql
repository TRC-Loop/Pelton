-- the outbox is the queue every send goes through. enqueuing is instant, a
-- background worker drains it, so sending feels immediate online and survives
-- offline by staying queued and retrying.
--
-- raw_message is the fully built mime, already encrypted when encryption was
-- requested, so plaintext of a protected message is never written here.
--
-- envelope_from and recipients store the smtp envelope. recipients is needed
-- separately from raw_message because Bcc recipients must receive the mail but
-- deliberately do not appear in the transmitted headers, so they cannot be
-- recovered by parsing raw_message. recipients is newline separated.
--
-- state is one of the values owned by internal/outbox (queued/sending/sent/
-- failed); it is stored as text rather than an enum so the queue package is the
-- single source of truth for the allowed values. next_attempt_at drives the
-- retry backoff: a queued row is only eligible once now has passed it.
CREATE TABLE outbox (
    id              INTEGER PRIMARY KEY,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    envelope_from   TEXT NOT NULL,
    recipients      TEXT NOT NULL,
    raw_message     BLOB NOT NULL,
    state           TEXT NOT NULL,
    attempts        INTEGER NOT NULL DEFAULT 0,
    last_error      TEXT NOT NULL DEFAULT '',
    next_attempt_at TEXT NOT NULL,
    created_at      TEXT NOT NULL
);

-- the worker claims the oldest due, queued row, so index by the columns it
-- filters and orders on.
CREATE INDEX idx_outbox_due ON outbox(state, next_attempt_at);

-- drafts note: a draft is saved by an imap APPEND to the Drafts folder with the
-- \Draft flag, it does not need a row here. replacing an edited server draft
-- would mean remembering the appended draft's uid (returned by APPEND when the
-- server supports UIDPLUS) and, on re-save, appending the new version then
-- deleting the old uid. that uid would live on the messages row for the draft
-- when draft editing is built; it is a later refinement and needs no schema now.
