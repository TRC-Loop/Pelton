-- Charset detection for messages that are wrong about their own encoding
-- (#311).
--
-- Text that named no charset, or named one no table knows, used to be stored
-- exactly as the sender wrote it and rendered as utf-8 later, which is where
-- the mojibake came from. It is now detected and converted at the parser, and
-- these two columns are what the rest of the app needs to know about that.

-- what the body was read as when it had to be guessed at: an encoding name, or
-- 'detected' when the decode happened inside the charset hook and the table it
-- settled on is not reported back. Empty for mail that declared a charset that
-- exists, which is nearly all of it.
ALTER TABLE messages ADD COLUMN charset_guess TEXT NOT NULL DEFAULT '';

-- set on mail that was cached before the fix and is sitting in the database
-- with bytes that are not valid utf-8. Nothing local can undo that reliably:
-- the raw source is not kept, so the repair is to ask the server for the
-- message again, which the next sync of that folder does. A message the server
-- no longer has keeps the mark and stays as it is.
ALTER TABLE messages ADD COLUMN needs_refetch INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_messages_needs_refetch
    ON messages(folder_id) WHERE needs_refetch = 1;
