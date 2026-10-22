package sim

import (
	"bytes"
	"encoding/json"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

var anchor = time.Date(2026, 10, 22, 18, 0, 0, 0, time.UTC)

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestDeterminism(t *testing.T) {
	a, b := NewWorld(anchor), NewWorld(anchor)
	for _, now := range []time.Time{anchor, anchor.Add(37 * time.Minute), anchor.Add(3 * 24 * time.Hour)} {
		for name, f := range map[string]func(*World) any{
			"fixtures": func(w *World) any { return w.Fixtures(now) },
			"classics": func(w *World) any { return w.Classics() },
			"logs":     func(w *World) any { return w.Logs(now) },
			"preds":    func(w *World) any { return w.Predictions(now) },
			"demo": func(w *World) any {
				l, p, f := w.DemoHistory("u1", now)
				return []any{l, p, f}
			},
		} {
			if !bytes.Equal(mustJSON(t, f(a)), mustJSON(t, f(b))) {
				t.Errorf("%s differs between identical worlds at %v", name, now)
			}
		}
	}
	for _, m := range append(a.Fixtures(anchor), a.Classics()...) {
		if !bytes.Equal(mustJSON(t, a.Crowd(m.ID)), mustJSON(t, b.Crowd(m.ID))) {
			t.Errorf("crowd %s differs", m.ID)
		}
	}
	for _, l := range a.Logs(anchor)[:40] {
		if a.BaseLikes(l.ID) != b.BaseLikes(l.ID) || a.BaseThereToo(l.ID) != b.BaseThereToo(l.ID) ||
			!bytes.Equal(mustJSON(t, a.Replies(l.ID, anchor)), mustJSON(t, b.Replies(l.ID, anchor))) {
			t.Errorf("log extras differ for %s", l.ID)
		}
	}
	// IDs never depend on the anchor.
	other := NewWorld(anchor.Add(17 * time.Hour))
	for k, m := range a.Fixtures(anchor) {
		if other.Fixtures(anchor)[k].ID != m.ID {
			t.Fatal("fixture ids depend on the anchor")
		}
	}
	// A literal World (no cached state) behaves the same.
	lit := &World{Anchor: anchor}
	if !bytes.Equal(mustJSON(t, lit.Logs(anchor)), mustJSON(t, a.Logs(anchor))) {
		t.Error("literal World differs from NewWorld")
	}
}

func TestStateCoverageAtAnchor(t *testing.T) {
	w := NewWorld(anchor)
	ms := w.Fixtures(anchor)
	if len(ms) < 25 {
		t.Fatalf("only %d fixtures", len(ms))
	}
	var liveLO, liveFirst, liveChase, liveTest, lockSoon, next48, last24, recent, winnerless int
	comps, formats, genders := map[domain.CompetitionID]bool{}, map[domain.Format]bool{}, map[domain.Gender]bool{}
	indMen, indWomen := false, false
	for k, m := range ms {
		if k > 0 && ms[k-1].StartTime > m.StartTime {
			t.Error("fixtures not sorted by start time")
		}
		f := fixtureByID[m.ID]
		comps[m.CompetitionID], formats[m.Format], genders[m.Gender] = true, true, true
		if m.CompetitionID == "international" {
			indMen = indMen || slices.Contains(m.TeamIDs[:], "ind")
			indWomen = indWomen || slices.Contains(m.TeamIDs[:], "ind-w")
		}
		start, end := w.start(f), w.end(f)
		switch m.State {
		case domain.Live:
			if m.Live == nil || m.Live.Innings == nil || m.Result != nil {
				t.Errorf("%s: live without live score", m.ID)
				continue
			}
			if m.Format == domain.FormatTest {
				if strings.HasPrefix(m.Live.StatusText, "Day 2") || strings.HasPrefix(m.Live.StatusText, "Day 3") {
					liveTest++
				}
				continue
			}
			liveLO++
			el := anchor.Sub(start).Minutes()
			if len(m.Live.Innings) == 1 && el >= 60 && el <= 80 {
				liveFirst++
			}
			if len(m.Live.Innings) == 2 && strings.Contains(m.Live.StatusText, "need") {
				liveChase++
			}
		case domain.Upcoming:
			if m.Live != nil || m.Result != nil {
				t.Errorf("%s: upcoming leaks live/result", m.ID)
			}
			d := start.Sub(anchor)
			if d >= 15*time.Minute && d <= 35*time.Minute {
				lockSoon++
			} else if d <= 48*time.Hour {
				next48++
			}
		case domain.Completed:
			if m.Result == nil || m.Result.Innings == nil {
				t.Errorf("%s: completed without result", m.ID)
				continue
			}
			ago := anchor.Sub(end)
			if ago <= 24*time.Hour {
				last24++
			} else if ago >= 48*time.Hour && ago <= 10*24*time.Hour {
				recent++
			}
			if m.Result.WinnerID == nil {
				winnerless++
			}
		}
	}
	check := func(name string, got, want int) {
		if got < want {
			t.Errorf("%s: got %d, want >= %d", name, got, want)
		}
	}
	check("live limited-overs", liveLO, 2)
	check("live first innings ~70 min", liveFirst, 1)
	check("live chase", liveChase, 1)
	check("live Test on day 2/3", liveTest, 1)
	check("lock within ~25 min", lockSoon, 1)
	check("more starts within 48h", next48, 4)
	check("completed last 24h", last24, 3)
	check("completed 2-10 days ago", recent, 14)
	if winnerless != 1 {
		t.Errorf("want exactly one completed tie/no result, got %d", winnerless)
	}
	for _, c := range []domain.CompetitionID{"international", "ipl", "wpl", "world-cup", "bbl", "sa20", "the-hundred"} {
		if !comps[c] {
			t.Errorf("competition %s missing", c)
		}
	}
	for _, f := range []domain.Format{domain.FormatT20, domain.FormatODI, domain.FormatTest, domain.FormatHundred} {
		if !formats[f] {
			t.Errorf("format %s missing", f)
		}
	}
	if !genders[domain.Men] || !genders[domain.Women] || !indMen || !indWomen {
		t.Error("need men's and women's fixtures, incl. India men and India women internationals")
	}
}

func TestTeamsAndPlayers(t *testing.T) {
	w := NewWorld(anchor)
	teams, players := w.Teams(), w.Players()
	for _, id := range strings.Fields(`ind aus eng sa nz pak sl wi ban afg zim ire ned sco usa kenya nam uae
		ind-w aus-w eng-w sa-w nz-w pak-w sl-w wi-w ban-w ire-w csk mi rcb kkr srh rr dc pbks gt lsg mi-w rcb-w dc-w upw gg-w
		sixers scorchers heat stars renegades strikers hurricanes thunder mict jsk pc sec dsg prc
		invincibles brave phoenix spirit originals superchargers rockets fire`) {
		tm, ok := teams[id]
		if !ok || tm.Name == "" || tm.Short == "" || !strings.HasPrefix(tm.Colour, "#") || tm.Kind == "" {
			t.Errorf("team %s missing or incomplete", id)
		}
	}
	tieTop := false
	for _, m := range append(w.Fixtures(anchor.Add(30*24*time.Hour)), w.Classics()...) {
		for _, id := range m.TeamIDs {
			if _, ok := teams[id]; !ok {
				t.Errorf("%s: unknown team %q", m.ID, id)
			}
		}
		if m.Kind != "fixture" {
			continue
		}
		inXI := map[string]bool{}
		for team, xi := range m.XI {
			if len(xi) != 11 || !slices.Contains(m.TeamIDs[:], team) {
				t.Errorf("%s: bad XI for %s", m.ID, team)
			}
			for _, pid := range xi {
				p, ok := players[pid]
				if !ok || p.TeamID != team {
					t.Errorf("%s: XI player %s missing", m.ID, pid)
				}
				inXI[pid] = true
			}
		}
		r := m.Result
		ids := append(append([]string{}, r.TopRunScorerIDs...), r.TopWicketTakerIDs...)
		if r.OfficialPotmID != nil {
			ids = append(ids, *r.OfficialPotmID)
		}
		ids = append(ids, fixtureByID[m.ID].fav)
		for _, id := range ids {
			if !inXI[id] {
				t.Errorf("%s: performer %s not in XI", m.ID, id)
			}
		}
		if len(r.TopRunScorerIDs) > 1 {
			tieTop = true
		}
		if r.WinnerID != nil && !slices.Contains(m.TeamIDs[:], *r.WinnerID) {
			t.Errorf("%s: winner not playing", m.ID)
		}
	}
	if !tieTop {
		t.Error("want a fixture with a tie for top scorer")
	}
}

// walk fails on any null except the nullable contract fields.
func walk(t *testing.T, where, key string, v any) {
	switch x := v.(type) {
	case nil:
		if key != "winnerId" && key != "officialPotmId" && key != "fanPotmId" {
			t.Errorf("%s: null at %q", where, key)
		}
	case map[string]any:
		for k, c := range x {
			walk(t, where, k, c)
		}
	case []any:
		for _, c := range x {
			walk(t, where, key, c)
		}
	}
}

func TestNoNulls(t *testing.T) {
	w := NewWorld(anchor)
	far := anchor.Add(60 * 24 * time.Hour)
	dl, dp, df := w.DemoHistory("u1", anchor)
	for name, v := range map[string]any{
		"fixtures": w.Fixtures(anchor), "fixturesLater": w.Fixtures(far), "fixturesEarlier": w.Fixtures(anchor.Add(-30 * 24 * time.Hour)),
		"classics": w.Classics(), "logs": w.Logs(far), "preds": w.Predictions(far), "demoLogs": dl, "demoPreds": dp, "follows": df,
		"friends": w.Friends(), "public": w.PublicFans(), "teams": w.Teams(), "players": w.Players(),
	} {
		var x any
		if err := json.Unmarshal(mustJSON(t, v), &x); err != nil {
			t.Fatal(err)
		}
		walk(t, name, "", x)
	}
}

func TestPeopleAndActivity(t *testing.T) {
	w := NewWorld(anchor)
	if n := len(w.Friends()); n != 8 {
		t.Errorf("friends = %d", n)
	}
	if n := len(w.PublicFans()); n < 14 {
		t.Errorf("public fans = %d", n)
	}
	for _, u := range append(w.Friends(), w.PublicFans()...) {
		if got, ok := w.User(u.ID); !ok || got.Name != u.Name || u.Bio == "" || u.Handle == "" {
			t.Errorf("user %s", u.ID)
		}
		if _, ok := w.Teams()[u.FavouriteTeamID]; !ok {
			t.Errorf("user %s: unknown favourite team", u.ID)
		}
	}
	for _, now := range []time.Time{anchor, anchor.Add(60 * 24 * time.Hour)} {
		logs := w.Logs(now)
		if len(logs) == 0 {
			t.Fatal("no logs")
		}
		ids := map[string]bool{}
		for k, l := range logs {
			if ids[l.ID] {
				t.Errorf("duplicate log id %s", l.ID)
			}
			ids[l.ID] = true
			at, _ := time.Parse(time.RFC3339, l.CreatedAt)
			if at.After(now) || (k > 0 && logs[k-1].CreatedAt < l.CreatedAt) {
				t.Errorf("log %s: bad createdAt/order", l.ID)
			}
			m, ok := w.Match(l.MatchID, now)
			if !ok || m.State != domain.Completed {
				t.Errorf("log %s on a match that isn't completed", l.ID)
			}
			if l.Rating < 0.5 || l.Rating > 5 || math.Mod(l.Rating*2, 1) != 0 {
				t.Errorf("log %s: rating %v", l.ID, l.Rating)
			}
			for pid, r := range l.PlayerRatings {
				if r < 1 || r > 10 || (m.XI != nil && !slices.Contains(append(m.XI[m.TeamIDs[0]], m.XI[m.TeamIDs[1]]...), pid)) {
					t.Errorf("log %s: bad player rating %s=%d", l.ID, pid, r)
				}
			}
			if strings.ContainsAny(l.Review+l.PlaceNote, "{}") {
				t.Errorf("log %s leaks a placeholder: %q / %q", l.ID, l.Review, l.PlaceNote)
			}
			for _, rp := range w.Replies(l.ID, now) {
				if rp.UserID == l.UserID || rp.CreatedAt < l.CreatedAt || rp.CreatedAt > now.Format(time.RFC3339) {
					t.Errorf("reply %s invalid", rp.ID)
				}
			}
			if !ok {
				continue
			}
		}
		for _, p := range w.Predictions(now) {
			f := fixtureByID[p.MatchID]
			at, _ := time.Parse(time.RFC3339, p.CreatedAt)
			if f == nil || !at.Before(w.start(f)) || at.After(now) || !slices.Contains(f.teams[:], p.WinnerID) ||
				!f.inXI(p.TopRunScorerID) || !f.inXI(p.TopWicketTakerID) {
				t.Errorf("bad prediction %+v", p)
			}
		}
	}
}

func TestCrowd(t *testing.T) {
	w := NewWorld(anchor)
	if len(w.Crowd("nope").RatingBuckets) != 10 {
		t.Error("unknown match crowd")
	}
	disagree, upsets := 0, 0
	for _, m := range append(w.Fixtures(anchor), w.Classics()...) {
		c := w.Crowd(m.ID)
		if len(c.RatingBuckets) != 10 || c.PlayerTallies == nil || c.Watch == nil || c.Picks == nil {
			t.Fatalf("%s: malformed crowd", m.ID)
		}
		n := 0
		for _, b := range c.RatingBuckets {
			n += b
		}
		if n < 140 || n > 3100 {
			t.Errorf("%s: %d crowd ratings", m.ID, n)
		}
		if m.Kind != "fixture" || m.State != domain.Completed {
			continue
		}
		best, bestAvg := "", 0.0
		for _, xi := range [][]string{m.XI[m.TeamIDs[0]], m.XI[m.TeamIDs[1]]} {
			for _, pid := range xi {
				tl := c.PlayerTallies[pid]
				if avg := float64(tl.Sum) / float64(tl.Count); tl.Count >= 3 && avg > bestAvg {
					best, bestAvg = pid, avg
				}
			}
		}
		if p := m.Result.OfficialPotmID; p != nil && *p != best {
			disagree++
		}
		if wid := m.Result.WinnerID; wid != nil {
			share := float64(c.Picks[*wid]) / float64(c.Picks[m.TeamIDs[0]]+c.Picks[m.TeamIDs[1]])
			if share < 0.3 {
				upsets++
			}
		}
	}
	if disagree < 2 {
		t.Errorf("want >= 2 completed fixtures where fans disagree, got %d", disagree)
	}
	if upsets < 1 {
		t.Error("want a completed upset")
	}
}

func TestFriendsWatching(t *testing.T) {
	w := NewWorld(anchor)
	for _, m := range w.Fixtures(anchor) {
		got := w.FriendsWatching(m.ID, anchor)
		start, _ := time.Parse(time.RFC3339, m.StartTime)
		switch {
		case m.State == domain.Live && (len(got) < 2 || len(got) > 4):
			t.Errorf("%s live: %v", m.ID, got)
		case m.State == domain.Upcoming && start.Sub(anchor) <= 2*time.Hour && (len(got) < 1 || len(got) > 2):
			t.Errorf("%s soon: %v", m.ID, got)
		case m.State == domain.Completed && len(got) != 0:
			t.Errorf("%s completed: %v", m.ID, got)
		}
	}
	if w.FriendsWatching("nope", anchor) == nil {
		t.Error("nil slice")
	}
}

func TestLiveScoresThroughout(t *testing.T) {
	w := NewWorld(anchor)
	for _, f := range fixtures {
		for e := 0; e < f.dur; e += 7 {
			now := w.start(f).Add(time.Duration(e) * time.Minute)
			m := w.fixtureMatch(f, now)
			if m.State != domain.Live || m.Live == nil || m.Live.Innings == nil || m.Live.StatusText == "" {
				t.Fatalf("%s +%d: not live", f.id, e)
			}
			if strings.Contains(m.Live.StatusText, "need -") || strings.Contains(m.Live.StatusText, "need 0") {
				t.Errorf("%s +%d: %s", f.id, e, m.Live.StatusText)
			}
		}
	}
}

func TestDemoHistory(t *testing.T) {
	w := NewWorld(anchor)
	for _, now := range []time.Time{anchor, anchor.Add(9 * 24 * time.Hour)} {
		logs, preds, follows := w.DemoHistory("guest-1", now)
		if len(logs) < 4 || len(logs) > 6 {
			t.Errorf("demo logs = %d", len(logs))
		}
		classics := 0
		for k, l := range logs {
			if l.ID != "guest-1-demo-"+string(rune('1'+k)) || l.StubNo != k+1 || l.UserID != "guest-1" {
				t.Errorf("demo log id/stub: %s %d", l.ID, l.StubNo)
			}
			if strings.ContainsAny(l.Review+l.PlaceNote, "{}") || l.Review == "" {
				t.Errorf("demo review %q", l.Review)
			}
			if _, ok := classicByID[l.MatchID]; ok {
				classics++
				continue
			}
			f := fixtureByID[l.MatchID]
			if f == nil || w.end(f).After(now.Add(-48*time.Hour)) {
				t.Errorf("demo log on %s not finished 2+ days before now", l.MatchID)
			}
		}
		if len(classicEntries) >= 2 && classics != 2 {
			t.Errorf("demo classics = %d", classics)
		}
		right, upset, all3 := 0, false, false
		for _, p := range preds {
			f := fixtureByID[p.MatchID]
			at, _ := time.Parse(time.RFC3339, p.CreatedAt)
			if f == nil || w.end(f).After(now.Add(-48*time.Hour)) || !at.Before(w.start(f)) {
				t.Errorf("demo prediction on %s invalid", p.MatchID)
				continue
			}
			if p.WinnerID == f.winnerID() {
				right++
				upset = upset || f.upset
				all3 = all3 || (slices.Contains(f.topBatIDs, p.TopRunScorerID) && slices.Contains(f.topBowlID, p.TopWicketTakerID))
			}
		}
		if len(preds) < 5 || !upset || !all3 || right*10 < len(preds)*5 || right == len(preds) {
			t.Errorf("demo predictions: n=%d right=%d upset=%v all3=%v", len(preds), right, upset, all3)
		}
		if len(follows) != 5 {
			t.Errorf("follows = %v", follows)
		}
		for _, id := range follows {
			if _, ok := w.User(id); !ok || !isFriend(id) {
				t.Errorf("follow %s is not a friend", id)
			}
		}
	}
}

func TestClassics(t *testing.T) {
	w := NewWorld(anchor)
	cs := w.Classics()
	for k, m := range cs {
		if m.Kind != "classic" || m.State != domain.Completed || m.Result == nil || m.Classic == nil || m.XI != nil {
			t.Errorf("classic %s malformed", m.ID)
		}
		if !strings.HasSuffix(m.StartTime, "T12:00:00Z") || (k > 0 && cs[k-1].StartTime > m.StartTime) {
			t.Errorf("classic %s start/order", m.ID)
		}
		if got, ok := w.Match(m.ID, anchor); !ok || got.ID != m.ID {
			t.Errorf("Match(%s) not found", m.ID)
		}
	}
}
