-- What the receiving server said about the sender's authentication, read out of
-- the Authentication-Results header at fetch time. The header is not kept after
-- parsing, so these columns are the only record of it.
--
-- Empty means the header said nothing, which is not a failure, and it is also
-- what every message cached before this columns existed carries. Reply-To goes
-- here too: a reply that would leave for another domain is one of the signals.
ALTER TABLE messages ADD COLUMN auth_spf TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN auth_dkim TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN auth_dmarc TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN auth_spf_domain TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN auth_dkim_domain TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN reply_to TEXT NOT NULL DEFAULT '';
