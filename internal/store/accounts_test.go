package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// oldDB writes a database migrated up to and including upTo, then runs extra statements on it: the shape of a
// local innings.db from before a later migration.
func oldDB(t *testing.T, upTo string, extra ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "old.db")
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err := raw.Exec(`CREATE TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"0001_init", "0002_accounts", "0003_feed", "0004_chat", "0005_call_picks", "0006_chat_calls"} {
		schema, err := migrationsFS.ReadFile("migrations/" + name + ".sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := raw.Exec(string(schema)); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		raw.Exec(`INSERT INTO schema_migrations VALUES (?, '2026-10-22T17:54:42Z')`, name)
		if name == upTo {
			break
		}
	}
	for _, q := range extra {
		if _, err := raw.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

// TestMigrateGuestDB upgrades a database holding only 0001 and a guest (the shape of every local innings.db before
// sign-in) and checks the guest survives with no email, not onboarded, and no way in.
func TestMigrateGuestDB(t *testing.T) {
	ctx := context.Background()
	path := oldDB(t, "0001_init", `INSERT INTO users (id, token, name, handle, avatar_colour, favourite_team_id, created_at)
		VALUES ('u-guest', 'old-guest-token', 'Guest Fan', 'fan-abc123', '#3B3F6B', 'ind', '2026-10-22T18:00:00Z')`)
	s, err := Open(path)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	defer s.Close()
	u, ok, err := s.UserByID(ctx, "u-guest")
	if err != nil || !ok || u.Handle != "fan-abc123" {
		t.Fatalf("guest after migrate: %+v %v %v", u, ok, err)
	}
	a, err := s.Account(ctx, "u-guest")
	if err != nil || a.Email != "" || a.Onboarded {
		t.Fatalf("guest account: %+v %v", a, err)
	}
	if _, ok, _ := s.UserBySession(ctx, "old-guest-token"); ok {
		t.Fatal("a guest token must not be a session")
	}
	if listed, _ := s.ListedUsers(ctx); len(listed) != 0 {
		t.Fatalf("guest is listed: %v", listed)
	}
}

// TestMigratePhoneDB upgrades a database from the phone-sign-in era. Everything in it stays (the migration only
// adds), but the phone account's session no longer signs anyone in.
func TestMigratePhoneDB(t *testing.T) {
	ctx := context.Background()
	path := oldDB(t, "0006_chat_calls",
		`INSERT INTO users (id, token, name, handle, avatar_colour, favourite_team_id, created_at, phone, phone_hash, onboarded_at)
			VALUES ('u-phone', 'acct-u-phone', 'Phone Fan', 'phonefan', '#3B3F6B', 'ind', '2026-10-25T10:00:00Z', '+919800000001', 'h', '2026-10-25T10:01:00Z')`,
		`INSERT INTO sessions (token_hash, user_id, created_at) VALUES ('phone-session', 'u-phone', '2026-10-25T10:00:00Z')`,
		`INSERT INTO logs (id, user_id, match_id, rating, watched_at, created_at, stub_no) VALUES ('l1', 'u-phone', 'm1', 4, 'home', '2026-10-25T11:00:00Z', 1)`,
		`INSERT INTO otp_sends (phone, sent_at) VALUES ('+919800000001', '2026-10-25T10:00:00Z')`)
	s, err := Open(path)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	defer s.Close()
	if _, ok, _ := s.UserByID(ctx, "u-phone"); !ok {
		t.Fatal("phone account lost")
	}
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM logs WHERE user_id = 'u-phone'`:     1,
		`SELECT COUNT(*) FROM users WHERE phone IS NOT NULL`:      1,
		`SELECT COUNT(*) FROM otp_sends`:                          1,
		`SELECT COUNT(*) FROM sessions WHERE user_id = 'u-phone'`: 1,
	} {
		if got := count(t, s, q); got != want {
			t.Errorf("%s = %d, want %d", q, got, want)
		}
	}
	if _, ok, _ := s.UserBySession(ctx, "phone-session"); ok {
		t.Fatal("a phone account's session must not sign anyone in")
	}
	if a, _ := s.Account(ctx, "u-phone"); a.Email != "" || !a.Onboarded {
		t.Fatalf("phone account: %+v", a)
	}
	if u, _, _ := s.UserByID(ctx, "u-phone"); len(u.TeamIDs) != 1 || u.TeamIDs[0] != "ind" {
		t.Fatalf("an existing fan backs their favourite team: %+v", u)
	}
}

