-- A chat message can be a fan posting their own call (internal/feed/calls.go): call_id names it, '' for a plain
-- message.
ALTER TABLE chat_messages ADD COLUMN call_id TEXT NOT NULL DEFAULT '';
