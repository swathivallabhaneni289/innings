package sim

import (
	"strings"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// sameGame is a match's day, moved by shift days, and its two teams in either order.
func sameGame(m domain.Match, shift int) string {
	a, b := m.TeamIDs[0], m.TeamIDs[1]
	if b < a {
		a, b = b, a
	}
	day, _ := time.Parse(time.DateOnly, m.StartTime[:10])
	return day.AddDate(0, 0, shift).Format(time.DateOnly) + "|" + a + "|" + b
}

// TestArchive checks archive.json the way nobody could by hand: every match has teams the app knows, a winner who
// played, innings by its two sides and a real date; no id comes twice and nothing repeats a Vault match.
func TestArchive(t *testing.T) {
	now := time.Now()
	w := NewWorld(now)
	ms, teams := w.Archive(), w.Teams()
	if len(ms) == 0 {
		t.Fatal("archive.json is empty or unreadable")
	}
	vault := map[string]bool{}
	for _, c := range w.Classics() {
		vault[sameGame(c, 0)] = true
	}
	seen := map[string]bool{}
	for _, m := range ms {
		if seen[m.ID] || !strings.HasPrefix(m.ID, "ar-") {
			t.Errorf("id %q repeated or not an archive id", m.ID)
		}
		seen[m.ID] = true
		a, b := m.TeamIDs[0], m.TeamIDs[1]
		for _, id := range m.TeamIDs {
			if _, ok := teams[id]; !ok {
				t.Errorf("%s: unknown team %q", m.ID, id)
			}
		}
		if a == b || m.State != domain.Completed || m.Result == nil || m.Result.Text == "" {
			t.Errorf("%s: %+v", m.ID, m)
			continue
		}
		if win := m.Result.WinnerID; win != nil && *win != a && *win != b {
			t.Errorf("%s: winner %s didn't play", m.ID, *win)
		}
		for _, in := range m.Result.Innings {
			if in.TeamID != a && in.TeamID != b {
				t.Errorf("%s: innings by %s", m.ID, in.TeamID)
			}
		}
		if _, err := time.Parse(time.RFC3339, m.StartTime); err != nil {
			t.Errorf("%s: start %q", m.ID, m.StartTime)
		}
		if vault[sameGame(m, -1)] || vault[sameGame(m, 0)] || vault[sameGame(m, 1)] { // a reserve day may be dated either way
			t.Errorf("%s repeats a Vault match", m.ID)
		}
		if got, ok := w.Match(m.ID, now); !ok || got.ID != m.ID || !w.Archived(m.ID) {
			t.Errorf("%s: not found by id", m.ID)
		}
	}

	// Every World Cup, Champions Trophy and Asia Cup since the first, and every IPL season: about three thousand games.
	// A rebuild that comes out much smaller has lost a tournament.
	if len(ms) < 2900 || ms[0].StartTime[:4] != "1973" || ms[len(ms)-1].StartTime[:4] != "2026" {
		t.Errorf("archive has %d matches, %s to %s", len(ms), ms[0].StartTime, ms[len(ms)-1].StartTime)
	}

	// Known ones, as their scorecards have them. India v Pakistan at the 2023 World Cup:
	m, ok := w.Match("ar-1384403", now)
	if !ok || m.TeamIDs != [2]string{"pak", "ind"} || m.StartTime != "2023-10-14T12:00:00Z" || m.Venue.City != "Ahmedabad" ||
		m.Result.WinnerID == nil || *m.Result.WinnerID != "ind" || m.Result.Text != "India won by 7 wickets" || len(m.Result.Innings) != 2 ||
		m.Result.Innings[0] != (domain.InningsScore{TeamID: "pak", Runs: 191, Wickets: 10, Overs: "42.5"}) ||
		m.Result.Innings[1] != (domain.InningsScore{TeamID: "ind", Runs: 192, Wickets: 3, Overs: "30.3"}) {
		t.Fatalf("IND v PAK 2023 = %+v %+v", m, m.Result)
	}
	for _, k := range []struct {
		id, day, winner string
		first, second   domain.InningsScore
	}{
		// The first World Cup's first day: England's 334 at Lord's, India's 132 for 3 in 60 overs.
		{"ar-65035", "1975-06-07", "eng", domain.InningsScore{TeamID: "eng", Runs: 334, Wickets: 4, Overs: "60"}, domain.InningsScore{TeamID: "ind", Runs: 132, Wickets: 3, Overs: "60"}},
		// Eden Gardens, 1996: the semi-final the crowd stopped, awarded to Sri Lanka.
		{"ar-65190", "1996-03-13", "sl", domain.InningsScore{TeamID: "sl", Runs: 251, Wickets: 8, Overs: "50"}, domain.InningsScore{TeamID: "ind", Runs: 120, Wickets: 8, Overs: "34.1"}},
		// Durban, 2007: tied on 141, India won the bowl-out.
		{"ar-2007-09-14-ind-pak", "2007-09-14", "ind", domain.InningsScore{TeamID: "ind", Runs: 141, Wickets: 9, Overs: "20"}, domain.InningsScore{TeamID: "pak", Runs: 141, Wickets: 7, Overs: "20"}},
		// The 2012 IPL final in Chennai.
		{"ar-548381", "2012-05-27", "kkr", domain.InningsScore{TeamID: "csk", Runs: 190, Wickets: 3, Overs: "20"}, domain.InningsScore{TeamID: "kkr", Runs: 192, Wickets: 5, Overs: "19.4"}},
	} {
		m, ok := w.Match(k.id, now)
		if !ok || m.StartTime[:10] != k.day || m.Result.WinnerID == nil || *m.Result.WinnerID != k.winner || len(m.Result.Innings) != 2 ||
			m.Result.Innings[0] != k.first || m.Result.Innings[1] != k.second {
			t.Errorf("%s = %+v %+v", k.id, m, m.Result)
		}
	}
}
