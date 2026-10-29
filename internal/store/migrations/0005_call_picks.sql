-- Fans' picks on live calls (internal/feed/calls.go): one per fan per call, changeable while the call is open.
CREATE TABLE call_picks (
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    call_id    TEXT NOT NULL,
    match_id   TEXT NOT NULL,
    option_id  TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (user_id, call_id)
);
CREATE INDEX call_picks_match ON call_picks (match_id);
