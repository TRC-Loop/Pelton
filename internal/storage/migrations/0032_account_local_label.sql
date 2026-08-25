-- A mailbox name for you, separate from the one recipients get (#326).
--
-- display_name did both jobs: it went out in the From header and it was what
-- the sidebar showed. So renaming a mailbox to something useful locally, like
-- "work junk", shipped that name to everyone you wrote to.

-- what this app calls the mailbox: sidebar, mailbox list, command palette.
-- Never sent anywhere. Empty means it was never set.
ALTER TABLE accounts ADD COLUMN local_label TEXT NOT NULL DEFAULT '';

-- whether local_label is in use. Kept separate from the label itself so
-- turning it off leaves the name you typed where it was, ready to switch back
-- on. Existing accounts start off, which is what makes this change invisible
-- to them.
ALTER TABLE accounts ADD COLUMN use_local_label INTEGER NOT NULL DEFAULT 0;
