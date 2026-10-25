package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// Accounts, sessions and sign-in codes (migrations/0002_accounts.sql, docs/API.md → Auth).

var (
	// ErrHandleTaken: another user already holds this @username (compared case-insensitively).
	ErrHandleTaken = errors.New("handle taken")
	// ErrPhoneTaken: an account with this phone exists (a concurrent first sign-in for the same number won).
	ErrPhoneTaken = errors.New("phone taken")
	// ErrTooSoon: a code went to this number less than the resend gap ago.
	ErrTooSoon = errors.New("code sent too recently")
	// ErrTooMany: this number has had its hourly allowance of codes.
	ErrTooMany = errors.New("too many codes this hour")
)

// isUniqueViolation matches SQLite's and Postgres's unique-constraint errors by message, so callers can turn a lost
// race on a unique index into a domain error without importing a driver.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "duplicate key value")
}

// PhoneHash is the SHA-256 hex of an E.164 number: users.phone_hash, and what the app sends for contact matching.
func PhoneHash(phone string) string {
	sum := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(sum[:])
}

func parseTS(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// ---- accounts ----

// CreateAccount inserts a fan who just verified phone for the first time. users.token gets "acct-"+id: the column
// is NOT NULL UNIQUE from the guest era but sessions no longer use it, so any unique value nobody can present will do.
func (s *Store) CreateAccount(ctx context.Context, u domain.User, phone string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO users (id, token, name, handle, avatar_colour, favourite_team_id, bio, photo_uri, created_at, phone, phone_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, "acct-"+u.ID, u.Name, u.Handle, u.AvatarColour, u.FavouriteTeamID, u.Bio, u.PhotoURI, ts(now), phone, PhoneHash(phone))
	if isUniqueViolation(err) {
		return ErrPhoneTaken
	}
	return err
}

func (s *Store) UserByPhone(ctx context.Context, phone string) (domain.User, bool, error) {
	return s.userWhere(ctx, `phone = ?`, phone)
}

// Account returns userID's private account details (phone is "" for a pre-sign-in guest).
func (s *Store) Account(ctx context.Context, userID string) (domain.Account, error) {
	var a domain.Account
	var phone, onboarded sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT phone, onboarded_at, created_at FROM users WHERE id = ?`, userID).Scan(&phone, &onboarded, &a.CreatedAt)
	a.Phone, a.Onboarded = phone.String, onboarded.Valid
	return a, err
}

// SetOnboarded marks userID as having finished sign-up; the first time is kept if it is called again.
func (s *Store) SetOnboarded(ctx context.Context, userID string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET onboarded_at = COALESCE(onboarded_at, ?) WHERE id = ?`, ts(now), userID)
	return err
}

// ListedUsers returns the real users who finished sign-up, oldest first. Only they may show up in search, friend
// lists and contact matches; half-signed-up accounts and guests from before sign-in stay out of sight.
func (s *Store) ListedUsers(ctx context.Context) ([]domain.User, error) {
	return s.queryUsers(ctx, `SELECT `+userCols+` FROM users WHERE onboarded_at IS NOT NULL ORDER BY created_at, id`)
}

func (s *Store) queryUsers(ctx context.Context, query string, args ...any) ([]domain.User, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// HandleTaken reports whether a real user other than exceptID holds handle, ignoring case.
func (s *Store) HandleTaken(ctx context.Context, handle, exceptID string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE lower(handle) = ? AND id <> ?`, strings.ToLower(handle), exceptID).Scan(&n)
	return n > 0, err
}

// phoneHashChunk keeps each IN list well under SQLite's and Postgres's bound-parameter limits.
const phoneHashChunk = 500

// ListedUsersByPhoneHash returns the onboarded users whose phone hash is in hashes, ordered by name. It queries in
// chunks and keeps nothing.
func (s *Store) ListedUsersByPhoneHash(ctx context.Context, hashes []string) ([]domain.User, error) {
	out := []domain.User{}
	for start := 0; start < len(hashes); start += phoneHashChunk {
		chunk := hashes[start:min(start+phoneHashChunk, len(hashes))]
		args := make([]any, len(chunk))
		for i, h := range chunk {
			args[i] = h
		}
		marks := strings.TrimSuffix(strings.Repeat("?, ", len(chunk)), ", ")
		us, err := s.queryUsers(ctx, `SELECT `+userCols+` FROM users WHERE onboarded_at IS NOT NULL AND phone_hash IN (`+marks+`)`, args...)
		if err != nil {
			return nil, err
		}
		out = append(out, us...)
	}
	return out, nil
}

// ---- sessions ----

// CreateSession records a signed-in device. Only the SHA-256 of its bearer token is ever stored.
func (s *Store) CreateSession(ctx context.Context, tokenHash, userID string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO sessions (token_hash, user_id, created_at) VALUES (?, ?, ?)`, tokenHash, userID, ts(now))
	return err
}

// UserBySession resolves a session's token hash to its user.
func (s *Store) UserBySession(ctx context.Context, tokenHash string) (domain.User, bool, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, `SELECT u.id, u.name, u.handle, u.avatar_colour, u.favourite_team_id, u.bio, u.photo_uri
		FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token_hash = ?`, tokenHash))
	if errors.Is(err, sql.ErrNoRows) {
		return u, false, nil
	}
	return u, err == nil, err
}

// DeleteSession signs one device out.
func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// ---- sign-in codes ----

// OTP is the live sign-in code for one number. CodeHash is the SHA-256 hex of a code this server made, or "" when
// the SMS provider made and holds it. SentAt identifies the code: a new code always has a later SentAt.
type OTP struct {
	Phone     string
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
	SentAt    time.Time
}

