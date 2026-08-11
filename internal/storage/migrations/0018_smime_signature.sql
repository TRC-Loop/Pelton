-- the outcome of verifying a received message's s/mime signature.
--
-- Verification runs during sync, where the raw rfc 822 bytes it needs are
-- already in hand. Those bytes are not cached (the source viewer refetches them
-- over imap), so verifying when a message is opened would mean a network round
-- trip per message and no verdict at all offline. The verdict is therefore
-- recorded once, and describes the signature as it stood when the mail arrived,
-- which is the question a signature answers.
--
-- smime_status is '' for the overwhelming majority of mail, which carries no
-- signature at all, and otherwise one of valid, untrusted or invalid. The
-- remaining columns describe the signing certificate for display and are empty
-- when there is no signature. Messages cached before this existed keep '' and
-- show no verdict until they are synced again.
ALTER TABLE messages ADD COLUMN smime_status TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN smime_signer TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN smime_email TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN smime_issuer TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN smime_detail TEXT NOT NULL DEFAULT '';
