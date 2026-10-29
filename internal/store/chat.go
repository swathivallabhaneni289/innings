package store

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

func (s *Store) AddChatMessage(ctx context.Context, m domain.ChatMessage) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO chat_messages (id, match_id, user_id, text, created_at) VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.MatchID, m.UserID, m.Text, m.CreatedAt)
	return err
}

// ChatMessages returns a match's messages, oldest first: the latest `limit` when after is "", else up to `limit`
// from after on (inclusive, so a message stamped in the same millisecond is never skipped; callers drop the ones
// they have by id). Messages in the same millisecond keep the order they arrived in (SQLite's rowid; on Postgres
// this would be a serial column).
func (s *Store) ChatMessages(ctx context.Context, matchID, after string, limit int) ([]domain.ChatMessage, error) {
	q := `SELECT id, match_id, user_id, text, created_at FROM chat_messages WHERE match_id = ? ORDER BY created_at DESC, rowid DESC LIMIT ?`
	args := []any{matchID, limit}
	if after != "" {
		q = `SELECT id, match_id, user_id, text, created_at FROM chat_messages WHERE match_id = ? AND created_at >= ? ORDER BY created_at, rowid LIMIT ?`
		args = []any{matchID, after, limit}
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ChatMessage{}
	for rows.Next() {
		var m domain.ChatMessage
		if err := rows.Scan(&m.ID, &m.MatchID, &m.UserID, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if after == "" {
		slices.Reverse(out)
	}
	return out, nil
}

// LastChatAt is when userID last posted in any match chat ("" if never).
func (s *Store) LastChatAt(ctx context.Context, userID string) (string, error) {
	var at string
	err := s.db.QueryRowContext(ctx, `SELECT created_at FROM chat_messages WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&at)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return at, err
}
