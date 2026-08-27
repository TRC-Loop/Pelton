-- Folders the user unchecked, so a mailbox with a huge archive does not have to
-- sync it. Defaults to 0, so every existing folder and every newly discovered
-- one syncs as before and only an explicit uncheck changes anything.
ALTER TABLE folders ADD COLUMN sync_excluded INTEGER NOT NULL DEFAULT 0;
