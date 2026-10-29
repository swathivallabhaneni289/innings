package store

import (
	"context"
	"time"
)

// Feed row kinds (migrations/0003_feed.sql).
const (
	FeedMatch  = "match"
	FeedTeam   = "team"
	FeedPlayer = "player"
)

// SaveFeedRows upserts rows (id → JSON body) of one kind in a single transaction.
func (s *Store) SaveFeedRows(ctx context.Context, kind string, rows map[string][]byte, now time.Time) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for id, body := range rows {
		if _, err := tx.ExecContext(ctx, `INSERT INTO feed_rows (kind, id, body, updated_at) VALUES (?, ?, ?, ?)
			ON CONFLICT (kind, id) DO UPDATE SET body = excluded.body, updated_at = excluded.updated_at`, kind, id, string(body), ts(now)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// FeedRows returns every row of one kind, id → JSON body.
func (s *Store) FeedRows(ctx context.Context, kind string) (map[string][]byte, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, body FROM feed_rows WHERE kind = ?`, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]byte{}
	for rows.Next() {
		var id, body string
		if err := rows.Scan(&id, &body); err != nil {
			return nil, err
		}
		out[id] = []byte(body)
	}
	return out, rows.Err()
}