// StartOTP makes o the number's live code if the resend gap and the hourly cap allow, logging the send. Checking and
// writing in one transaction means concurrent starts can't both slip under a limit (SQLite runs it serially on our
// single connection; Postgres would want SELECT ... FOR UPDATE or SERIALIZABLE). On refusal it returns ErrTooSoon or
// ErrTooMany and how long to wait. It also prunes expired codes and sends older than an hour, for every number, so
// phone numbers aren't kept longer than the limits need them.
func (s *Store) StartOTP(ctx context.Context, o OTP, gap time.Duration, perHour int) (time.Duration, error) {
	now := o.SentAt
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM otp_sends WHERE sent_at <= ?`, ts(now.Add(-time.Hour))); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM otp_codes WHERE expires_at <= ?`, ts(now)); err != nil {
		return 0, err
	}
	var first, last sql.NullString
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT MIN(sent_at), MAX(sent_at), COUNT(*) FROM otp_sends WHERE phone = ?`, o.Phone).Scan(&first, &last, &n); err != nil {
		return 0, err
	}
	// The gap is measured from the send log, not the code row, so using up or verifying a code doesn't reset it.
	if last.Valid {
		if wait := gap - now.Sub(parseTS(last.String)); wait > 0 {
			return wait, ErrTooSoon
		}
	}
	if n >= perHour {
		return max(parseTS(first.String).Add(time.Hour).Sub(now), time.Second), ErrTooMany
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO otp_sends (phone, sent_at) VALUES (?, ?)`, o.Phone, ts(now)); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO otp_codes (phone, code_hash, expires_at, attempts, sent_at) VALUES (?, ?, ?, 0, ?)
		ON CONFLICT (phone) DO UPDATE SET code_hash = excluded.code_hash, expires_at = excluded.expires_at, attempts = 0, sent_at = excluded.sent_at`,
		o.Phone, o.CodeHash, ts(o.ExpiresAt), ts(now)); err != nil {
		return 0, err
	}
	return 0, tx.Commit()
}

// DropOTP deletes the number's code (the send failed, or the provider says it is gone). The send log stays, so the
// hourly cap still counts it.
func (s *Store) DropOTP(ctx context.Context, phone string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM otp_codes WHERE phone = ?`, phone)
	return err
}

// TryOTP spends one try on the number's live code before the caller compares anything, so parallel guesses can't
// share a try. It returns the code with Attempts counting this try, or ok=false when there is no live code (none
// sent, expired, or maxTries already spent); a dead row is deleted.
func (s *Store) TryOTP(ctx context.Context, phone string, now time.Time, maxTries int) (OTP, bool, error) {
	o := OTP{Phone: phone}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return o, false, err
	}
	defer tx.Rollback()
	var expires, sent string
	err = tx.QueryRowContext(ctx, `SELECT code_hash, expires_at, attempts, sent_at FROM otp_codes WHERE phone = ?`, phone).Scan(&o.CodeHash, &expires, &o.Attempts, &sent)
	if errors.Is(err, sql.ErrNoRows) {
		return o, false, nil
	}
	if err != nil {
		return o, false, err
	}
	o.ExpiresAt, o.SentAt = parseTS(expires), parseTS(sent)
	if !now.Before(o.ExpiresAt) || o.Attempts >= maxTries {
		if _, err := tx.ExecContext(ctx, `DELETE FROM otp_codes WHERE phone = ?`, phone); err != nil {
			return o, false, err
		}
		return o, false, tx.Commit()
	}
	o.Attempts++
	if _, err := tx.ExecContext(ctx, `UPDATE otp_codes SET attempts = ? WHERE phone = ?`, o.Attempts, phone); err != nil {
		return o, false, err
	}
	return o, true, tx.Commit()
}

// ConsumeOTP deletes the code sent at sentAt so it can't be used twice. false when it is already gone: a concurrent
// verify used it, or a newer code replaced it.
func (s *Store) ConsumeOTP(ctx context.Context, phone string, sentAt time.Time) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM otp_codes WHERE phone = ? AND sent_at = ?`, phone, ts(sentAt))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// ---- deleting an account ----

// DeleteUser removes userID and everything they made, in one transaction: sessions, logs, predictions, follows both
// ways, likes, there-too taps, replies, top 5, watching, and sign-in codes for their phone; plus other people's
// likes, there-too taps and replies on this user's logs. Tables with no foreign key to users (follows.target_id,
// reactions and replies by log_id, the OTP tables) are cleared explicitly; the rest are too, so the result doesn't
// depend on the foreign_keys pragma.
func (s *Store) DeleteUser(ctx context.Context, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var phone sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT phone FROM users WHERE id = ?`, userID).Scan(&phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	myLogs := `(SELECT id FROM logs WHERE user_id = ?)`
	steps := []struct {
		q    string
		args []any
	}{
		{`DELETE FROM likes WHERE log_id IN ` + myLogs + ` OR user_id = ?`, []any{userID, userID}},
		{`DELETE FROM there_too WHERE log_id IN ` + myLogs + ` OR user_id = ?`, []any{userID, userID}},
		{`DELETE FROM replies WHERE log_id IN ` + myLogs + ` OR user_id = ?`, []any{userID, userID}},
		{`DELETE FROM follows WHERE user_id = ? OR target_id = ?`, []any{userID, userID}},
		{`DELETE FROM logs WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM predictions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM top5 WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM watching WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM sessions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM otp_codes WHERE phone = ?`, []any{phone.String}},
		{`DELETE FROM otp_sends WHERE phone = ?`, []any{phone.String}},
		{`DELETE FROM users WHERE id = ?`, []any{userID}},
	}
	for _, st := range steps {
		if _, err := tx.ExecContext(ctx, st.q, st.args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}
