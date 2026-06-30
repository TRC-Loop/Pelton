-- settings is a key value store for ui preferences only. value holds a plain
-- string or a json document for structured values. nothing sensitive here.
CREATE TABLE settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
