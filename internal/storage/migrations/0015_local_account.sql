-- Local Folders: mail imported from a file or another mail client, which has
-- no server behind it. It hangs off an ordinary account row, so the sidebar,
-- message list, search and attachment storage need no special case; is_local
-- marks it so sync, idle and the mailbox backup skip it. Nothing stored under
-- a local account is ever sent to a server.
ALTER TABLE accounts ADD COLUMN is_local INTEGER NOT NULL DEFAULT 0;
