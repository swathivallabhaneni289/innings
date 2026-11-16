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

// Accounts, sessions, emailed codes and password resets (migrations/0002_accounts.sql and 0007_email_sign_in.sql,
// docs/API.md → Auth).

var (
	// ErrHandleTaken: another user already holds this @username (compared case-insensitively).
	ErrHandleTaken = errors.New("handle taken")
	// ErrEmailTaken: a confirmed account already has this email address.
	ErrEmailTaken = errors.New("email taken")
	// ErrTooSoon: a code went to this address less than the resend gap ago.
	ErrTooSoon = errors.New("code sent too recently")
	// ErrTooMany: this address has had its hourly allowance of codes.
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

// EmailHash is the SHA-256 hex of a normalised (trimmed, lower-cased) email address: users.email_hash, and what the
// app sends for contact matching.
func EmailHash(email string) string {
	sum := sha256.Sum256([]byte(email))
	return hex.EncodeToString(sum[:])
}

func parseTS(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// ---- accounts ----

// UnconfirmedTTL is how long an account whose email was never confirmed is kept. Nobody can sign in to one, so
// after a week it only holds an address its owner may not have meant to give us.
const UnconfirmedTTL = 7 * 24 * time.Hour

// Credentials is what signing in needs to know about an account. Phone-era accounts have no password hash.
type Credentials struct {
	UserID       string
	Email        string
	PasswordHash string
	Confirmed    bool // the emailed code was entered (or a password reset proved the inbox)
}

const credCols = `id, COALESCE(email, ''), COALESCE(password_hash, ''), email_verified_at IS NOT NULL`

func (s *Store) credentialsWhere(ctx context.Context, where string, arg any) (Credentials, bool, error) {
	var c Credentials
	err := s.db.QueryRowContext(ctx, `SELECT `+credCols+` FROM users WHERE `+where, arg).Scan(&c.UserID, &c.Email, &c.PasswordHash, &c.Confirmed)
	if errors.Is(err, sql.ErrNoRows) {
		return c, false, nil
	}
	return c, err == nil, err
}

// CredentialsByEmail looks an account up by its normalised email address.
func (s *Store) CredentialsByEmail(ctx context.Context, email string) (Credentials, bool, error) {
	return s.credentialsWhere(ctx, `email = ?`, email)
}

// CredentialsByID looks an account up by user id.
func (s *Store) CredentialsByID(ctx context.Context, userID string) (Credentials, bool, error) {
	return s.credentialsWhere(ctx, `id = ?`, userID)
}

// SignUp records a new account for email, unconfirmed, with u's starting profile, and returns its user id. If an
// unconfirmed account already has the address, its password is replaced and its id returned: only whoever reads the
// inbox can confirm either way, so the latest attempt wins. A confirmed account with the address is ErrEmailTaken.
// Unconfirmed accounts older than UnconfirmedTTL are cleared first. users.token gets "acct-"+id: the column is
// NOT NULL UNIQUE from the guest era, but sessions don't use it, so any unique value nobody can present will do.
func (s *Store) SignUp(ctx context.Context, u domain.User, email, passwordHash string, now time.Time) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE email IS NOT NULL AND email_verified_at IS NULL AND created_at <= ?`, ts(now.Add(-UnconfirmedTTL))); err != nil {
		return "", err
	}
	var id string
	var confirmed bool
	err = tx.QueryRowContext(ctx, `SELECT id, email_verified_at IS NOT NULL FROM users WHERE email = ?`, email).Scan(&id, &confirmed)
	switch {
	case err == nil && confirmed:
		return "", ErrEmailTaken
	case err == nil:
		if _, err := tx.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id); err != nil {
			return "", err
		}
	case errors.Is(err, sql.ErrNoRows):
		id = u.ID
		_, err = tx.ExecContext(ctx, `INSERT INTO users (id, token, name, handle, avatar_colour, favourite_team_id, team_ids, bio, photo_uri, created_at, email, email_hash, password_hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			u.ID, "acct-"+u.ID, u.Name, u.Handle, u.AvatarColour, u.FavouriteTeamID, teamsJSON(u), u.Bio, u.PhotoURI, ts(now), email, EmailHash(email), passwordHash)
		if isUniqueViolation(err) {
			return "", ErrEmailTaken
		}
		if err != nil {
			return "", err
		}
	default:
		return "", err
	}
	return id, tx.Commit()
}

// ConfirmEmail marks userID's address as confirmed; the first time is kept if it is called again.
func (s *Store) ConfirmEmail(ctx context.Context, userID string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET email_verified_at = COALESCE(email_verified_at, ?) WHERE id = ?`, ts(now), userID)
	return err
}

// SetPassword replaces userID's password hash and signs out every session except keep (a token hash; "" signs out
// all of them), in one transaction.
func (s *Store) SetPassword(ctx context.Context, userID, passwordHash, keep string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ? AND token_hash <> ?`, userID, keep); err != nil {
		return err
	}
	return tx.Commit()
}