func TestSignUpAndCredentials(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	fresh := func(id string) domain.User {
		return domain.User{ID: id, AvatarColour: "#3B3F6B", FavouriteTeamID: "ind"}
	}

	id, err := s.SignUp(ctx, fresh("u-a"), "a@example.com", "hash-1", now)
	if err != nil || id != "u-a" {
		t.Fatalf("sign up: %q %v", id, err)
	}
	c, ok, _ := s.CredentialsByEmail(ctx, "a@example.com")
	if !ok || c.UserID != "u-a" || c.PasswordHash != "hash-1" || c.Confirmed || c.Email != "a@example.com" {
		t.Fatalf("credentials: %+v %v", c, ok)
	}
	if _, ok, _ := s.CredentialsByEmail(ctx, "nobody@example.com"); ok {
		t.Fatal("unknown address found")
	}

	// Unconfirmed: a second sign-up for the address keeps the account and replaces the password.
	if id, err := s.SignUp(ctx, fresh("u-other"), "a@example.com", "hash-2", now.Add(time.Minute)); err != nil || id != "u-a" {
		t.Fatalf("take over unconfirmed: %q %v", id, err)
	}
	if c, _, _ := s.CredentialsByID(ctx, "u-a"); c.PasswordHash != "hash-2" {
		t.Fatalf("password not replaced: %+v", c)
	}
	if _, ok, _ := s.UserByID(ctx, "u-other"); ok {
		t.Fatal("a second account was made for the same address")
	}

	// Confirmed: the address is taken.
	if err := s.ConfirmEmail(ctx, "u-a", now); err != nil {
		t.Fatal(err)
	}
	s.ConfirmEmail(ctx, "u-a", now.Add(time.Hour)) // keeps the first time
	if _, err := s.SignUp(ctx, fresh("u-x"), "a@example.com", "hash-3", now); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("confirmed address: %v", err)
	}
	if c, _, _ := s.CredentialsByID(ctx, "u-a"); !c.Confirmed || c.PasswordHash != "hash-2" {
		t.Fatalf("after a refused sign-up: %+v", c)
	}
	acct, _ := s.Account(ctx, "u-a")
	if acct.Email != "a@example.com" || acct.Onboarded || acct.CreatedAt != "2026-03-01T10:00:00Z" {
		t.Fatalf("account: %+v", acct)
	}

	// Unconfirmed accounts older than a week are cleared by the next sign-up; confirmed ones never are.
	s.SignUp(ctx, fresh("u-stale"), "stale@example.com", "h", now)
	later := now.Add(UnconfirmedTTL + time.Second)
	s.SignUp(ctx, fresh("u-new"), "new@example.com", "h", later)
	if _, ok, _ := s.UserByID(ctx, "u-stale"); ok {
		t.Fatal("stale unconfirmed account kept")
	}
	if _, ok, _ := s.UserByID(ctx, "u-a"); !ok {
		t.Fatal("confirmed account cleared")
	}
	if _, ok, _ := s.UserByID(ctx, "u-new"); !ok {
		t.Fatal("new account missing")
	}
}

