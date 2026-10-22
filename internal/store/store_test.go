package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	// Re-opening must not re-apply migrations.
	s2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	s2.Close()
	return s
}

func TestStoreRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	u := domain.User{ID: "u-1", Name: "Guest", Handle: "guest", AvatarColour: "#000", FavouriteTeamID: "ind"}
	if err := s.CreateUser(ctx, u, "tok", now); err != nil {
		t.Fatal(err)
	}
	if got, ok, err := s.UserByToken(ctx, "tok"); err != nil || !ok || got.ID != "u-1" {
		t.Fatalf("by token: %v %v %v", got, ok, err)
	}
	if _, ok, _ := s.UserByToken(ctx, "nope"); ok {
		t.Fatal("unknown token resolved")
	}

	l := domain.MatchLog{ID: "log-a", UserID: "u-1", MatchID: "m1", Rating: 4, WatchedAt: domain.WatchHome, CreatedAt: now.Format(time.RFC3339)}
	saved, created, err := s.SaveLog(ctx, l)
	if err != nil || !created || saved.StubNo != 1 || saved.PlayerRatings == nil || saved.WatchedWith == nil {
		t.Fatalf("create: %+v %v %v", saved, created, err)
	}
	l2 := domain.MatchLog{ID: "log-b", UserID: "u-1", MatchID: "m2", Rating: 3, WatchedAt: domain.WatchStadium, CreatedAt: now.Format(time.RFC3339)}
	if saved, _, _ := s.SaveLog(ctx, l2); saved.StubNo != 2 {
		t.Fatalf("second stub = %d", saved.StubNo)
	}
	l.ID, l.Rating, l.PlayerRatings = "log-new", 2.5, map[string]int{"p": 9}
	saved, created, err = s.SaveLog(ctx, l)
	if err != nil || created || saved.ID != "log-a" || saved.StubNo != 1 {
		t.Fatalf("update: %+v %v %v", saved, created, err)
	}
	got, ok, _ := s.LogByID(ctx, "log-a")
	if !ok || got.Rating != 2.5 || got.PlayerRatings["p"] != 9 {
		t.Fatalf("reload: %+v", got)
	}

	on, _ := s.ToggleReaction(ctx, Likes, "sim-log", "u-1", now)
	off, _ := s.ToggleReaction(ctx, Likes, "sim-log", "u-1", now)
	if !on || off {
		t.Fatalf("toggle %v %v", on, off)
	}
	if f, _ := s.ToggleFollow(ctx, "u-1", "friend-1", now); !f {
		t.Fatal("follow")
	}
	if fs, _ := s.Follows(ctx, "u-1"); len(fs) != 1 {
		t.Fatalf("follows %v", fs)
	}
	if err := s.SetTop5(ctx, "u-1", []string{"m2", "m1"}); err != nil {
		t.Fatal(err)
	}
	if top, _ := s.Top5(ctx, "u-1"); len(top) != 2 || top[0] != "m2" {
		t.Fatalf("top5 %v", top)
	}
	place := domain.WatchCafe
	s.SetWatching(ctx, "u-1", "m3", &place, now)
	if w, _ := s.Watching(ctx, "m3"); w["u-1"] != domain.WatchCafe {
		t.Fatalf("watching %v", w)
	}
	s.SetWatching(ctx, "u-1", "m3", nil, now)
	if w, _ := s.Watching(ctx, "m3"); len(w) != 0 {
		t.Fatalf("cleared watching %v", w)
	}
	if ok, _ := s.DeleteLog(ctx, "someone-else", "log-a"); ok {
		t.Fatal("deleted someone else's log")
	}
	if ok, _ := s.DeleteLog(ctx, "u-1", "log-a"); !ok {
		t.Fatal("delete")
	}
	// Foreign keys are on: a log for an unknown user fails.
	if _, _, err := s.SaveLog(ctx, domain.MatchLog{ID: "x", UserID: "ghost", MatchID: "m", Rating: 1, WatchedAt: domain.WatchHome, CreatedAt: "t"}); err == nil {
		t.Fatal("expected FK error")
	}
}

func TestSeedHistoryUsesFreshIDs(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	now := time.Now()
	demo := []domain.MatchLog{
		{ID: "same", MatchID: "m1", Rating: 4, WatchedAt: domain.WatchHome, CreatedAt: "2026-01-02T00:00:00Z"},
		{ID: "same", MatchID: "m0", Rating: 3, WatchedAt: domain.WatchHome, CreatedAt: "2026-01-01T00:00:00Z"},
	}
	for _, id := range []string{"a", "b"} {
		s.CreateUser(ctx, domain.User{ID: id, Name: id, Handle: id}, "t"+id, now)
		if err := s.SeedHistory(ctx, id, demo, []domain.Prediction{{MatchID: "m1", WinnerID: "ind"}}, []string{"f1", "f2"}, now); err != nil {
			t.Fatal(err)
		}
	}
	logs, _ := s.Logs(ctx)
	if len(logs) != 4 {
		t.Fatalf("logs %d", len(logs))
	}
	preds, _ := s.Predictions(ctx)
	if len(preds) != 2 {
		t.Fatalf("preds %d", len(preds))
	}
}
