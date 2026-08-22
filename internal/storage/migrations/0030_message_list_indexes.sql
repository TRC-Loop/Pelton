-- Indexes for the message list (#310).
--
-- The list reads one folder, or a handful for a unified view, newest first.
-- Until now the only index on messages was folder_id alone, so every list load
-- found the folder's rows and then sorted all of them in a temporary b-tree to
-- pick the newest fifty, reading each row's bodies to do it. At 14k messages
-- that is 165ms for the first page and 664ms for the fourth, on every load,
-- which is what made the window feel frozen during a first sync and slow
-- afterwards. With the ordering in the index the same reads are under a
-- millisecond.

-- the list index. date and uid descending match the ORDER BY exactly, so the
-- rows come out of the index already in order and LIMIT stops early.
CREATE INDEX idx_messages_list ON messages(folder_id, date DESC, uid DESC);

-- unread badges and folder counts ask about flags rather than dates, and
-- answering them from the row store meant touching every row in the folder.
-- These four columns are everything those queries read, so the answer comes out
-- of the index.
CREATE INDEX idx_messages_state ON messages(folder_id, flags, pending_delete, snooze_hidden);

-- redundant now: folder_id is the leading column of both indexes above, so
-- anything that used this can use those instead, and one less index is one less
-- write per stored message.
DROP INDEX IF EXISTS idx_messages_folder;
