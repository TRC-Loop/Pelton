-- sidebar ordering and pinning. all three columns are local display state;
-- nothing here is sent to or read from a server.
--
-- position is 0 until the user drags something in that group, and the list
-- queries sort unpositioned rows last by id, so an existing install keeps its
-- current order (discovery order for folders, creation order for accounts) and
-- a newly discovered folder lands at the end of its group instead of jumping to
-- the front.
ALTER TABLE folders ADD COLUMN position INTEGER NOT NULL DEFAULT 0;
ALTER TABLE accounts ADD COLUMN position INTEGER NOT NULL DEFAULT 0;

-- pinned_position doubles as the pinned flag: 0 means not pinned, anything
-- higher is the folder's rank in the Pinned group. One column rather than a
-- separate boolean, so there is no way to store a pinned folder with no rank.
-- A pinned folder still appears in its own account's tree; the Pinned group is
-- a mirror, not a move.
ALTER TABLE folders ADD COLUMN pinned_position INTEGER NOT NULL DEFAULT 0;