func TestHandlesAndListing(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	a := domain.User{ID: "u-a", AvatarColour: "#3B3F6B", FavouriteTeamID: "ind"}
	b := domain.User{ID: "u-b", AvatarColour: "#3B3F6B", FavouriteTeamID: "ind"}
	// Two blank handles are fine: only chosen handles are unique.
	if _, err := s.SignUp(ctx, a, "a@example.com", "h", now); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SignUp(ctx, b, "b@example.com", "h", now); err != nil {
		t.Fatal(err)
	}

	a.Name, a.Handle = "Aditi", "Swathi_9"
	a.FavouriteTeamID, a.TeamIDs = "rcb", []string{"rcb", "ind"}
	if err := s.UpdateUser(ctx, a); err != nil {
		t.Fatal(err)
	}
	if u, _, _ := s.UserByID(ctx, "u-a"); u.FavouriteTeamID != "rcb" || len(u.TeamIDs) != 2 || u.TeamIDs[1] != "ind" {
		t.Fatalf("teams round trip: %+v", u)
	}
	if u, _, _ := s.UserByID(ctx, "u-b"); len(u.TeamIDs) != 1 || u.TeamIDs[0] != "ind" {
		t.Fatalf("a new account backs its favourite team: %+v", u)
	}
	if taken, _ := s.HandleTaken(ctx, "swathi_9", "u-b"); !taken {
		t.Fatal("handle should be taken for b, ignoring case")
	}
	if taken, _ := s.HandleTaken(ctx, "swathi_9", "u-a"); taken {
		t.Fatal("my own handle isn't taken for me")
	}
	b.Handle = "swathi_9" // the unique index catches what a check-then-write race would miss
	if err := s.UpdateUser(ctx, b); !errors.Is(err, ErrHandleTaken) {
		t.Fatalf("duplicate handle: %v", err)
	}

	if listed, _ := s.ListedUsers(ctx); len(listed) != 0 {
		t.Fatalf("nobody is onboarded yet: %v", listed)
	}
	s.ConfirmEmail(ctx, "u-a", now)
	if err := s.SetOnboarded(ctx, "u-a", now); err != nil {
		t.Fatal(err)
	}
	s.SetOnboarded(ctx, "u-a", now.Add(time.Hour)) // keeps the first time
	if acct, _ := s.Account(ctx, "u-a"); !acct.Onboarded {
		t.Fatal("not onboarded")
	}
	if listed, _ := s.ListedUsers(ctx); len(listed) != 1 || listed[0].ID != "u-a" {
		t.Fatalf("listed: %v", listed)
	}

	// Contact matching: only onboarded users with a confirmed address, any number of hashes (chunked).
	hashes := []string{EmailHash("b@example.com")}
	for i := 0; i < 1200; i++ {
		hashes = append(hashes, EmailHash(fmt.Sprintf("fan%d@example.com", i)))
	}
	hashes = append(hashes, EmailHash("a@example.com"))
	got, err := s.ListedUsersByEmailHash(ctx, hashes)
	if err != nil || len(got) != 1 || got[0].ID != "u-a" {
		t.Fatalf("by email hash: %v %v", got, err)
	}
	if got, err := s.ListedUsersByEmailHash(ctx, nil); err != nil || got == nil || len(got) != 0 {
		t.Fatalf("no hashes: %v %v", got, err)
	}
}

func TestSessionsAndPasswords(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Now()
	s.SignUp(ctx, domain.User{ID: "u-a"}, "a@example.com", "old", now)
	for _, h := range []string{"h1", "h2", "h3"} {
		s.CreateSession(ctx, h, "u-a", now)
	}
	if err := s.DeleteSession(ctx, "h1"); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.UserBySession(ctx, "h1"); ok {
		t.Fatal("deleted session still works")
	}
	if _, ok, _ := s.UserBySession(ctx, "h2"); !ok {
		t.Fatal("the other session was signed out too")
	}
	// A new password keeps only the session that changed it.
	if err := s.SetPassword(ctx, "u-a", "new", "h2"); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.UserBySession(ctx, "h3"); ok {
		t.Fatal("another device stayed signed in after a password change")
	}
	if _, ok, _ := s.UserBySession(ctx, "h2"); !ok {
		t.Fatal("the device that changed the password was signed out")
	}
	if c, _, _ := s.CredentialsByID(ctx, "u-a"); c.PasswordHash != "new" {
		t.Fatalf("password: %+v", c)
	}
	// A rehash leaves sessions alone; "" as keep signs out everything.
	s.RehashPassword(ctx, "u-a", "stronger")
	if _, ok, _ := s.UserBySession(ctx, "h2"); !ok {
		t.Fatal("rehash signed a device out")
	}
	s.SetPassword(ctx, "u-a", "newer", "")
	if n := count(t, s, `SELECT COUNT(*) FROM sessions WHERE user_id = 'u-a'`); n != 0 {
		t.Fatalf("%d sessions left", n)
	}
}

