// Package store persists real (signed-in) users and everything they do, in SQLite via modernc.org/sqlite.
// The SQL is kept portable to Postgres (see migrations/0001_init.sql); only Open and the pragmas are SQLite-specific.
package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Store struct{ db *sql.DB }

// Open opens (creating if needed) the database and applies pending migrations.
func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// One connection: SQLite serialises writes anyway, and this rules out SQLITE_BUSY between our own goroutines.
	// Never hold an open *sql.Rows while issuing another query (it would deadlock).
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return err
	}
	names, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return err
	}
	slices.Sort(names)
	for _, name := range names {
		version := strings.TrimSuffix(strings.TrimPrefix(name, "migrations/"), ".sql")
		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		body, err := migrationsFS.ReadFile(name)
		if err != nil {
			return err
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, ts(time.Now())); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// NewID returns prefix + 16 random hex chars.
func NewID(prefix string) string { return prefix + RandomHex(8) }

func RandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func ts(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// ---- users ----

const userCols = `id, name, handle, avatar_colour, favourite_team_id, bio, photo_uri`

func scanUser(sc interface{ Scan(...any) error }) (domain.User, error) {
	var u domain.User
	err := sc.Scan(&u.ID, &u.Name, &u.Handle, &u.AvatarColour, &u.FavouriteTeamID, &u.Bio, &u.PhotoURI)
	return u, err
}

func (s *Store) userWhere(ctx context.Context, where string, arg any) (domain.User, bool, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, `SELECT `+userCols+` FROM users WHERE `+where, arg))
	if errors.Is(err, sql.ErrNoRows) {
		return u, false, nil
	}
	return u, err == nil, err
}

func (s *Store) UserByID(ctx context.Context, id string) (domain.User, bool, error) {
	return s.userWhere(ctx, `id = ?`, id)
}

// Users returns every real user (onboarded or not), oldest first. See ListedUsers for the ones others may find.
func (s *Store) Users(ctx context.Context) ([]domain.User, error) {
	return s.queryUsers(ctx, `SELECT `+userCols+` FROM users ORDER BY created_at, id`)
}

