-- the raw rfc 822 source of pgp protected mail, kept so it can be decrypted and
-- verified whenever it is opened.
--
-- Decryption needs the exact bytes the sender produced, and those are not
-- otherwise cached: the source viewer refetches them over imap. Without this,
-- encrypted mail would be unreadable offline and every open would cost a
-- network round trip. What is stored is the ciphertext, already encrypted by
-- the sender, so keeping it gives away nothing that the message did not already
-- carry. The decrypted plaintext is never written anywhere.
--
-- A separate table rather than a column on messages: this is a blob, and the
-- list queries select every message column, so putting it inline would drag the
-- whole ciphertext through queries that only want a subject and a date. It also
-- holds only the small minority of mail that is protected.
CREATE TABLE message_pgp (
    message_id INTEGER PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    raw        BLOB NOT NULL
);