func TestCodeLimitsAndTries(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	t0 := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	const email = "a@example.com"
	start := func(purpose string, at time.Time) (time.Duration, error) {
		return s.StartCode(ctx, EmailCode{Email: email, Purpose: purpose, CodeHash: "h-" + purpose, ExpiresAt: at.Add(10 * time.Minute), SentAt: at}, 30*time.Second, 5)
	}
	if _, err := start("verify", t0); err != nil {
		t.Fatal(err)
	}
	// The gap counts every code for the address, whatever it's for.
	if wait, err := start("reset", t0.Add(10*time.Second)); !errors.Is(err, ErrTooSoon) || wait != 20*time.Second {
		t.Fatalf("resend: %v %v", wait, err)
	}
	for i := 1; i < 5; i++ {
		if _, err := start("reset", t0.Add(time.Duration(i)*time.Minute)); err != nil {
			t.Fatalf("send %d: %v", i+1, err)
		}
	}
	if wait, err := start("verify", t0.Add(10*time.Minute)); !errors.Is(err, ErrTooMany) || wait != 50*time.Minute {
		t.Fatalf("sixth send: %v %v", wait, err)
	}
	// The two purposes keep separate codes.
	c, ok, _ := s.TryCode(ctx, email, "verify", t0.Add(5*time.Minute), 5)
	if !ok || c.CodeHash != "h-verify" || c.Attempts != 1 {
		t.Fatalf("verify code: %+v %v", c, ok)
	}
	if c, ok, _ := s.TryCode(ctx, email, "reset", t0.Add(5*time.Minute), 5); !ok || c.CodeHash != "h-reset" {
		t.Fatalf("reset code: %+v %v", c, ok)
	}
	if _, err := s.StartCode(ctx, EmailCode{Email: "b@example.com", Purpose: "verify", CodeHash: "x", ExpiresAt: t0.Add(time.Hour), SentAt: t0.Add(10 * time.Minute)}, 30*time.Second, 5); err != nil {
		t.Fatalf("another address is unaffected: %v", err)
	}

	// Tries: each TryCode spends one; after 5 the code is dead.
	sent := t0.Add(time.Hour + time.Second)
	if _, err := start("verify", sent); err != nil {
		t.Fatalf("after the first send ages out: %v", err)
	}
	for i := 1; i <= 5; i++ {
		c, ok, err := s.TryCode(ctx, email, "verify", sent, 5)
		if err != nil || !ok || c.Attempts != i || !c.SentAt.Equal(sent) {
			t.Fatalf("try %d: %+v %v %v", i, c, ok, err)
		}
	}
	if _, ok, _ := s.TryCode(ctx, email, "verify", sent, 5); ok {
		t.Fatal("sixth try allowed")
	}

	// Expiry and single use.
	next := sent.Add(time.Minute)
	start("verify", next)
	if _, ok, _ := s.TryCode(ctx, email, "verify", next.Add(10*time.Minute), 5); ok {
		t.Fatal("expired code allowed")
	}
	later := next.Add(11 * time.Minute)
	start("verify", later)
	c, ok, _ = s.TryCode(ctx, email, "verify", later, 5)
	if !ok {
		t.Fatal("fresh code refused")
	}
	if used, _ := s.ConsumeCode(ctx, email, "verify", c.SentAt); !used {
		t.Fatal("consume")
	}
	if used, _ := s.ConsumeCode(ctx, email, "verify", c.SentAt); used {
		t.Fatal("consumed twice")
	}
	if _, ok, _ := s.TryCode(ctx, email, "verify", later, 5); ok {
		t.Fatal("used code allowed")
	}
	// A dropped code is gone, but its send still counts towards the cap.
	start("reset", later.Add(time.Minute))
	s.DropCode(ctx, email, "reset")
	if _, ok, _ := s.TryCode(ctx, email, "reset", later.Add(time.Minute), 5); ok {
		t.Fatal("dropped code allowed")
	}
	if n := count(t, s, `SELECT COUNT(*) FROM email_sends WHERE email = ?`, email); n != 4 {
		t.Fatalf("send log: %d", n)
	}
}