// UpdateUser saves u's profile. ErrHandleTaken when another user got the same handle first (the lower(handle)
// unique index catches the race that a check-then-write would miss).
func (s *Store) UpdateUser(ctx context.Context, u domain.User) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET name = ?, handle = ?, avatar_colour = ?, favourite_team_id = ?, bio = ?, photo_uri = ? WHERE id = ?`,
		u.Name, u.Handle, u.AvatarColour, u.FavouriteTeamID, u.Bio, u.PhotoURI, u.ID)
	if isUniqueViolation(err) {
		return ErrHandleTaken
	}
	return err
}

// ---- logs ----

const logCols = `id, user_id, match_id, rating, instant_classic, player_ratings, fan_potm_id, watched_at, place_note,
	watched_with, companions, photo_uri, review, created_at, stub_no`

func scanLog(sc interface{ Scan(...any) error }) (domain.MatchLog, error) {
	var (
		l             domain.MatchLog
		instant       int
		ratings, with string
		potm          sql.NullString
		companions    sql.NullInt64
		watchedAt     string
	)
	if err := sc.Scan(&l.ID, &l.UserID, &l.MatchID, &l.Rating, &instant, &ratings, &potm, &watchedAt, &l.PlaceNote,
		&with, &companions, &l.PhotoURI, &l.Review, &l.CreatedAt, &l.StubNo); err != nil {
		return l, err
	}
	l.InstantClassic = instant != 0
	l.WatchedAt = domain.WatchPlace(watchedAt)
	l.PlayerRatings = map[string]int{}
	l.WatchedWith = []string{}
	_ = json.Unmarshal([]byte(ratings), &l.PlayerRatings)
	_ = json.Unmarshal([]byte(with), &l.WatchedWith)
	if l.PlayerRatings == nil {
		l.PlayerRatings = map[string]int{}
	}
	if l.WatchedWith == nil {
		l.WatchedWith = []string{}
	}
	if potm.Valid {
		l.FanPotmID = &potm.String
	}
	if companions.Valid {
		c := int(companions.Int64)
		l.Companions = &c
	}
	return l, nil
}

func (s *Store) queryLogs(ctx context.Context, where string, args ...any) ([]domain.MatchLog, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+logCols+` FROM logs `+where+` ORDER BY created_at DESC, stub_no DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MatchLog{}
	for rows.Next() {
		l, err := scanLog(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// Logs returns every real user's logs, newest first.
func (s *Store) Logs(ctx context.Context) ([]domain.MatchLog, error) { return s.queryLogs(ctx, "") }

func (s *Store) LogByID(ctx context.Context, id string) (domain.MatchLog, bool, error) {
	ls, err := s.queryLogs(ctx, `WHERE id = ?`, id)
	if err != nil || len(ls) == 0 {
		return domain.MatchLog{}, false, err
	}
	return ls[0], true, nil
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func upsertLog(ctx context.Context, ex execer, l domain.MatchLog) error {
	ratings, _ := json.Marshal(nonNilMap(l.PlayerRatings))
	with, _ := json.Marshal(nonNilSlice(l.WatchedWith))
	_, err := ex.ExecContext(ctx, `INSERT INTO logs (`+logCols+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id, match_id) DO UPDATE SET rating = excluded.rating, instant_classic = excluded.instant_classic,
		player_ratings = excluded.player_ratings, fan_potm_id = excluded.fan_potm_id, watched_at = excluded.watched_at,
		place_note = excluded.place_note, watched_with = excluded.watched_with, companions = excluded.companions,
		photo_uri = excluded.photo_uri, review = excluded.review`,
		l.ID, l.UserID, l.MatchID, l.Rating, boolInt(l.InstantClassic), string(ratings), l.FanPotmID, string(l.WatchedAt), l.PlaceNote,
		string(with), l.Companions, l.PhotoURI, l.Review, l.CreatedAt, l.StubNo)
	return err
}

// SaveLog creates the user's log for l.MatchID, or updates the existing one keeping its id, stubNo and createdAt.
// For a new log, ID and CreatedAt must be set; stubNo becomes the user's highest + 1.
func (s *Store) SaveLog(ctx context.Context, l domain.MatchLog) (saved domain.MatchLog, created bool, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return l, false, err
	}
	defer tx.Rollback()
	var id, createdAt string
	var stub int
	err = tx.QueryRowContext(ctx, `SELECT id, stub_no, created_at FROM logs WHERE user_id = ? AND match_id = ?`, l.UserID, l.MatchID).Scan(&id, &stub, &createdAt)
	switch {
	case err == nil:
		l.ID, l.StubNo, l.CreatedAt = id, stub, createdAt
	case errors.Is(err, sql.ErrNoRows):
		created = true
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(stub_no), 0) + 1 FROM logs WHERE user_id = ?`, l.UserID).Scan(&l.StubNo); err != nil {
			return l, false, err
		}
	default:
		return l, false, err
	}
	if err := upsertLog(ctx, tx, l); err != nil {
		return l, false, err
	}
	if err := tx.Commit(); err != nil {
		return l, false, err
	}
	l.PlayerRatings, l.WatchedWith = nonNilMap(l.PlayerRatings), nonNilSlice(l.WatchedWith)
	return l, created, nil
}

// DeleteLog removes userID's log logID with its likes, there-too taps and replies. false if not theirs / missing.
func (s *Store) DeleteLog(ctx context.Context, userID, logID string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `DELETE FROM logs WHERE id = ? AND user_id = ?`, logID, userID)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return false, nil
	}
	for _, q := range []string{`DELETE FROM likes WHERE log_id = ?`, `DELETE FROM there_too WHERE log_id = ?`, `DELETE FROM replies WHERE log_id = ?`} {
		if _, err := tx.ExecContext(ctx, q, logID); err != nil {
			return false, err
		}
	}
	return true, tx.Commit()
}

// ---- predictions ----

func (s *Store) Predictions(ctx context.Context) ([]domain.Prediction, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id, match_id, winner_id, top_run_scorer_id, top_wicket_taker_id, created_at FROM predictions ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Prediction{}
	for rows.Next() {
		var p domain.Prediction
		if err := rows.Scan(&p.UserID, &p.MatchID, &p.WinnerID, &p.TopRunScorerID, &p.TopWicketTakerID, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func upsertPrediction(ctx context.Context, ex execer, p domain.Prediction) error {
	_, err := ex.ExecContext(ctx, `INSERT INTO predictions (user_id, match_id, winner_id, top_run_scorer_id, top_wicket_taker_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT (user_id, match_id) DO UPDATE SET winner_id = excluded.winner_id,
		top_run_scorer_id = excluded.top_run_scorer_id, top_wicket_taker_id = excluded.top_wicket_taker_id, created_at = excluded.created_at`,
		p.UserID, p.MatchID, p.WinnerID, p.TopRunScorerID, p.TopWicketTakerID, p.CreatedAt)
	return err
}

func (s *Store) SavePrediction(ctx context.Context, p domain.Prediction) error {
	return upsertPrediction(ctx, s.db, p)
}

// ---- follows ----

// Follows returns the ids userID follows (sim or real users), oldest first.
func (s *Store) Follows(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT target_id FROM follows WHERE user_id = ? ORDER BY created_at, target_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *Store) ToggleFollow(ctx context.Context, userID, targetID string, now time.Time) (bool, error) {
	return s.toggle(ctx, `follows`, `user_id`, userID, `target_id`, targetID, now)
}

func (s *Store) toggle(ctx context.Context, table, colA, a, colB, b string, now time.Time) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE `+colA+` = ? AND `+colB+` = ?`, a, b)
	if err != nil {
		return false, err
	}
	on := false
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO `+table+` (`+colA+`, `+colB+`, created_at) VALUES (?, ?, ?)`, a, b, ts(now)); err != nil {
			return false, err
		}
		on = true
	}
	return on, tx.Commit()
}

