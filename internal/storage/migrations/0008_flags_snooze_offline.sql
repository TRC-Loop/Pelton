-- Per-message extras that back several ui features. All are additive columns
-- with safe defaults so existing cached rows keep working untouched.
--
-- flag_color: 0 means no color; 1..8 map to the eight palette colors. This is
-- separate from the flags bitmask (which models the boolean imap flags) because
-- a color is a small enum, not a flag bit, and can optionally sync to an imap
-- keyword ($Label1..$Label8) when the user turns syncing on.
--
-- snooze_until / snooze_hidden drive the local snooze ("resend to me"): a row is
-- revived (marked unread) once snooze_until passes; snooze_hidden hides it from
-- the list in the meantime when the user asked to also hide it now.
--
-- offline marks a message the user has explicitly pinned to keep offline (via the
-- single download action or a bulk range download), which drives the little
-- downloaded indicator. Normal sync already caches bodies; this is the deliberate
-- "keep this available" signal, distinct from incidental caching.
ALTER TABLE messages ADD COLUMN flag_color    INTEGER NOT NULL DEFAULT 0;
ALTER TABLE messages ADD COLUMN snooze_until  TEXT    NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN snooze_hidden INTEGER NOT NULL DEFAULT 0;
ALTER TABLE messages ADD COLUMN offline       INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_messages_snooze ON messages(snooze_until) WHERE snooze_until != '';

-- address_book harvests every address seen in sent/received mail for compose
-- autocomplete. use_count and last_used drive both ranking and the eviction
-- policy (least used, then oldest) once the book grows past its cap.
CREATE TABLE address_book (
    email      TEXT    PRIMARY KEY,
    name       TEXT    NOT NULL DEFAULT '',
    use_count  INTEGER NOT NULL DEFAULT 1,
    last_used  TEXT    NOT NULL,
    created_at TEXT    NOT NULL
);

CREATE INDEX idx_address_book_rank ON address_book(use_count DESC, last_used DESC);
