-- Per-account export of a local .eml copy when a message is archived, so mail
-- can be thinned from the server later without being lost.
--
-- Off by default, with no directory: nothing is ever written to disk until the
-- user turns it on and picks where the files go.
ALTER TABLE accounts ADD COLUMN export_on_archive INTEGER NOT NULL DEFAULT 0;
ALTER TABLE accounts ADD COLUMN export_dir TEXT NOT NULL DEFAULT '';
ALTER TABLE accounts ADD COLUMN export_subfolders TEXT NOT NULL DEFAULT 'none';
ALTER TABLE accounts ADD COLUMN export_name_template TEXT NOT NULL DEFAULT '';
