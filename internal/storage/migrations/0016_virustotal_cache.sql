-- cached VirusTotal verdicts, so re-opening a message does not re-ask about
-- links and attachments it has already looked up. VirusTotal's free tier
-- allows only a few requests per minute and a few hundred per day, which a
-- single auto-scanned newsletter can exhaust on its own.
--
-- target is the url, or the lowercase sha-256 of the attachment. Rows exist
-- only for targets the user chose to scan, and are deleted wholesale when the
-- integration is turned off.
CREATE TABLE IF NOT EXISTS virustotal_verdicts (
    kind       TEXT    NOT NULL,
    target     TEXT    NOT NULL,
    status     TEXT    NOT NULL,
    malicious  INTEGER NOT NULL DEFAULT 0,
    suspicious INTEGER NOT NULL DEFAULT 0,
    total      INTEGER NOT NULL DEFAULT 0,
    permalink  TEXT    NOT NULL DEFAULT '',
    checked_at TEXT    NOT NULL,
    PRIMARY KEY (kind, target)
);
