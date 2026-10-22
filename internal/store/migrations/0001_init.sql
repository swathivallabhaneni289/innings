-- Innings schema. Kept portable to Postgres: TEXT/INTEGER/REAL columns, RFC 3339 timestamps as TEXT,
-- JSON stored as TEXT (would be JSONB in Postgres), ON CONFLICT upserts, no AUTOINCREMENT.
-- Ids that point at simulated data (match ids, sim user ids, sim log ids) have no foreign keys.

CREATE TABLE users (
    id                TEXT PRIMARY KEY,
    token             TEXT NOT NULL UNIQUE,
    name              TEXT NOT NULL,
    handle            TEXT NOT NULL,
    avatar_colour     TEXT NOT NULL,
    favourite_team_id TEXT NOT NULL,
    bio               TEXT NOT NULL DEFAULT '',
    photo_uri         TEXT NOT NULL DEFAULT '',
    created_at        TEXT NOT NULL
);

CREATE TABLE logs (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_id        TEXT NOT NULL,
    rating          REAL NOT NULL,
    instant_classic INTEGER NOT NULL DEFAULT 0,
    player_ratings  TEXT NOT NULL DEFAULT '{}',
    fan_potm_id     TEXT,
    watched_at      TEXT NOT NULL,
    place_note      TEXT NOT NULL DEFAULT '',
    watched_with    TEXT NOT NULL DEFAULT '[]',
    companions      INTEGER,
    photo_uri       TEXT NOT NULL DEFAULT '',
    review          TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL,
    stub_no         INTEGER NOT NULL,
    UNIQUE (user_id, match_id)
);
CREATE INDEX logs_match ON logs (match_id);

CREATE TABLE predictions (
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_id            TEXT NOT NULL,
    winner_id           TEXT NOT NULL,
    top_run_scorer_id   TEXT NOT NULL,
    top_wicket_taker_id TEXT NOT NULL,
    created_at          TEXT NOT NULL,
    PRIMARY KEY (user_id, match_id)
);
CREATE INDEX predictions_match ON predictions (match_id);

CREATE TABLE follows (
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id  TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (user_id, target_id)
);

CREATE TABLE likes (
    log_id     TEXT NOT NULL,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL,
    PRIMARY KEY (log_id, user_id)
);

CREATE TABLE there_too (
    log_id     TEXT NOT NULL,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL,
    PRIMARY KEY (log_id, user_id)
);

CREATE TABLE replies (
    id         TEXT PRIMARY KEY,
    log_id     TEXT NOT NULL,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX replies_log ON replies (log_id);

CREATE TABLE top5 (
    user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    match_id TEXT NOT NULL,
    PRIMARY KEY (user_id, position)
);

CREATE TABLE watching (
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_id   TEXT NOT NULL,
    place      TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (user_id, match_id)
);
