-- The match chat: fans talking while a match is on, one room per match. created_at is RFC 3339 UTC with
-- milliseconds in a fixed width ("2026-10-29T15:04:05.123Z"), so it still sorts as text: a busy room gets several
-- messages a second.
CREATE TABLE chat_messages (
    id         TEXT PRIMARY KEY,
    match_id   TEXT NOT NULL,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX chat_messages_match ON chat_messages (match_id, created_at);
CREATE INDEX chat_messages_user ON chat_messages (user_id, created_at);
