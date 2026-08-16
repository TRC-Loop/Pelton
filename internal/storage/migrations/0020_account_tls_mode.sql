-- Connection security per account, stored instead of guessed from the port.
-- Deriving it meant STARTTLS was only reachable on 143 and 587, which locked
-- out anything on a custom port (Proton Mail Bridge, tunnels, test servers).
--
-- '' keeps the old behaviour for every existing account: the imap and smtp
-- layers read it as their TLSAuto mode and infer from the port, so nothing that
-- already works changes. 'ssl' and 'starttls' pin the transport explicitly.
ALTER TABLE accounts ADD COLUMN imap_tls TEXT NOT NULL DEFAULT '';
ALTER TABLE accounts ADD COLUMN smtp_tls TEXT NOT NULL DEFAULT '';
