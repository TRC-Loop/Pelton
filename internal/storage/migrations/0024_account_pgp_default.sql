-- The account's starting point for protecting outgoing mail: '' leaves every
-- message unprotected until the user says otherwise, 'sign' signs whenever the
-- account has its own key, and 'auto' signs and encrypts whenever every
-- recipient has one.
--
-- It is only a default. The compose window always shows what is actually going
-- to happen to the message in front of you, and a default can never turn into a
-- send that would fail: with no key, 'sign' resolves back to unprotected.
ALTER TABLE accounts ADD COLUMN pgp_default TEXT NOT NULL DEFAULT '';
