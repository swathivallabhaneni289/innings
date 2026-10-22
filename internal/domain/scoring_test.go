package domain

import (
	"testing"
	"time"
)

func completed(id, start string, winner *string, scorers, wickets []string) Match {
	return Match{
		ID: id, Kind: "fixture", TeamIDs: [2]string{"ind", "aus"}, StartTime: start, State: Completed,
		Result: &MatchResult{WinnerID: winner, TopRunScorerIDs: scorers, TopWicketTakerIDs: wickets, Innings: []InningsScore{}},
	}
}

func TestIsLocked(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		state MatchState
		start time.Time
		want  bool
	}{
		{"upcoming future", Upcoming, now.Add(time.Minute), false},
		{"upcoming at first ball", Upcoming, now, true},
		{"upcoming past start", Upcoming, now.Add(-time.Minute), true},
		{"live", Live, now.Add(time.Hour), true},
		{"completed", Completed, now.Add(time.Hour), true},
	}
	for _, c := range cases {
		m := Match{State: c.state, StartTime: c.start.Format(time.RFC3339)}
		if got := IsLocked(m, now); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestStreakMultiplier(t *testing.T) {
	for streak, want := range map[int]float64{0: 1, 1: 1, 2: 1, 3: 1.5, 4: 1.5, 5: 2, 9: 2} {
		if got := StreakMultiplier(streak); got != want {
			t.Errorf("streak %d: got %v want %v", streak, got, want)
		}
	}
}

func TestScoreOne(t *testing.T) {
	ind := Ptr("ind")
	m := completed("m1", "2026-01-01T10:00:00Z", ind, []string{"a", "b"}, []string{"w"})
	pred := func(win, bat, bowl string) Prediction {
		return Prediction{UserID: "u", MatchID: "m1", WinnerID: win, TopRunScorerID: bat, TopWicketTakerID: bowl}
	}
	cases := []struct {
		name         string
		p            Prediction
		m            Match
		streakBefore int
		share        float64
		known        bool
		wantBase     int
		wantPoints   int
		wantStreak   int
		wantUpset    bool
		wantSettled  bool
	}{
		{"all right", pred("ind", "a", "w"), m, 0, 0.6, true, 40, 40, 1, false, true},
		{"tied top scorer counts", pred("ind", "b", "x"), m, 0, 0.6, true, 25, 25, 1, false, true},
		{"wrong winner keeps player points", pred("aus", "a", "w"), m, 4, 0.6, true, 30, 30, 0, false, true},
		{"upset bonus", pred("ind", "x", "x"), m, 0, 0.2, true, 20, 20, 1, true, true},
		{"no picks no upset", pred("ind", "x", "x"), m, 0, 0, false, 10, 10, 1, false, true},
		{"streak 3 ×1.5", pred("ind", "a", "x"), m, 2, 0.6, true, 25, 38, 3, false, true},
		{"streak 5 ×2 incl upset", pred("ind", "x", "x"), m, 4, 0.1, true, 20, 40, 5, true, true},
		{"tie leaves streak", pred("ind", "a", "x"), completed("m1", "2026-01-01T10:00:00Z", nil, []string{"a"}, []string{}), 3, 0.5, true, 15, 15, 3, false, true},
		{"unsettled", pred("ind", "a", "w"), Match{ID: "m1", State: Live}, 2, 0, false, 0, 0, 2, false, false},
	}
	for _, c := range cases {
		o := ScoreOne(c.p, c.m, c.streakBefore, c.share, c.known)
		if o.Base != c.wantBase || o.Points != c.wantPoints || o.Streak != c.wantStreak || o.Upset != c.wantUpset || o.Settled != c.wantSettled {
			t.Errorf("%s: got base=%d points=%d streak=%d upset=%v settled=%v", c.name, o.Base, o.Points, o.Streak, o.Upset, o.Settled)
		}
	}
}

func TestSettleHistoryOrdersByStartAndCarriesStreak(t *testing.T) {
	ind := Ptr("ind")
	matches := map[string]Match{}
	var preds []Prediction
	// Six correct winner picks, given out of order; the 6th match (by start) is a tie.
	for i, day := range []int{5, 1, 3, 2, 4, 6} {
		id := string(rune('a' + i))
		start := time.Date(2026, 1, day, 10, 0, 0, 0, time.UTC).Format(time.RFC3339)
		w := ind
		if day == 6 {
			w = nil
		}
		matches[id] = completed(id, start, w, []string{}, []string{})
		preds = append(preds, Prediction{UserID: "u", MatchID: id, WinnerID: "ind"})
	}
	preds = append(preds, Prediction{UserID: "u", MatchID: "missing", WinnerID: "ind"})
	lookup := func(id string) (Match, bool) { m, ok := matches[id]; return m, ok }
	share := func(Match) (float64, bool) { return 0.5, true }
	out := SettleHistory(preds, lookup, share)
	if len(out) != 6 {
		t.Fatalf("want 6 outcomes, got %d", len(out))
	}
	wantStreaks := []int{1, 2, 3, 4, 5, 5}
	wantPoints := []int{10, 10, 15, 15, 20, 0}
	for i, o := range out {
		if o.Streak != wantStreaks[i] || o.Points != wantPoints[i] {
			t.Errorf("#%d: streak=%d points=%d, want %d/%d", i, o.Streak, o.Points, wantStreaks[i], wantPoints[i])
		}
	}
	s := SummariseHistory(out, map[string]bool{"f": true})
	if s.Points != 70 || s.Streak != 5 || s.MaxStreak != 5 || s.Accuracy == nil || *s.Accuracy != 1 {
		t.Errorf("summary: %+v", s)
	}
	b := Badges(s, 3)
	ids := []string{}
	for _, x := range b {
		ids = append(ids, x.ID)
	}
	if len(ids) != 3 || ids[0] != "first-call" || ids[1] != "five-match-streak" || ids[2] != "stadium-regular" {
		t.Errorf("badges: %v", ids)
	}
}

func TestBadges(t *testing.T) {
	cases := []struct {
		name    string
		s       HistoryStats
		stadium int
		want    int
	}{
		{"none", HistoryStats{}, 2, 0},
		{"all", HistoryStats{Correct: 1, Upsets: 1, MaxStreak: 5, Perfect: 1}, 3, 5},
	}
	for _, c := range cases {
		got := Badges(c.s, c.stadium)
		if got == nil || len(got) != c.want {
			t.Errorf("%s: got %v", c.name, got)
		}
	}
}
