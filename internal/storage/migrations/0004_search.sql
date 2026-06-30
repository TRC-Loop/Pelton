-- full text index over messages. external content (content='messages') keeps
-- the text in the messages table only, the fts table holds just the index, and
-- the triggers below keep the two in sync on insert, update and delete.
CREATE VIRTUAL TABLE messages_fts USING fts5(
    subject,
    body_plain,
    from_address,
    content='messages',
    content_rowid='id'
);

CREATE TRIGGER messages_fts_ai AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(rowid, subject, body_plain, from_address)
    VALUES (new.id, new.subject, new.body_plain, new.from_address);
END;

-- the special 'delete' row tells fts5 to remove the old index entry. external
-- content tables require the old column values to be passed back here.
CREATE TRIGGER messages_fts_ad AFTER DELETE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, subject, body_plain, from_address)
    VALUES ('delete', old.id, old.subject, old.body_plain, old.from_address);
END;

CREATE TRIGGER messages_fts_au AFTER UPDATE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, subject, body_plain, from_address)
    VALUES ('delete', old.id, old.subject, old.body_plain, old.from_address);
    INSERT INTO messages_fts(rowid, subject, body_plain, from_address)
    VALUES (new.id, new.subject, new.body_plain, new.from_address);
END;
