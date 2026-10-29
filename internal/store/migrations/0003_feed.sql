-- Real match data (internal/feed): what the scores feed last told us, kept so a match stays findable after the
-- feed stops listing it (fans' logs and predictions point at it) and so a restart starts from what we had.
--
-- kind is 'match', 'team' or 'player'; body is that object as JSON.
CREATE TABLE feed_rows (
    kind       TEXT NOT NULL,
    id         TEXT NOT NULL,
    body       TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (kind, id)
);
