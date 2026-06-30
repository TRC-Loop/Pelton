-- accounts holds only non sensitive metadata. credentials live in the os
-- keyring, referenced by account id, and never touch this database.
CREATE TABLE accounts (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    email        TEXT    NOT NULL,
    display_name TEXT    NOT NULL DEFAULT '',
    imap_host    TEXT    NOT NULL,
    imap_port    INTEGER NOT NULL,
    smtp_host    TEXT    NOT NULL DEFAULT '',
    smtp_port    INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT    NOT NULL
);

-- folders mirrors the server mailbox hierarchy per account. imap_path is the
-- raw mailbox name the server returned, and delimiter is that server's
-- hierarchy separator, both stored because they differ across providers.
-- uid_validity lets us detect a server side reset and invalidate the cache.
CREATE TABLE folders (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id   INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name         TEXT    NOT NULL,
    imap_path    TEXT    NOT NULL,
    delimiter    TEXT    NOT NULL DEFAULT '',
    parent_id    INTEGER REFERENCES folders(id) ON DELETE CASCADE,
    attributes   TEXT    NOT NULL DEFAULT '',
    uid_validity INTEGER NOT NULL DEFAULT 0,
    UNIQUE(account_id, imap_path)
);

CREATE INDEX idx_folders_account ON folders(account_id);
CREATE INDEX idx_folders_parent ON folders(parent_id);

-- messages caches envelope metadata and bodies. uid is the stable imap
-- identifier (never a sequence number) and is unique within its folder.
-- flags is a bitmask, see the flag constants in messages.go: one integer
-- column stays compact and maps directly onto the imap flag set instead of
-- growing a new boolean column every time a flag is added.
CREATE TABLE messages (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    folder_id       INTEGER NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
    uid             INTEGER NOT NULL,
    message_id      TEXT    NOT NULL DEFAULT '',
    subject         TEXT    NOT NULL DEFAULT '',
    from_address    TEXT    NOT NULL DEFAULT '',
    from_name       TEXT    NOT NULL DEFAULT '',
    to_addresses    TEXT    NOT NULL DEFAULT '',
    cc_addresses    TEXT    NOT NULL DEFAULT '',
    date            TEXT    NOT NULL DEFAULT '',
    flags           INTEGER NOT NULL DEFAULT 0,
    body_plain      TEXT    NOT NULL DEFAULT '',
    body_html       TEXT    NOT NULL DEFAULT '',
    has_attachments INTEGER NOT NULL DEFAULT 0,
    size_bytes      INTEGER NOT NULL DEFAULT 0,
    UNIQUE(folder_id, uid)
);

CREATE INDEX idx_messages_folder ON messages(folder_id);
CREATE INDEX idx_messages_account ON messages(account_id);
CREATE INDEX idx_messages_message_id ON messages(message_id);
