-- Revocation checking for s/mime signatures (#226).
--
-- Verification happens once, at sync, and the verdict is stored. Revocation
-- cannot work that way: a certificate that was in force when the mail arrived
-- can be withdrawn the week after, which is precisely the case worth catching.
-- So the check runs when a message is opened, and what is stored here is the
-- material it needs plus the answer it got.

-- the signing certificate and its issuer, in der, concatenated with a length
-- prefix per entry. The raw message is not kept after a sync, and asking an
-- authority about a certificate needs the certificate itself, so this is the
-- only way a later check has anything to work from. Empty for the overwhelming
-- majority of mail, which carries no signature at all.
ALTER TABLE messages ADD COLUMN smime_certs BLOB NOT NULL DEFAULT x'';

-- sha-256 of the signing certificate, the key into the cache below. Kept on the
-- message so looking up an answer costs no parsing.
ALTER TABLE messages ADD COLUMN smime_fingerprint TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_messages_smime_fingerprint
    ON messages(smime_fingerprint) WHERE smime_fingerprint != '';

-- one row per certificate, not per message: a thread from one sender is one
-- question, asked once. next_update is the authority's own word on how long its
-- answer holds, so the cache expires when it says rather than on a guess.
CREATE TABLE smime_revocation (
    fingerprint TEXT PRIMARY KEY,
    -- good, revoked, or unknown when the authority could not be reached.
    status      TEXT NOT NULL,
    detail      TEXT NOT NULL DEFAULT '',
    revoked_at  TEXT NOT NULL DEFAULT '',
    checked_at  TEXT NOT NULL,
    next_update TEXT NOT NULL DEFAULT ''
);
