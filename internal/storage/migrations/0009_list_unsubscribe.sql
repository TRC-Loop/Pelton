-- List-Unsubscribe support (#71): persist the unsubscribe headers at sync
-- time so the reading pane can offer the button without refetching. Empty
-- means the message advertised none (or was cached before this migration).
ALTER TABLE messages ADD COLUMN list_unsubscribe TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN list_unsubscribe_post INTEGER NOT NULL DEFAULT 0;
