-- Phone sign-in (docs/API.md → Auth). Still portable to Postgres: ADD COLUMN, partial and expression indexes,
-- RFC 3339 UTC timestamps as TEXT (compared as strings, which orders correctly because the format is fixed).
--
-- users.token (NOT NULL UNIQUE since 0001) held the guest bearer token. Sessions now live in their own table and are
-- looked up by hash, so users.token is never read again: new accounts fill it with an unusable unique value
-- ("acct-" + id) to satisfy the constraint, and old guest tokens stop working by design (the app sends those fans
-- to sign-in).

ALTER TABLE users ADD COLUMN phone TEXT;        -- E.164; NULL for guests from before sign-in existed
ALTER TABLE users ADD COLUMN phone_hash TEXT;   -- SHA-256 hex of the E.164 string, for contact matching
ALTER TABLE users ADD COLUMN onboarded_at TEXT; -- set once a name and @username are chosen; NULL until then

CREATE UNIQUE INDEX users_phone ON users (phone) WHERE phone IS NOT NULL;
CREATE INDEX users_phone_hash ON users (phone_hash);
-- New accounts start with handle '' (many at once), so only chosen handles must be unique.
CREATE UNIQUE INDEX users_handle_lower ON users (lower(handle)) WHERE handle <> '';

-- One row per signed-in device. Only the SHA-256 of the bearer token is stored.
CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL
);
CREATE INDEX sessions_user ON sessions (user_id);

-- The live sign-in code per number. code_hash is the SHA-256 of a code this server made (dev mode) and '' when the
-- SMS provider made and holds the code (Twilio Verify); either way expiry and tries are tracked here.
CREATE TABLE otp_codes (
    phone      TEXT PRIMARY KEY,
    code_hash  TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    attempts   INTEGER NOT NULL DEFAULT 0,
    sent_at    TEXT NOT NULL
);

-- Every code sent, for the per-number hourly cap. Rows older than an hour are pruned as new codes go out.
CREATE TABLE otp_sends (
    phone   TEXT NOT NULL,
    sent_at TEXT NOT NULL
);
CREATE INDEX otp_sends_phone ON otp_sends (phone, sent_at);
