-- signatures are reusable header/footer blocks the user writes once (in markdown
-- or html) and assigns as per-mailbox defaults, then optionally changes per
-- message in compose. kind is 'header' (inserted at the top of a new message) or
-- 'footer' (appended at the bottom). format is 'markdown' or 'html', so the
-- compose pane can insert the content appropriately for the editor mode.
CREATE TABLE signatures (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    kind       TEXT NOT NULL,
    format     TEXT NOT NULL DEFAULT 'markdown',
    content    TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- per-account default header/footer assignment. a null id means "no default"; a
-- deleted signature sets the reference back to null rather than blocking the
-- delete, so removing a block never strands an account.
CREATE TABLE account_signatures (
    account_id INTEGER PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    header_id  INTEGER REFERENCES signatures(id) ON DELETE SET NULL,
    footer_id  INTEGER REFERENCES signatures(id) ON DELETE SET NULL
);