// RehashPassword swaps userID's password hash for one with the current work factor, leaving sessions alone.
func (s *Store) RehashPassword(ctx context.Context, userID, passwordHash string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, userID)
	return err
}

// Account returns userID's private account details (email is "" for an account from before email sign-in).
func (s *Store) Account(ctx context.Context, userID string) (domain.Account, error) {
	var a domain.Account
	var email, onboarded sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT email, onboarded_at, created_at FROM users WHERE id = ?`, userID).Scan(&email, &onboarded, &a.CreatedAt)
	a.Email, a.Onboarded = email.String, onboarded.Valid
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

// hashChunk keeps each IN list well under SQLite's and Postgres's bound-parameter limits.
const hashChunk = 500

// ListedUsersByEmailHash returns the onboarded users with a confirmed address whose email hash is in hashes. It
// queries in chunks and keeps nothing.
func (s *Store) ListedUsersByEmailHash(ctx context.Context, hashes []string) ([]domain.User, error) {
	out := []domain.User{}
	for start := 0; start < len(hashes); start += hashChunk {
		chunk := hashes[start:min(start+hashChunk, len(hashes))]
		args := make([]any, len(chunk))
		for i, h := range chunk {
			args[i] = h
		}
		marks := strings.TrimSuffix(strings.Repeat("?, ", len(chunk)), ", ")
		us, err := s.queryUsers(ctx, `SELECT `+userCols+` FROM users
			WHERE onboarded_at IS NOT NULL AND email_verified_at IS NOT NULL AND email_hash IN (`+marks+`)`, args...)
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

// UserBySession resolves a session's token hash to its user. Sessions of accounts with no password (phone-era
// accounts, which can no longer sign in) resolve to nothing, so those devices are sent back to sign up.
func (s *Store) UserBySession(ctx context.Context, tokenHash string) (domain.User, bool, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, `SELECT u.id, u.name, u.handle, u.avatar_colour, u.favourite_team_id, u.team_ids, u.bio, u.photo_uri
		FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token_hash = ? AND u.password_hash IS NOT NULL`, tokenHash))
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

// ---- emailed codes ----

// EmailCode is the live code for one address and purpose ("verify" or "reset"). CodeHash is the SHA-256 hex of the
// 6 digits. SentAt identifies the code: a new code always has a later SentAt.
type EmailCode struct {
	Email     string
	Purpose   string
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
	SentAt    time.Time
}

// StartCode makes c the live code for its address and purpose if the resend gap and the hourly cap allow, logging
// the send. Both limits count every code made for the address, whatever its purpose. Checking and writing in one
// transaction means concurrent starts can't both slip under a limit (SQLite runs it serially on our single
// connection). On refusal it returns ErrTooSoon or ErrTooMany and how long to wait. It also prunes expired codes and
// sends older than an hour, for every address, so addresses aren't kept longer than the limits need them.
func (s *Store) StartCode(ctx context.Context, c EmailCode, gap time.Duration, perHour int) (time.Duration, error) {
	now := c.SentAt
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM email_sends WHERE sent_at <= ?`, ts(now.Add(-time.Hour))); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM email_codes WHERE expires_at <= ?`, ts(now)); err != nil {
		return 0, err
	}
	var first, last sql.NullString
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT MIN(sent_at), MAX(sent_at), COUNT(*) FROM email_sends WHERE email = ?`, c.Email).Scan(&first, &last, &n); err != nil {
		return 0, err
	}
	// The gap is measured from the send log, not the code row, so using up a code doesn't reset it.
	if last.Valid {
		if wait := gap - now.Sub(parseTS(last.String)); wait > 0 {
			return wait, ErrTooSoon
		}
	}
	if n >= perHour {
		return max(parseTS(first.String).Add(time.Hour).Sub(now), time.Second), ErrTooMany
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO email_sends (email, sent_at) VALUES (?, ?)`, c.Email, ts(now)); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO email_codes (email, purpose, code_hash, expires_at, attempts, sent_at) VALUES (?, ?, ?, ?, 0, ?)
		ON CONFLICT (email, purpose) DO UPDATE SET code_hash = excluded.code_hash, expires_at = excluded.expires_at, attempts = 0, sent_at = excluded.sent_at`,
		c.Email, c.Purpose, c.CodeHash, ts(c.ExpiresAt), ts(now)); err != nil {
		return 0, err
	}
	return 0, tx.Commit()
}

