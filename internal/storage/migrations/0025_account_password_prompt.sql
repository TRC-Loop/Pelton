-- Whether the user told the missing-password prompt to stop asking for this
-- account. The prompt interrupts every sync until a password is stored, and
-- there are good reasons to leave a mailbox unusable for a while, so the answer
-- has to survive a restart.
--
-- Storing a password clears it again: the next time one goes missing is a new
-- situation and worth one prompt.
ALTER TABLE accounts ADD COLUMN password_prompt_dismissed INTEGER NOT NULL DEFAULT 0;
