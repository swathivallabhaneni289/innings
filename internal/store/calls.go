package store

import (
	"context"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// SaveCallPick sets (or changes) a fan's pick on a live call.
func (s *Store) SaveCallPick(ctx context.Context, p domain.CallPick) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO call_picks (user_id, call_id, match_id, option_id, created_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id, call_id) DO UPDATE SET option_id = excluded.option_id, created_at = excluded.created_at`,
		p.UserID, p.CallID, p.MatchID, p.OptionID, p.CreatedAt)
	return err
}

// CallPicks returns every pick on every live call.
func (s *Store) CallPicks(ctx context.Context) ([]domain.CallPick, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id, call_id, match_id, option_id, created_at FROM call_picks ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CallPick{}
	for rows.Next() {
		var p domain.CallPick
		if err := rows.Scan(&p.UserID, &p.CallID, &p.MatchID, &p.OptionID, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
