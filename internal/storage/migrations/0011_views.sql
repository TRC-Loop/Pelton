-- views are user-defined saved searches ("preset searches") the user pins as
-- their own entries in the sidebar/menu. a view is a stored search query plus
-- scope filters, with a name, icon and accent color for display. every view is
-- eager-run on startup and after each sync to keep a live match count; there is
-- no per-view body prefetch, because sync already caches every message body and
-- attachment locally.
CREATE TABLE views (
    id             INTEGER PRIMARY KEY,
    name           TEXT NOT NULL,
    icon           TEXT NOT NULL DEFAULT '',
    color          TEXT NOT NULL DEFAULT '',

    -- search query fields, mirroring internal/search.Query. an empty field is
    -- unconstrained. when all four are empty the view is a pure scope filter and
    -- runs against the store directly instead of the full-text index.
    query_text     TEXT NOT NULL DEFAULT '',
    query_from     TEXT NOT NULL DEFAULT '',
    query_to       TEXT NOT NULL DEFAULT '',
    query_subject  TEXT NOT NULL DEFAULT '',

    -- relative date window in days: 0 means no date bound, otherwise the view
    -- matches mail newer than (now - within_days). relative rather than absolute
    -- so a saved view does not go stale.
    within_days    INTEGER NOT NULL DEFAULT 0,

    -- scope filters applied on top of the query. flags are not in the full-text
    -- index, so these are evaluated against the store.
    unread_only    INTEGER NOT NULL DEFAULT 0,
    flagged_only   INTEGER NOT NULL DEFAULT 0,
    has_attachment INTEGER NOT NULL DEFAULT 0,

    -- account scope: null means all accounts, otherwise restrict to one. a
    -- deleted account drops its scoped views rather than orphaning them.
    account_id     INTEGER REFERENCES accounts(id) ON DELETE CASCADE,

    -- sidebar ordering, low to high.
    position       INTEGER NOT NULL DEFAULT 0,

    created_at     TEXT NOT NULL,
    updated_at     TEXT NOT NULL
);
