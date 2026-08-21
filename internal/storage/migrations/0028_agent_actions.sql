-- MCP write actions (#127).
--
-- Two tables, for the two things an agent doing more than reading needs: a
-- record of what it did, and a queue for the one action it is not allowed to
-- complete on its own.

-- every write an agent made, so a move or a deletion can be told from one the
-- user made themselves. Without this there is no way to answer "why is this in
-- Archive", which is the question that matters most when something went wrong.
CREATE TABLE agent_actions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    -- the mcp tool name, which is also the permission key.
    tool       TEXT NOT NULL,
    -- the message acted on, 0 for an action that names none.
    message_id INTEGER NOT NULL DEFAULT 0,
    -- a short, already-worded line for the log view. Written by the backend,
    -- never by the agent, so nothing an agent says lands in the ui as fact.
    summary    TEXT NOT NULL DEFAULT '',
    -- '' when the action succeeded, otherwise why it did not.
    error      TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);

CREATE INDEX idx_agent_actions_created ON agent_actions(created_at DESC);

-- mail an agent proposes sending. It is not a draft and never touches the
-- server: an unapproved message must not be uploaded anywhere, so it waits
-- here until the person at the keyboard reads it and says yes.
CREATE TABLE agent_proposals (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    to_addrs   TEXT NOT NULL DEFAULT '',
    cc_addrs   TEXT NOT NULL DEFAULT '',
    bcc_addrs  TEXT NOT NULL DEFAULT '',
    subject    TEXT NOT NULL DEFAULT '',
    body       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);
