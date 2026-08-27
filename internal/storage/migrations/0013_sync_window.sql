-- sync_floor_uid is the lowest uid a folder's cache covers. messages with a
-- lower uid still exist on the server but are deliberately not fetched, so the
-- first sync of a ten year old mailbox does not download 2016 before the user
-- sees today's mail (#175). 0 means no floor: the whole folder is cached, which
-- is also the state every folder synced by an older version is already in.
ALTER TABLE folders ADD COLUMN sync_floor_uid INTEGER NOT NULL DEFAULT 0;