func TestResetTokens(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	s.SignUp(ctx, domain.User{ID: "u-a"}, "a@example.com", "old", now) // never confirmed
	s.CreateSession(ctx, "h1", "u-a", now)
	s.CreateSession(ctx, "h2", "u-a", now)
	if err := s.CreateResetToken(ctx, "t1", "u-a", now, now.Add(15*time.Minute)); err != nil {
		t.Fatal(err)
	}
	s.CreateResetToken(ctx, "t2", "u-a", now, now.Add(15*time.Minute))

	if c, ok, _ := s.ResetTokenUser(ctx, "t1", now.Add(time.Minute)); !ok || c.UserID != "u-a" || c.Email != "a@example.com" {
		t.Fatalf("peek: %+v %v", c, ok)
	}
	if _, ok, _ := s.ResetTokenUser(ctx, "t1", now.Add(15*time.Minute)); ok {
		t.Fatal("expired token allowed")
	}
	if _, ok, _ := s.ResetPassword(ctx, "t1", "new", now.Add(15*time.Minute)); ok {
		t.Fatal("expired token reset the password")
	}
	userID, ok, err := s.ResetPassword(ctx, "t1", "new", now.Add(time.Minute))
	if err != nil || !ok || userID != "u-a" {
		t.Fatalf("reset: %q %v %v", userID, ok, err)
	}
	c, _, _ := s.CredentialsByID(ctx, "u-a")
	if c.PasswordHash != "new" || !c.Confirmed {
		t.Fatalf("after reset: %+v", c)
	}
	if n := count(t, s, `SELECT COUNT(*) FROM sessions WHERE user_id = 'u-a'`); n != 0 {
		t.Fatalf("sessions after reset: %d", n)
	}
	// Every token for the account is spent, not just the one used.
	for _, tok := range []string{"t1", "t2"} {
		if _, ok, _ := s.ResetPassword(ctx, tok, "again", now.Add(2*time.Minute)); ok {
			t.Fatalf("token %s worked twice", tok)
		}
	}
	if _, ok, _ := s.ResetTokenUser(ctx, "nope", now); ok {
		t.Fatal("unknown token")
	}
}

