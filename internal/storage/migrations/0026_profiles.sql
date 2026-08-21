-- Profiles (#270): one install keeping work and private life apart without two
-- copies of the app.
--
-- Everything stays in one database and a profile is a scope rather than a
-- directory. That is what makes sharing possible per account: an address can be
-- visible in two profiles with one copy of its mail, synced once.
--
-- What a profile scopes: which accounts it shows, its settings (which is also
-- where the VIP list and every ui preference live), its signatures and its
-- saved views. Everything else, the mail itself included, is the install's.
--
-- Sharing is per area. share_* = 1 means the profile reads the main profile's
-- rows and writes land there too, so a change shows up in both. 0 means it owns
-- its own. Copying from another profile at creation time is a one-off duplicate
-- of the rows, so it needs nothing here.

CREATE TABLE profiles (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    -- an emoji or short glyph the ui shows next to the name, so the current
    -- profile is recognisable at a glance rather than by reading.
    icon       TEXT    NOT NULL DEFAULT '',
    position   INTEGER NOT NULL DEFAULT 0,
    -- the profile the install started as. It can be renamed but never deleted:
    -- it owns the rows the sharing profiles point at, and switching has to have
    -- somewhere to land.
    is_main    INTEGER NOT NULL DEFAULT 0,
    -- exactly one row has this set; it is the profile the app opens.
    is_active  INTEGER NOT NULL DEFAULT 0,
    share_settings   INTEGER NOT NULL DEFAULT 0,
    share_signatures INTEGER NOT NULL DEFAULT 0,
    share_views      INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL
);

-- the profile every existing install already is. The table is new and empty, so
-- this is id 1, which the backfills below rely on.
INSERT INTO profiles (name, icon, position, is_main, is_active, created_at)
VALUES ('Main', '', 1, 1, 1, strftime('%Y-%m-%dT%H:%M:%SZ', 'now'));

-- which accounts a profile shows. A join rather than a column on accounts,
-- because one account can be in several profiles.
CREATE TABLE profile_accounts (
    profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    PRIMARY KEY (profile_id, account_id)
);

INSERT INTO profile_accounts (profile_id, account_id)
SELECT (SELECT id FROM profiles WHERE is_main = 1), id FROM accounts;

-- settings become per profile. The key alone was the primary key, and sqlite
-- cannot widen one in place, so the table is rebuilt. Every existing row
-- belongs to main, which is the profile the install has been all along.
CREATE TABLE settings_scoped (
    key        TEXT    NOT NULL,
    profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    value      TEXT    NOT NULL,
    updated_at TEXT    NOT NULL,
    PRIMARY KEY (key, profile_id)
);

INSERT INTO settings_scoped (key, profile_id, value, updated_at)
SELECT key, (SELECT id FROM profiles WHERE is_main = 1), value, updated_at FROM settings;

DROP TABLE settings;
ALTER TABLE settings_scoped RENAME TO settings;

-- signatures and saved views carry their owning profile. Existing rows are
-- main's.
ALTER TABLE signatures ADD COLUMN profile_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE views ADD COLUMN profile_id INTEGER NOT NULL DEFAULT 1;

UPDATE signatures SET profile_id = (SELECT id FROM profiles WHERE is_main = 1);
UPDATE views SET profile_id = (SELECT id FROM profiles WHERE is_main = 1);

CREATE INDEX idx_signatures_profile ON signatures(profile_id);
CREATE INDEX idx_views_profile ON views(profile_id);
