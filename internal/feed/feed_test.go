package feed

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/sim"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// The testdata files are trimmed copies of ESPN's real answers on 29 October 2026, around 14:00 UTC: England v Sri
// Lanka (ODI) and South Africa v Australia (3rd Test, day 3) live, a WBBL game and UAE v Oman just finished,
// Afghanistan v Zimbabwe and New Zealand v India still to come, plus a Ranji Trophy and a Nepal Premier League game
// the app doesn't show.
var now = time.Date(2026, 10, 29, 14, 0, 0, 0, time.UTC)

func known() map[string]domain.Team { return sim.NewWorld(now).Teams() }

// espnFake serves the testdata as ESPN would, counting calls per path.
type espnFake struct {
	mu    sync.Mutex
	calls map[string]int
}

var summaries = map[string]string{"1550718": "summary-engsl.json", "1539435": "summary-nzind.json", "1543945": "summary-wbbl.json"}

func (f *espnFake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.calls[r.URL.Path]++
	f.mu.Unlock()
	file := ""
	switch {
	case r.URL.Path == "/v2/scoreboard/header" && r.URL.Query().Get("dates") == "":
		file = "scoreboard.json"
	case r.URL.Path == "/v2/scoreboard/header" && r.URL.Query().Get("dates") == "20261101":
		file = "scoreboard-20261101.json" // the week-ahead sweep finds the series' next T20I
	case r.URL.Path == "/v2/scoreboard/header":
		file = "scoreboard-empty.json"
	case strings.HasPrefix(r.URL.Path, "/site/v2/sports/cricket/") && strings.HasSuffix(r.URL.Path, "/summary"):
		file = summaries[r.URL.Query().Get("event")]
	}
	if file == "" {
		http.NotFound(w, r)
		return
	}
	b, err := os.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (f *espnFake) count(prefix string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for p, c := range f.calls {
		if strings.HasPrefix(p, prefix) {
			n += c
		}
	}
	return n
}

func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "feed.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func testSource(t *testing.T, f *espnFake, st *store.Store, clock *time.Time) *Source {
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	c := newClient()
	c.base = srv.URL + "/"
	return newSource(c, st, known(), func() time.Time { return *clock }, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func byID(ms []domain.Match) map[string]domain.Match {
	out := map[string]domain.Match{}
	for _, m := range ms {
		out[m.ID] = m
	}
	return out
}

func TestScoreboardMapping(t *testing.T) {
	f := &espnFake{calls: map[string]int{}}
	clock := now
	s := testSource(t, f, openStore(t), &clock)
	events, err := s.c.scoreboard(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	s.merge(events, now)
	ms := byID(s.Fixtures())
	if len(ms) != 6 {
		t.Fatalf("got %d matches, want 6 (Ranji Trophy and Nepal Premier League left out): %v", len(ms), keys(ms))
	}

	engsl := ms["es-1550718"]
	if engsl.CompetitionID != "international" || engsl.Format != domain.FormatODI || engsl.Stage != "6th Match" || engsl.TeamIDs != [2]string{"eng", "sl"} ||
		engsl.Venue.Name != "Gaddafi Stadium" || engsl.Venue.City != "Lahore" || engsl.State != domain.Live || engsl.StartTime != "2026-10-29T08:00:00Z" {
		t.Fatalf("ENG v SL = %+v", engsl)
	}
	if l := engsl.Live; l.BattingTeamID != "sl" || len(l.Innings) != 2 || l.Innings[0] != (domain.InningsScore{TeamID: "eng", Runs: 386, Wickets: 5}) ||
		l.Innings[1] != (domain.InningsScore{TeamID: "sl", Runs: 130, Wickets: 2, Overs: "21.2"}) || !strings.Contains(l.StatusText, "Sri Lanka require") {
		t.Fatalf("ENG v SL live = %+v", l)
	}

	// A Test: the toss says South Africa batted first, so the innings run SA, AUS, SA.
	test := ms["es-1525661"]
	if test.Format != domain.FormatTest || test.Stage != "3rd Test" {
		t.Fatalf("SA v AUS = %+v", test)
	}
	want := []domain.InningsScore{{TeamID: "sa", Runs: 329, Wickets: 10}, {TeamID: "aus", Runs: 429, Wickets: 10}, {TeamID: "sa", Runs: 63, Wickets: 2, Overs: "19.4"}}
	if got := test.Live.Innings; len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("SA v AUS innings = %+v", got)
	}

	// World Cup League 2 is a qualifier, not the World Cup; Oman is a nation the simulation doesn't have.
	uae := ms["es-1559565"]
	if uae.CompetitionID != "international" || uae.TeamIDs != [2]string{"uae", "oman"} || uae.State != domain.Completed ||
		uae.Result.WinnerID == nil || *uae.Result.WinnerID != "uae" || uae.Result.Innings[0].TeamID != "oman" {
		t.Fatalf("UAE v Oman = %+v %+v", uae, uae.Result)
	}
	if oman := s.Teams()["oman"]; oman.Kind != "international" || oman.Short != "OMAN" && oman.Short != "OMA" {
		t.Fatalf("Oman = %+v", oman)
	}

	// The WBBL sides wear their BBL colours.
	wbbl := ms["es-1543945"]
	if wbbl.CompetitionID != "bbl" || wbbl.Gender != domain.Women || wbbl.TeamIDs != [2]string{"stars-w", "hurricanes-w"} ||
		wbbl.Result.WinnerID == nil || *wbbl.Result.WinnerID != "stars-w" || wbbl.Result.Text != "Stars won by 18 runs" {
		t.Fatalf("WBBL = %+v %+v", wbbl, wbbl.Result)
	}
	if stars := s.Teams()["stars-w"]; stars.Colour != known()["stars"].Colour || stars.Short != "STA-W" {
		t.Fatalf("stars-w = %+v", stars)
	}

	nzind := ms["es-1539435"]
	if nzind.State != domain.Upcoming || nzind.TeamIDs != [2]string{"nz", "ind"} || nzind.Live != nil || nzind.Result != nil {
		t.Fatalf("NZ v IND = %+v", nzind)
	}
}

func keys(m map[string]domain.Match) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestPollFetchesSummariesSavesAndReloads(t *testing.T) {
	f := &espnFake{calls: map[string]int{}}
	st := openStore(t)
	clock := now
	s := testSource(t, f, st, &clock)
	wait := s.poll(context.Background())
	if wait != busyEvery {
		t.Fatalf("wait = %v with matches live", wait)
	}
	ms := byID(s.Fixtures())

	// Live: the playing XIs and the per-innings lines (with overs) from the summary.
	engsl := ms["es-1550718"]
	if len(engsl.XI["eng"]) != 11 || len(engsl.XI["sl"]) != 11 || engsl.Live.Innings[0].Overs != "50" || engsl.Venue.Country != "Pakistan" {
		t.Fatalf("ENG v SL after summary: xi %d/%d innings %+v venue %+v", len(engsl.XI["eng"]), len(engsl.XI["sl"]), engsl.Live.Innings, engsl.Venue)
	}
	// Upcoming: the announced squads, with roles.
	nzind := ms["es-1539435"]
	if len(nzind.XI["nz"]) != 14 || len(nzind.XI["ind"]) != 16 {
		t.Fatalf("NZ v IND squads: %d/%d", len(nzind.XI["nz"]), len(nzind.XI["ind"]))
	}
	pl := s.Players()
	if p := pl["esp-502714"]; p.Name != "Mitchell Santner" || p.Role != "AR" || p.TeamID != "nz" {
		t.Fatalf("Santner = %+v", p)
	}
	if p := pl["esp-1125976"]; p.Role != "BOWL" || p.TeamID != "ind" {
		t.Fatalf("Arshdeep = %+v", p)
	}
	if p := pl["esp-877051"]; p.Role != "WK" {
		t.Fatalf("Banton = %+v", p)
	}
	// Finished: top run-scorer, top wicket-taker and player of the match.
	res := ms["es-1543945"].Result
	if strings.Join(res.TopRunScorerIDs, ",") != "esp-329336" || strings.Join(res.TopWicketTakerIDs, ",") != "esp-1071705" ||
		res.OfficialPotmID == nil || *res.OfficialPotmID != "esp-1071705" {
		t.Fatalf("WBBL result = %+v", res)
	}
	if pl["esp-329336"].Name != "Meg Lanning" {
		t.Fatalf("Lanning = %+v", pl["esp-329336"])
	}

	// Today's scoreboard plus one per day of the week ahead, which found the series' 5th T20I on 1 November.
	boards, sums := f.count("/v2/scoreboard"), f.count("/site/")
	if boards != 1+aheadDays {
		t.Fatalf("scoreboard calls = %d", boards)
	}
	if fifth := ms["es-1539436"]; fifth.State != domain.Upcoming || fifth.Stage != "5th T20I" || fifth.Venue.Name != "Bay Oval" {
		t.Fatalf("week ahead: %+v", fifth)
	}

	// A minute later: the scoreboard again (not the week ahead), no summary due yet, and the summaries' additions
	// survive the merge.
	clock = now.Add(time.Minute)
	s.poll(context.Background())
	if f.count("/site/") != sums || f.count("/v2/scoreboard") != boards+1 {
		t.Fatalf("calls after a minute: %v", f.calls)
	}
	ms = byID(s.Fixtures())
	if len(ms["es-1550718"].XI["eng"]) != 11 || ms["es-1543945"].Result.OfficialPotmID == nil || ms["es-1550718"].Live.Innings[0].Overs != "50" {
		t.Fatal("the scoreboard merge dropped what the summaries added")
	}
	// The live matches' summaries come round again after two minutes; the finished WBBL game's never does.
	clock = now.Add(3 * time.Minute)
	before := f.calls["/site/v2/sports/cricket/21284/summary"]
	s.poll(context.Background())
	if f.calls["/site/v2/sports/cricket/1550706/summary"] < 2 || f.calls["/site/v2/sports/cricket/21284/summary"] != before {
		t.Fatalf("summary calls = %v", f.calls)
	}

	// A restart reads everything back without calling ESPN.
	calls := f.count("/")
	s2 := testSource(t, f, st, &clock)
	if err := s2.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := byID(s2.Fixtures())
	if len(got) != 7 || len(got["es-1539435"].XI["ind"]) != 16 || got["es-1543945"].Result.OfficialPotmID == nil || s2.Players()["esp-502714"].Name != "Mitchell Santner" {
		t.Fatalf("reloaded: %+v", got)
	}
	if f.count("/") != calls {
		t.Fatal("loading called ESPN")
	}
	if s2.Teams()["oman"].Name != "Oman" {
		t.Fatal("feed teams not reloaded")
	}
}

func TestQuietWhenNothingIsOn(t *testing.T) {
	clock := now
	s := testSource(t, &espnFake{calls: map[string]int{}}, openStore(t), &clock)
	s.recs["es-1"] = &record{Match: domain.Match{ID: "es-1", State: domain.Upcoming, StartTime: now.Add(5 * time.Hour).Format(time.RFC3339)}}
	if s.busy(now) {
		t.Fatal("a match five hours off isn't busy")
	}
	s.recs["es-1"].Match.StartTime = now.Add(30 * time.Minute).Format(time.RFC3339)
	if !s.busy(now) {
		t.Fatal("a match half an hour off is busy")
	}
}

func TestScoreboardDownKeepsWhatWeHave(t *testing.T) {
	f := &espnFake{calls: map[string]int{}}
	st := openStore(t)
	clock := now
	s := testSource(t, f, st, &clock)
	s.poll(context.Background())
	n := len(s.Fixtures())
	s.c.base = "http://127.0.0.1:1/" // ESPN unreachable
	clock = now.Add(10 * time.Minute)
	s.poll(context.Background())
	if len(s.Fixtures()) != n {
		t.Fatal("a failed poll lost matches")
	}
}

func TestWorldKeepsTheSimulationAwayFromRealMatches(t *testing.T) {
	clock := now
	simWorld := sim.NewWorld(now.Add(-72 * time.Hour))
	s := testSource(t, &espnFake{calls: map[string]int{}}, openStore(t), &clock)
	s.recs["es-1"] = &record{Match: domain.Match{ID: "es-1", TeamIDs: [2]string{"nepal", "oman"}, State: domain.Upcoming}}
	s.feedTeams["nepal"] = domain.Team{ID: "nepal", Name: "Nepal"}
	w := NewWorld(simWorld, s)

	if fx := w.Fixtures(now); len(fx) != 1 || fx[0].ID != "es-1" {
		t.Fatalf("fixtures = %+v", fx)
	}
	if _, ok := w.Teams()["nepal"]; !ok {
		t.Fatal("feed teams missing")
	}
	if _, ok := w.Teams()["ind"]; !ok {
		t.Fatal("simulation teams missing")
	}
	classics := map[string]bool{}
	for _, m := range w.Classics() {
		classics[m.ID] = true
	}
	logs := w.Logs(now)
	if len(logs) == 0 {
		t.Fatal("the simulated fans' Vault logs should stay")
	}
	for _, l := range logs {
		if !classics[l.MatchID] {
			t.Fatalf("log of a made-up fixture leaked: %s", l.MatchID)
		}
	}
	if len(w.Predictions(now)) != 0 || len(w.FriendsWatching("es-1", now)) != 0 {
		t.Fatal("simulated predictions or watchers on real matches")
	}
	c := w.Crowd("es-1")
	if len(c.RatingBuckets) != 10 || len(c.Picks) != 0 || c.PlayerTallies == nil || c.Watch == nil {
		t.Fatalf("crowd = %+v", c)
	}
	if _, ok := w.Match("es-1", now); !ok {
		t.Fatal("real match not found")
	}
}

func TestSquadsAndTossAreCheckedOftenEnough(t *testing.T) {
	clock := now
	s := testSource(t, &espnFake{calls: map[string]int{}}, openStore(t), &clock)
	add := func(id string, startsIn, lastLook time.Duration, xi bool) {
		m := domain.Match{ID: id, State: domain.Upcoming, StartTime: now.Add(startsIn).Format(time.RFC3339)}
		if xi {
			m.XI = map[string][]string{"ind": {"esp-1"}, "nz": {"esp-2"}}
		}
		s.recs[id] = &record{Match: m, SummaryAt: now.Add(-lastLook)}
	}
	add("es-toss", 90*time.Minute, 6*time.Minute, true)      // toss due: every five minutes
	add("es-later", 5*time.Hour, 6*time.Minute, true)        // squads in hand, looked six minutes ago: not yet
	add("es-nosquad", 3*24*time.Hour, 61*time.Minute, false) // no squad yet: hourly
	add("es-stale", 4*24*time.Hour, 7*time.Hour, true)       // squads, last looked seven hours ago: every six
	add("es-far", 10*24*time.Hour, 30*24*time.Hour, false)   // beyond the week: not at all
	got := strings.Join(s.wantSummary(now), ",")
	if got != "es-toss,es-nosquad,es-stale" {
		t.Fatalf("due = %s", got)
	}
}