func count(t *testing.T, s *Store, q string, args ...any) int {
	t.Helper()
	var n int
	if err := s.db.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	const email = "gone@example.com"
	s.SignUp(ctx, domain.User{ID: "gone"}, email, "h", now)
	s.SignUp(ctx, domain.User{ID: "stays"}, "stays@example.com", "h", now)
	s.CreateSession(ctx, "h-gone", "gone", now)
	s.CreateSession(ctx, "h-stays", "stays", now)
	mine, _, _ := s.SaveLog(ctx, domain.MatchLog{ID: "log-gone", UserID: "gone", MatchID: "m1", Rating: 4, WatchedAt: domain.WatchHome, CreatedAt: ts(now)})
	theirs, _, _ := s.SaveLog(ctx, domain.MatchLog{ID: "log-stays", UserID: "stays", MatchID: "m1", Rating: 3, WatchedAt: domain.WatchHome, CreatedAt: ts(now)})
	s.SavePrediction(ctx, domain.Prediction{UserID: "gone", MatchID: "m1", WinnerID: "ind", CreatedAt: ts(now)})
	s.ToggleFollow(ctx, "gone", "stays", now)
	s.ToggleFollow(ctx, "stays", "gone", now)
	s.ToggleFollow(ctx, "stays", "f-aditi", now)
	for _, tbl := range []string{Likes, ThereToo} {
		s.ToggleReaction(ctx, tbl, theirs.ID, "gone", now) // mine on theirs
		s.ToggleReaction(ctx, tbl, mine.ID, "stays", now)  // theirs on mine
		s.ToggleReaction(ctx, tbl, theirs.ID, "stays", now)
	}
	s.AddReply(ctx, domain.Reply{ID: "r1", LogID: theirs.ID, UserID: "gone", Text: "hi", CreatedAt: ts(now)})
	s.AddReply(ctx, domain.Reply{ID: "r2", LogID: mine.ID, UserID: "stays", Text: "hi", CreatedAt: ts(now)})
	s.AddReply(ctx, domain.Reply{ID: "r3", LogID: theirs.ID, UserID: "stays", Text: "hi", CreatedAt: ts(now)})
	s.SetTop5(ctx, "gone", []string{"m1"})
	place := domain.WatchCafe
	s.SetWatching(ctx, "gone", "m2", &place, now)
	s.StartCode(ctx, EmailCode{Email: email, Purpose: "reset", CodeHash: "h", ExpiresAt: now.Add(10 * time.Minute), SentAt: now}, 30*time.Second, 5)
	s.CreateResetToken(ctx, "tok", "gone", now, now.Add(time.Hour))

	if err := s.DeleteUser(ctx, "gone"); err != nil {
		t.Fatal(err)
	}
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM users WHERE id = 'gone'`:                                    0,
		`SELECT COUNT(*) FROM sessions WHERE user_id = 'gone'`:                            0,
		`SELECT COUNT(*) FROM logs WHERE user_id = 'gone'`:                                0,
		`SELECT COUNT(*) FROM predictions WHERE user_id = 'gone'`:                         0,
		`SELECT COUNT(*) FROM follows WHERE user_id = 'gone' OR target_id = 'gone'`:       0,
		`SELECT COUNT(*) FROM likes WHERE user_id = 'gone' OR log_id = 'log-gone'`:        0,
		`SELECT COUNT(*) FROM there_too WHERE user_id = 'gone' OR log_id = 'log-gone'`:    0,
		`SELECT COUNT(*) FROM replies WHERE user_id = 'gone' OR log_id = 'log-gone'`:      0,
		`SELECT COUNT(*) FROM top5 WHERE user_id = 'gone'`:                                0,
		`SELECT COUNT(*) FROM watching WHERE user_id = 'gone'`:                            0,
		`SELECT COUNT(*) FROM reset_tokens WHERE user_id = 'gone'`:                        0,
		`SELECT COUNT(*) FROM email_codes WHERE email = '` + email + `'`:                  0,
		`SELECT COUNT(*) FROM email_sends WHERE email = '` + email + `'`:                  0,
		`SELECT COUNT(*) FROM users WHERE id = 'stays'`:                                   1,
		`SELECT COUNT(*) FROM sessions WHERE user_id = 'stays'`:                           1,
		`SELECT COUNT(*) FROM logs WHERE user_id = 'stays'`:                               1,
		`SELECT COUNT(*) FROM follows WHERE user_id = 'stays'`:                            1,
		`SELECT COUNT(*) FROM likes WHERE user_id = 'stays' AND log_id = 'log-stays'`:     1,
		`SELECT COUNT(*) FROM there_too WHERE user_id = 'stays' AND log_id = 'log-stays'`: 1,
		`SELECT COUNT(*) FROM replies WHERE id = 'r3'`:                                    1,
	} {
		if got := count(t, s, q); got != want {
			t.Errorf("%s = %d, want %d", q, got, want)
		}
	}
	if err := s.DeleteUser(ctx, "gone"); err != nil {
		t.Fatalf("deleting twice: %v", err)
	}
	// The address is free again.
	if _, err := s.SignUp(ctx, domain.User{ID: "back"}, email, "h", now); err != nil {
		t.Fatalf("sign up again: %v", err)
	}
}
