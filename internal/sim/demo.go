package sim

import (
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

var demoFollows = []string{"f-aditi", "f-kabir", "f-meera", "f-tom", "f-sipho"}

// demoLogPrefs are the fixtures a new guest has "already logged", in order of preference.
var demoLogPrefs = []string{"ind-aus-t20i-2", "wc-ind-w-pak-w-m10", "ipl-rcb-pbks-m17", "sl-eng-test-1", "ind-aus-t20i-1", "ipl-dc-mi-m18"}

// demoTop5 is the demo account's Top 5 (make demo), in order: Vault classics from five decades, men's and women's,
// ending with the 1983 final it has logged.
var demoTop5 = []string{"wc-2011-final", "test-2021-gabba", "wc-2019-final", "wwc-2025-final", "wc-1983-final"}

// demoPredPlan: fixture → outcome ("all" = winner + top scorer + top wicket-taker right, "win" = right winner, "lose" = wrong winner).
var demoPredPlan = []struct{ id, kind string }{
	{"wc-wi-w-sl-w-m12", "win"}, // the upset, called correctly
	{"ind-aus-t20i-1", "all"},
	{"ind-aus-t20i-2", "win"},
	{"ipl-rr-dc-m16", "lose"},
	{"wc-ind-w-pak-w-m10", "win"},
	{"ipl-dc-mi-m18", "lose"},
}

var demoPlaces = []struct {
	p    domain.WatchPlace
	note string
}{
	{domain.WatchHome, "My room, laptop propped on a pillow"},
	{domain.WatchFriends, "Karthik's place, way too many people"},
	{domain.WatchStadium, "{V}, Stand C, lost my voice"},
	{domain.WatchHostel, "Hostel common room, I had the good chair"},
	{domain.WatchCafe, "Corner cafe, nursed one coffee for three hours"},
}

var demoClassicPlaces = []struct {
	p    domain.WatchPlace
	note string
}{
	{domain.WatchHome, "My uncle's place in Hyderabad"},
	{domain.WatchHostel, "School hostel, one TV for eighty of us"},
}

// demoArchive are the past tournament finals a demo account on real matches has logged, where the simulated fixtures
// don't exist: a fan of India and Chennai Super Kings.
var demoArchive = []string{"ar-2021-10-15-csk-kkr", "ar-1466428", "ar-1496938"}

var demoArchivePlaces = []struct {
	p    domain.WatchPlace
	note string
}{
	{domain.WatchFriends, "Rooftop with the whole building"},
	{domain.WatchHome, "Family sofa, everyone shouting at the TV"},
	{domain.WatchOffice, "Office canteen, nobody did any work"},
}

// demoHistoryReal keeps demoHistory's Vault classics, puts finals from the archive in place of its fixtures, and
// leaves out its predictions, which are all on fixtures.
func (w *World) demoHistoryReal(userID string, now time.Time) ([]domain.MatchLog, []string) {
	all, _, follows := w.demoHistory(userID, now)
	logs := []domain.MatchLog{}
	for _, l := range all {
		if classicByID[l.MatchID] != nil {
			logs = append(logs, l)
		}
	}
	for k, id := range demoArchive {
		i, ok := archiveByID[id]
		if !ok {
			continue
		}
		m := archiveList[i]
		r := newRNG("demo-archive", userID, id)
		pl := demoArchivePlaces[k%len(demoArchivePlaces)]
		vals := map[string]string{"R": m.Result.Text, "Y": m.StartTime[:4], "V": m.Venue.Name}
		if m.Result.WinnerID != nil {
			vals["W"] = teamName(*m.Result.WinnerID)
		}
		l := domain.MatchLog{UserID: userID, MatchID: id, PlayerRatings: map[string]int{}, WatchedWith: []string{}, WatchedAt: pl.p, PlaceNote: pl.note}
		l.Rating = halfStep(4 + r.frange(0, 1))
		l.InstantClassic = l.Rating >= 4.5
		l.Companions = companionsFor(r, pl.p)
		l.Review = fillFrom(r, demoPool["archive"], vals)
		l.CreatedAt = ts(now.Add(-time.Duration((9-2*k)*day) * time.Minute))
		logs = append(logs, l)
	}
	return logs, follows
}

func (w *World) demoTop5() []string {
	ids := []string{}
	for _, id := range demoTop5 {
		if classicByID[id] != nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (w *World) demoHistory(userID string, now time.Time) ([]domain.MatchLog, []domain.Prediction, []string) {
	cutoff := now.Add(-48 * time.Hour)
	eligible := func(f *fixture) bool { return !w.end(f).After(cutoff) }

	type item struct {
		at  time.Time
		log domain.MatchLog
	}
	items := []item{}

	// Two classics: the most hyped, oldest first on ties.
	cl := make([]*classicEntry, 0, len(classicEntries))
	for k := range classicEntries {
		cl = append(cl, &classicEntries[k])
	}
	sort.SliceStable(cl, func(a, b int) bool { return cl[a].seed.Hype > cl[b].seed.Hype })
	for k, c := range cl[:min(2, len(cl))] {
		r := newRNG("demo-classic", userID, c.seed.ID)
		pl := demoClassicPlaces[k%len(demoClassicPlaces)]
		l := domain.MatchLog{UserID: userID, MatchID: c.seed.ID, PlayerRatings: map[string]int{}, WatchedWith: []string{}, WatchedAt: pl.p, PlaceNote: pl.note}
		l.Rating = halfStep(4 + 0.25*float64(c.seed.Hype-3) + r.frange(0, 0.5))
		l.InstantClassic = true
		l.Companions = domain.Ptr(r.between(3, 12))
		l.Review = fillFrom(r, demoPool["classic"], classicVals(c))
		items = append(items, item{at: now.Add(-time.Duration(12*day-k*90) * time.Minute), log: l})
	}

	// Fixtures that finished 2+ days ago.
	want := 6 - len(items)
	if len(items) == 0 {
		want = 4
	}
	chosen := []*fixture{}
	for _, id := range demoLogPrefs {
		if f := fixtureByID[id]; f != nil && eligible(f) && len(chosen) < want {
			chosen = append(chosen, f)
		}
	}
	for k := len(fixtures) - 1; k >= 0 && len(chosen) < want; k-- {
		if f := fixtures[k]; eligible(f) && !slices.Contains(chosen, f) {
			chosen = append(chosen, f)
		}
	}
	if len(items) == 2 && len(chosen) > 3 {
		chosen = chosen[:3] // 2 classics + 3 fixtures = 5 logs
	}
	for k, f := range chosen {
		r := newRNG("demo-fixture", userID, f.id)
		pl := demoPlaces[k%len(demoPlaces)]
		note, _ := fill(pl.note, map[string]string{"V": venueShort(f.venue)})
		l := domain.MatchLog{UserID: userID, MatchID: f.id, WatchedWith: []string{}, WatchedAt: pl.p, PlaceNote: note}
		l.Rating = fixtureRating(r, f, "")
		l.InstantClassic = l.Rating >= 4.5
		l.PlayerRatings, l.FanPotmID = playerRatings(r, f)
		l.Companions = companionsFor(r, pl.p)
		if k < 2 && len(friendIDs) > k {
			l.WatchedWith = append(l.WatchedWith, demoFollows[k])
		}
		l.Review = fillFrom(r, demoPool["fixture"], fixtureVals(f))
		items = append(items, item{at: w.end(f).Add(time.Duration(r.between(60, 180)) * time.Minute), log: l})
	}
	sort.SliceStable(items, func(a, b int) bool { return items[a].at.Before(items[b].at) })
	logs := []domain.MatchLog{}
	for k, it := range items {
		l := it.log
		l.ID = userID + "-demo-" + strconv.Itoa(k+1)
		l.StubNo = k + 1
		l.CreatedAt = ts(it.at)
		logs = append(logs, l)
	}

	// Settled predictions.
	preds := []domain.Prediction{}
	used := []string{}
	addPred := func(f *fixture, kind string) {
		r := newRNG("demo-pred", userID, f.id)
		p := domain.Prediction{UserID: userID, MatchID: f.id}
		switch kind {
		case "lose":
			p.WinnerID = f.loserID()
		default:
			p.WinnerID = f.winnerID()
		}
		other := f.teams[0]
		if other == p.WinnerID {
			other = f.teams[1]
		}
		p.TopRunScorerID = pick(r, batters(f.xi[p.WinnerID])[:5])
		p.TopWicketTakerID = pick(r, bowlers(f.xi[other]))
		if kind == "all" {
			p.TopRunScorerID = f.topBatIDs[0]
			if len(f.topBowlID) > 0 {
				p.TopWicketTakerID = f.topBowlID[0]
			}
		}
		p.CreatedAt = ts(w.start(f).Add(-time.Duration(r.between(2*hr, 20*hr)) * time.Minute))
		preds = append(preds, p)
		used = append(used, f.id)
	}
	for _, pp := range demoPredPlan {
		if f := fixtureByID[pp.id]; f != nil && eligible(f) && f.winner >= 0 {
			addPred(f, pp.kind)
		}
	}
	kinds := []string{"win", "lose", "win"}
	for k := len(fixtures) - 1; k >= 0 && len(preds) < 6; k-- {
		f := fixtures[k]
		if eligible(f) && f.winner >= 0 && !slices.Contains(used, f.id) {
			addPred(f, kinds[len(preds)%len(kinds)])
		}
	}
	return logs, preds, append([]string{}, demoFollows...)
}
