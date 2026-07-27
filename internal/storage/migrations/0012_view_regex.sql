-- use_regex switches a view's text criteria (text/from/to/subject) from plain
-- substring matching to regular-expression matching. query_from and query_to
-- also gain multi-value support: they now hold a newline-separated list of
-- addresses, matched as OR within the field. A single stored address stays a
-- one-item list, so existing views keep working unchanged.
ALTER TABLE views ADD COLUMN use_regex INTEGER NOT NULL DEFAULT 0;
