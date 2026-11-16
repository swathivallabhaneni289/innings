-- Email and password sign-in (docs/API.md → Auth), in place of phone numbers and texted codes. Additive only: the
-- phone columns and the otp_codes and otp_sends tables from 0002 stay in the file, but nothing reads or writes them
-- any more, so a local innings.db keeps everything it had. Accounts made with a phone number have no email or
-- password, so they can't sign in and their old sessions stop working; `make reset-db` clears them.

ALTER TABLE users ADD COLUMN email TEXT;             -- trimmed and lower-cased; NULL for accounts from the phone era
ALTER TABLE users ADD COLUMN email_hash TEXT;        -- SHA-256 hex of email, for contact matching
ALTER TABLE users ADD COLUMN password_hash TEXT;     -- pbkdf2-sha256$<iterations>$<salt>$<key> (internal/auth)
ALTER TABLE users ADD COLUMN email_verified_at TEXT; -- set when the emailed code is confirmed; NULL until then

CREATE UNIQUE INDEX users_email ON users (email) WHERE email IS NOT NULL;
CREATE INDEX users_email_hash ON users (email_hash);

-- The live code per address and purpose: 'verify' confirms the address after sign-up, 'reset' starts a new
-- password. code_hash is the SHA-256 of the 6 digits.
CREATE TABLE email_codes (
    email      TEXT NOT NULL,
    purpose    TEXT NOT NULL,
    code_hash  TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    attempts   INTEGER NOT NULL DEFAULT 0,
    sent_at    TEXT NOT NULL,
    PRIMARY KEY (email, purpose)
);

-- Every code made for an address, for the resend gap and the hourly cap. Rows older than an hour are pruned as new
-- codes are made.
CREATE TABLE email_sends (
    email   TEXT NOT NULL,
    sent_at TEXT NOT NULL
);
CREATE INDEX email_sends_email ON email_sends (email, sent_at);

-- The right reset code buys one of these: for 15 minutes, its holder may set a new password. Only the SHA-256 of the
-- token is stored.
CREATE TABLE reset_tokens (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TEXT NOT NULL
);
CREATE INDEX reset_tokens_user ON reset_tokens (user_id);