// DropCode deletes the address's code for purpose (its email couldn't be sent). The send log stays, so the hourly
// cap still counts it.
func (s *Store) DropCode(ctx context.Context, email, purpose string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM email_codes WHERE email = ? AND purpose = ?`, email, purpose)
	return err
}

// TryCode spends one try on the address's live code for purpose before the caller compares anything, so parallel
// guesses can't share a try. It returns the code with Attempts counting this try, or ok=false when there is no live
// code (none sent, expired, or maxTries already spent); a dead row is deleted.
func (s *Store) TryCode(ctx context.Context, email, purpose string, now time.Time, maxTries int) (EmailCode, bool, error) {
	c := EmailCode{Email: email, Purpose: purpose}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return c, false, err
	}
	defer tx.Rollback()
	var expires, sent string
	err = tx.QueryRowContext(ctx, `SELECT code_hash, expires_at, attempts, sent_at FROM email_codes WHERE email = ? AND purpose = ?`, email, purpose).
		Scan(&c.CodeHash, &expires, &c.Attempts, &sent)
	if errors.Is(err, sql.ErrNoRows) {
		return c, false, nil
	}
	if err != nil {
		return c, false, err
	}
	c.ExpiresAt, c.SentAt = parseTS(expires), parseTS(sent)
	if !now.Before(c.ExpiresAt) || c.Attempts >= maxTries {
		if _, err := tx.ExecContext(ctx, `DELETE FROM email_codes WHERE email = ? AND purpose = ?`, email, purpose); err != nil {
			return c, false, err
		}
		return c, false, tx.Commit()
	}
	c.Attempts++
	if _, err := tx.ExecContext(ctx, `UPDATE email_codes SET attempts = ? WHERE email = ? AND purpose = ?`, c.Attempts, email, purpose); err != nil {
		return c, false, err
	}
	return c, true, tx.Commit()
}

// ConsumeCode deletes the code sent at sentAt so it can't be used twice. false when it is already gone: a concurrent
// check used it, or a newer code replaced it.
func (s *Store) ConsumeCode(ctx context.Context, email, purpose string, sentAt time.Time) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM email_codes WHERE email = ? AND purpose = ? AND sent_at = ?`, email, purpose, ts(sentAt))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// ---- password resets ----

// CreateResetToken stores a reset token's hash for userID until expires, pruning expired tokens.
func (s *Store) CreateResetToken(ctx context.Context, tokenHash, userID string, now, expires time.Time) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM reset_tokens WHERE expires_at <= ?`, ts(now)); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO reset_tokens (token_hash, user_id, expires_at) VALUES (?, ?, ?)`, tokenHash, userID, ts(expires))
	return err
}

// ResetTokenUser returns the account a live reset token belongs to, without using the token up.
func (s *Store) ResetTokenUser(ctx context.Context, tokenHash string, now time.Time) (Credentials, bool, error) {
	var userID string
	err := s.db.QueryRowContext(ctx, `SELECT user_id FROM reset_tokens WHERE token_hash = ? AND expires_at > ?`, tokenHash, ts(now)).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Credentials{}, false, nil
	}
	if err != nil {
		return Credentials{}, false, err
	}
	return s.CredentialsByID(ctx, userID)
}

// ResetPassword spends a live reset token on a new password hash, in one transaction: every reset token for the
// account goes, the address counts as confirmed (the code proved the inbox), and every session is signed out. It
// returns the account's user id, or ok=false when the token is unknown, expired or already used.
func (s *Store) ResetPassword(ctx context.Context, tokenHash, passwordHash string, now time.Time) (string, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()
	var userID string
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM reset_tokens WHERE token_hash = ? AND expires_at > ?`, tokenHash, ts(now)).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	for _, st := range []struct {
		q    string
		args []any
	}{
		{`DELETE FROM reset_tokens WHERE user_id = ?`, []any{userID}},
		{`UPDATE users SET password_hash = ?, email_verified_at = COALESCE(email_verified_at, ?) WHERE id = ?`, []any{passwordHash, ts(now), userID}},
		{`DELETE FROM sessions WHERE user_id = ?`, []any{userID}},
	} {
		if _, err := tx.ExecContext(ctx, st.q, st.args...); err != nil {
			return "", false, err
		}
	}
	return userID, true, tx.Commit()
}

// ---- deleting an account ----

// DeleteUser removes userID and everything they made, in one transaction: sessions, logs, predictions, follows both
// ways, likes, there-too taps, replies, chat messages, call picks, top 5, watching, reset tokens, and the codes and
// send log for their address (or, for a phone-era account, their number); plus other people's likes, there-too taps
// and replies on this user's logs. Tables with no foreign key to users (follows.target_id, reactions and replies by
// log_id, the code tables) are cleared explicitly; the rest are too, so the result doesn't depend on the
// foreign_keys pragma.
func (s *Store) DeleteUser(ctx context.Context, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var email, phone sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT email, phone FROM users WHERE id = ?`, userID).Scan(&email, &phone)
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
		{`DELETE FROM chat_messages WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM call_picks WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM follows WHERE user_id = ? OR target_id = ?`, []any{userID, userID}},
		{`DELETE FROM logs WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM predictions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM top5 WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM watching WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM sessions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM reset_tokens WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM email_codes WHERE email = ?`, []any{email.String}},
		{`DELETE FROM email_sends WHERE email = ?`, []any{email.String}},
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
