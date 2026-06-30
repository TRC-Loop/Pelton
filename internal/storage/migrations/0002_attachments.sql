-- attachments stores metadata only. the bytes live on disk, never as blobs
-- here. disk_path is relative to the attachments root so the config dir stays
-- portable if the user moves it. content_id is set for inline cid images.
CREATE TABLE attachments (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id   INTEGER NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    filename     TEXT    NOT NULL,
    content_type TEXT    NOT NULL DEFAULT '',
    size_bytes   INTEGER NOT NULL DEFAULT 0,
    content_id   TEXT    NOT NULL DEFAULT '',
    disk_path    TEXT    NOT NULL
);

CREATE INDEX idx_attachments_message ON attachments(message_id);