// ---- likes / there too ----

// Reaction is one like or "I was there too" tap.
type Reaction struct {
	LogID, UserID, CreatedAt string
}

// Reaction tables.
const (
	Likes    = "likes"
	ThereToo = "there_too"
)

// ToggleReaction toggles userID's reaction (Likes or ThereToo) on logID; returns whether it is now on.
func (s *Store) ToggleReaction(ctx context.Context, table, logID, userID string, now time.Time) (bool, error) {
	if table != Likes && table != ThereToo {
		return false, fmt.Errorf("unknown reaction table %q", table)
	}
	return s.toggle(ctx, table, `log_id`, logID, `user_id`, userID, now)
}

// Reactions returns every row of a reaction table (Likes or ThereToo), oldest first.
func (s *Store) Reactions(ctx context.Context, table string) ([]Reaction, error) {
	if table != Likes && table != ThereToo {
		return nil, fmt.Errorf("unknown reaction table %q", table)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT log_id, user_id, created_at FROM `+table+` ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Reaction{}
	for rows.Next() {
		var r Reaction
		if err := rows.Scan(&r.LogID, &r.UserID, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---- replies ----

func (s *Store) AddReply(ctx context.Context, r domain.Reply) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO replies (id, log_id, user_id, text, created_at) VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.LogID, r.UserID, r.Text, r.CreatedAt)
	return err
}

// Replies returns every real reply, oldest first.
func (s *Store) Replies(ctx context.Context) ([]domain.Reply, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, log_id, user_id, text, created_at FROM replies ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Reply{}
	for rows.Next() {
		var r domain.Reply
		if err := rows.Scan(&r.ID, &r.LogID, &r.UserID, &r.Text, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---- top 5 ----

func (s *Store) Top5(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT match_id FROM top5 WHERE user_id = ? ORDER BY position`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *Store) SetTop5(ctx context.Context, userID string, matchIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM top5 WHERE user_id = ?`, userID); err != nil {
		return err
	}
	for i, id := range matchIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO top5 (user_id, position, match_id) VALUES (?, ?, ?)`, userID, i, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ---- watching ----

// Watching returns userId → place for everyone who said they're watching matchID.
func (s *Store) Watching(ctx context.Context, matchID string) (map[string]domain.WatchPlace, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id, place FROM watching WHERE match_id = ?`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]domain.WatchPlace{}
	for rows.Next() {
		var u, p string
		if err := rows.Scan(&u, &p); err != nil {
			return nil, err
		}
		out[u] = domain.WatchPlace(p)
	}
	return out, rows.Err()
}

// SetWatching sets (or with nil clears) where userID is watching matchID.
func (s *Store) SetWatching(ctx context.Context, userID, matchID string, place *domain.WatchPlace, now time.Time) error {
	if place == nil {
		_, err := s.db.ExecContext(ctx, `DELETE FROM watching WHERE user_id = ? AND match_id = ?`, userID, matchID)
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO watching (user_id, match_id, place, updated_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id, match_id) DO UPDATE SET place = excluded.place, updated_at = excluded.updated_at`,
		userID, matchID, string(*place), ts(now))
	return err
}

// ---- demo history ----

// SeedHistory writes demo logs, predictions and follows for userID in one transaction. Logs get fresh ids and
// sequential stub numbers (oldest first) so sim ids can never collide between users. New accounts start empty, so
// the server no longer calls it; it is kept for seeding demo data.
func (s *Store) SeedHistory(ctx context.Context, userID string, logs []domain.MatchLog, preds []domain.Prediction, follows []string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	sorted := slices.Clone(logs)
	slices.SortStableFunc(sorted, func(a, b domain.MatchLog) int { return strings.Compare(a.CreatedAt, b.CreatedAt) })
	seen := map[string]bool{}
	stub := 0
	for _, l := range sorted {
		if seen[l.MatchID] {
			continue
		}
		seen[l.MatchID] = true
		stub++
		l.ID, l.UserID, l.StubNo = NewID("log-"), userID, stub
		if l.CreatedAt == "" {
			l.CreatedAt = ts(now)
		}
		if err := upsertLog(ctx, tx, l); err != nil {
			return err
		}
	}
	for _, p := range preds {
		p.UserID = userID
		if err := upsertPrediction(ctx, tx, p); err != nil {
			return err
		}
	}
	for _, f := range follows {
		if _, err := tx.ExecContext(ctx, `INSERT INTO follows (user_id, target_id, created_at) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`, userID, f, ts(now)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nonNilMap(m map[string]int) map[string]int {
	if m == nil {
		return map[string]int{}
	}
	return m
}

func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
