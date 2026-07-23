-- Separate login username (#108): some servers authenticate on a username that
-- is not the email address. Empty means "log in with the email address", which
-- keeps every account created before this migration working unchanged.
ALTER TABLE accounts ADD COLUMN username TEXT NOT NULL DEFAULT '';
