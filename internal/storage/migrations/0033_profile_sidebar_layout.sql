-- The sidebar layout follows the profile (#325).
--
-- Pinned folders and the order of both folders and account sections were
-- columns on the rows themselves, so they were one arrangement for the whole
-- install. Profiles are meant to be separate working contexts, and they
-- already show different sets of accounts, so an order shared across all of
-- them orders a list you cannot see the whole of anyway.

-- one row per folder the user has arranged, per profile. A folder nobody
-- touched has no row and sorts by id, which is what position 0 used to mean.
CREATE TABLE IF NOT EXISTS profile_sidebar_layout (
    profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    folder_id INTEGER NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
    -- rank among siblings in the account tree, from 1.
    position INTEGER NOT NULL DEFAULT 0,
    -- rank in the Pinned group, from 1. 0 means not pinned in this profile.
    pinned_position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (profile_id, folder_id)
);

-- the same for the account sections.
CREATE TABLE IF NOT EXISTS profile_account_order (
    profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (profile_id, account_id)
);

-- whether this profile reads the main profile's layout instead of its own,
-- alongside the switches for settings, signatures and views. Off by default:
-- a new working context arranges itself. A profile that shares keeps its own
-- rows untouched, so unticking the box brings its arrangement back.
ALTER TABLE profiles ADD COLUMN share_layout INTEGER NOT NULL DEFAULT 0;

-- every profile starts from the arrangement the install already had, so
-- nothing looks like it was reset. Only folders and accounts the user actually
-- moved are copied; the rest have no row and keep sorting by id.
INSERT INTO profile_sidebar_layout (profile_id, folder_id, position, pinned_position)
SELECT p.id, f.id, f.position, f.pinned_position
FROM profiles p, folders f
WHERE f.position > 0 OR f.pinned_position > 0;

INSERT INTO profile_account_order (profile_id, account_id, position)
SELECT p.id, a.id, a.position
FROM profiles p, accounts a
WHERE a.position > 0;

-- the old columns would be a second, stale answer to the same question.
ALTER TABLE folders DROP COLUMN position;
ALTER TABLE folders DROP COLUMN pinned_position;
ALTER TABLE accounts DROP COLUMN position;
