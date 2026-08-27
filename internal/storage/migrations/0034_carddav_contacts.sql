-- Contacts from a CardDAV server (#168).
--
-- The address book Pelton had was harvested from mail: every address you wrote
-- to or heard from, ranked by how often. That is useful and it is not an
-- address book. These tables hold the real one, synced from the user's own
-- server, editable here and written back.

CREATE TABLE IF NOT EXISTS address_books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- the mail account this book was discovered from, or null for one the user
    -- added by hand. Only used to explain where it came from; the book keeps
    -- working if the account is removed, which is why this is nullable rather
    -- than a cascade.
    account_id INTEGER REFERENCES accounts(id) ON DELETE SET NULL,
    -- what the server calls it, or what the user renamed it to.
    name TEXT NOT NULL DEFAULT '',
    -- the endpoint the client connects to, and the collection path inside it.
    url TEXT NOT NULL,
    collection_path TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL DEFAULT '',
    -- the sync-collection token from the last successful run. Empty means the
    -- next sync reads the whole book, which is also the recovery path when a
    -- server forgets a token.
    sync_token TEXT NOT NULL DEFAULT '',
    -- set when the server refuses writes, so the ui can say why editing is off
    -- rather than failing at save time.
    read_only INTEGER NOT NULL DEFAULT 0,
    last_sync_at TEXT NOT NULL DEFAULT '',
    -- the last sync failure in the server's own words, cleared on success.
    last_error TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL REFERENCES address_books(id) ON DELETE CASCADE,
    -- where the contact lives on the server and the version this copy came
    -- from. The etag is what makes a conflicting edit detectable.
    path TEXT NOT NULL,
    etag TEXT NOT NULL DEFAULT '',
    uid TEXT NOT NULL DEFAULT '',
    full_name TEXT NOT NULL DEFAULT '',
    organization TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    -- the vCard exactly as the server holds it. A card written by another
    -- client can carry a birthday, a photo, a key url; editing a phone number
    -- here must not be what deletes them, so every write starts from this.
    card TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL DEFAULT '',
    UNIQUE (book_id, path)
);

-- addresses and numbers are their own rows because a contact has several and
-- autocomplete searches them directly.
CREATE TABLE IF NOT EXISTS contact_emails (
    contact_id INTEGER NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    -- the vCard TYPE, lowercased: 'home', 'work', or empty.
    label TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS contact_phones (
    contact_id INTEGER NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    label TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_contact_emails ON contact_emails(email);
CREATE INDEX IF NOT EXISTS idx_contact_emails_contact ON contact_emails(contact_id);
CREATE INDEX IF NOT EXISTS idx_contact_phones_contact ON contact_phones(contact_id);
CREATE INDEX IF NOT EXISTS idx_contacts_book ON contacts(book_id);
