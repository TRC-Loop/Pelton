-- last_seen_uid is the high water mark sync records per folder. basic sync still
-- compares the full uid set for correctness, this is a hint for later
-- incremental work and is reset to 0 when UIDVALIDITY changes.
ALTER TABLE folders ADD COLUMN last_seen_uid INTEGER NOT NULL DEFAULT 0;

-- pending push markers. pending_flags means a local flag change has not been
-- pushed to the server yet. pending_delete means the user deleted the message
-- locally and it still needs deleting on the server. sync clears these after a
-- successful push.
--
-- chosen over a separate pending_operations table because a message has at most
-- one pending flag state and at most one pending delete, so two columns on the
-- row stay simple and need no join. an op-log table would only pay off if we
-- needed ordered or multiple queued ops, which basic sync does not.
ALTER TABLE messages ADD COLUMN pending_flags INTEGER NOT NULL DEFAULT 0;
ALTER TABLE messages ADD COLUMN pending_delete INTEGER NOT NULL DEFAULT 0;

-- partial index so collecting the usually few pending rows stays cheap.
CREATE INDEX idx_messages_pending ON messages(folder_id)
WHERE pending_flags = 1 OR pending_delete = 1;
