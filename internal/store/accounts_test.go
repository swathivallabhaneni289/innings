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

// TestMigrateGuestDB upgrades a database holding only 0001 and a guest (the shape of every local innings.db before
// sign-in) and checks the guest survives with no phone, not onboarded, and no way in.
func TestMigrateGuestDB(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "old.db")
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	schema, _ := migrationsFS.ReadFile("migrations/0001_init.sql")
	for _, q := range []string{
		string(schema),
		`CREATE TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`,
		`INSERT INTO schema_migrations VALUES ('0001_init', '2026-10-22T17:54:42Z')`,
		`INSERT INTO users (id, token, name, handle, avatar_colour, favourite_team_id, created_at)
			VALUES ('u-guest', 'old-guest-token', 'Guest Fan', 'fan-abc123', '#3B3F6B', 'ind', '2026-10-22T18:00:00Z')`,
	} {
		if _, err := raw.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	raw.Close()

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
	if err != nil || a.Phone != "" || a.Onboarded {
		t.Fatalf("guest account: %+v %v", a, err)
	}
	if _, ok, _ := s.UserBySession(ctx, "old-guest-token"); ok {
		t.Fatal("a guest token must not be a session")
	}
	if listed, _ := s.ListedUsers(ctx); len(listed) != 0 {
		t.Fatalf("guest is listed: %v", listed)
	}
}

func TestAccountsAndHandles(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	a := domain.User{ID: "u-a", AvatarColour: "#3B3F6B", FavouriteTeamID: "ind"}
	b := domain.User{ID: "u-b", AvatarColour: "#3B3F6B", FavouriteTeamID: "ind"}
	// Two blank handles are fine: only chosen handles are unique.
	if err := s.CreateAccount(ctx, a, "+919800000001", now); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateAccount(ctx, b, "+919800000002", now); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateAccount(ctx, domain.User{ID: "u-c"}, "+919800000001", now); !errors.Is(err, ErrPhoneTaken) {
		t.Fatalf("same phone twice: %v", err)
	}
	if u, ok, _ := s.UserByPhone(ctx, "+919800000001"); !ok || u.ID != "u-a" {
		t.Fatalf("by phone: %+v %v", u, ok)
	}
	acct, _ := s.Account(ctx, "u-a")
	if acct.Phone != "+919800000001" || acct.Onboarded || acct.CreatedAt != "2026-03-01T10:00:00Z" {
		t.Fatalf("account: %+v", acct)
	}

	a.Name, a.Handle = "Aditi", "Swathi_9"
	if err := s.UpdateUser(ctx, a); err != nil {
		t.Fatal(err)
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

	// Contact matching: only onboarded users, any number of hashes (chunked).
	hashes := []string{PhoneHash("+919800000002")}
	for i := 0; i < 1200; i++ {
		hashes = append(hashes, PhoneHash(fmt.Sprintf("+44700%06d", i)))
	}
	hashes = append(hashes, PhoneHash("+919800000001"))
	got, err := s.ListedUsersByPhoneHash(ctx, hashes)
	if err != nil || len(got) != 1 || got[0].ID != "u-a" {
		t.Fatalf("by phone hash: %v %v", got, err)
	}
	if got, err := s.ListedUsersByPhoneHash(ctx, nil); err != nil || got == nil || len(got) != 0 {
		t.Fatalf("no hashes: %v %v", got, err)
	}
}

func TestSessions(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Now()
	s.CreateAccount(ctx, domain.User{ID: "u-a"}, "+919800000001", now)
	s.CreateSession(ctx, "h1", "u-a", now)
	s.CreateSession(ctx, "h2", "u-a", now)
	if err := s.DeleteSession(ctx, "h1"); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.UserBySession(ctx, "h1"); ok {
		t.Fatal("deleted session still works")
	}
	if _, ok, _ := s.UserBySession(ctx, "h2"); !ok {
		t.Fatal("the other session was signed out too")
	}
}

func TestOTPLimitsAndTries(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	t0 := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	const phone = "+919800000001"
	start := func(at time.Time) (time.Duration, error) {
		return s.StartOTP(ctx, OTP{Phone: phone, CodeHash: "h", ExpiresAt: at.Add(10 * time.Minute), SentAt: at}, 30*time.Second, 5)
	}
	if _, err := start(t0); err != nil {
		t.Fatal(err)
	}
	if wait, err := start(t0.Add(10 * time.Second)); !errors.Is(err, ErrTooSoon) || wait != 20*time.Second {
		t.Fatalf("resend: %v %v", wait, err)
	}
	for i := 1; i < 5; i++ {
		if _, err := start(t0.Add(time.Duration(i) * time.Minute)); err != nil {
			t.Fatalf("send %d: %v", i+1, err)
		}
	}
	if wait, err := start(t0.Add(10 * time.Minute)); !errors.Is(err, ErrTooMany) || wait != 50*time.Minute {
		t.Fatalf("sixth send: %v %v", wait, err)
	}
	if _, err := start(t0.Add(time.Hour + time.Second)); err != nil {
		t.Fatalf("after the first send ages out: %v", err)
	}

	// Tries: each TryOTP spends one; after 5 the code is dead.
	sent := t0.Add(time.Hour + time.Second)
	for i := 1; i <= 5; i++ {
		o, ok, err := s.TryOTP(ctx, phone, sent, 5)
		if err != nil || !ok || o.Attempts != i || o.CodeHash != "h" || !o.SentAt.Equal(sent) {
			t.Fatalf("try %d: %+v %v %v", i, o, ok, err)
		}
	}
	if _, ok, _ := s.TryOTP(ctx, phone, sent, 5); ok {
		t.Fatal("sixth try allowed")
	}

	// Expiry and single use.
	next := sent.Add(time.Minute)
	if _, err := start(next); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.TryOTP(ctx, phone, next.Add(10*time.Minute), 5); ok {
		t.Fatal("expired code allowed")
	}
	later := next.Add(11 * time.Minute)
	if _, err := start(later); err != nil {
		t.Fatal(err)
	}
	o, ok, _ := s.TryOTP(ctx, phone, later, 5)
	if !ok {
		t.Fatal("fresh code refused")
	}
	if used, _ := s.ConsumeOTP(ctx, phone, o.SentAt); !used {
		t.Fatal("consume")
	}
	if used, _ := s.ConsumeOTP(ctx, phone, o.SentAt); used {
		t.Fatal("consumed twice")
	}
	if _, ok, _ := s.TryOTP(ctx, phone, later, 5); ok {
		t.Fatal("used code allowed")
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
	const phone = "+919800000001"
	s.CreateAccount(ctx, domain.User{ID: "gone"}, phone, now)
	s.CreateAccount(ctx, domain.User{ID: "stays"}, "+919800000002", now)
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
	s.StartOTP(ctx, OTP{Phone: phone, CodeHash: "h", ExpiresAt: now.Add(10 * time.Minute), SentAt: now}, 30*time.Second, 5)

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
		`SELECT COUNT(*) FROM otp_codes WHERE phone = '` + phone + `'`:                    0,
		`SELECT COUNT(*) FROM otp_sends WHERE phone = '` + phone + `'`:                    0,
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
}
