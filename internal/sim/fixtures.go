package sim

import (
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

const (
	hr  = 60
	day = 24 * hr
)

// inn is one scripted innings: team index (0/1 into teams), runs, wickets, balls faced.
type inn struct{ t, r, w, b int }

func i(t, r, w, b int) inn { return inn{t, r, w, b} }

type fixtureDef struct {
	id       string
	comp     domain.CompetitionID
	format   domain.Format
	stage    string
	teams    [2]string
	venue    string
	offset   int   // minutes from Anchor to the first ball
	inns     []inn // scripted final innings, in batting order
	noResult bool
	topBat   []string // optional override (full player ids)
	topRuns  int
	topBowl  []string
	potm     string
	fanFav   string // fans' favourite when it differs from the official pick
	upset    bool   // crowd picked the winner < 30%
}

const (
	intl    = domain.CompetitionID("international")
	ipl     = domain.CompetitionID("ipl")
	wpl     = domain.CompetitionID("wpl")
	wc      = domain.CompetitionID("world-cup")
	bbl     = domain.CompetitionID("bbl")
	sa20    = domain.CompetitionID("sa20")
	hundred = domain.CompetitionID("the-hundred")
)

var fixtureDefs = []fixtureDef{
	// India men v Australia T20I series
	{id: "ind-aus-t20i-1", comp: intl, format: domain.FormatT20, stage: "1st T20I", teams: [2]string{"ind", "aus"}, venue: "arun-jaitley-stadium", offset: -5 * day,
		inns: []inn{i(0, 201, 6, 120), i(1, 178, 9, 120)}, topBat: []string{"ind-abhishek-sharma", "aus-travis-head"}, topRuns: 58},
	{id: "ind-aus-t20i-2", comp: intl, format: domain.FormatT20, stage: "2nd T20I", teams: [2]string{"ind", "aus"}, venue: "narendra-modi-stadium", offset: -3 * day,
		inns: []inn{i(1, 186, 7, 120), i(0, 187, 4, 113)}},
	{id: "ind-aus-t20i-3", comp: intl, format: domain.FormatT20, stage: "3rd T20I", teams: [2]string{"ind", "aus"}, venue: "m-chinnaswamy-stadium", offset: -70,
		inns: []inn{i(1, 188, 5, 120), i(0, 189, 7, 119)}},
	{id: "ind-aus-t20i-4", comp: intl, format: domain.FormatT20, stage: "4th T20I", teams: [2]string{"ind", "aus"}, venue: "ma-chidambaram-stadium", offset: 3 * day,
		inns: []inn{i(0, 172, 8, 120), i(1, 173, 5, 110)}},

	// IPL
	{id: "ipl-rr-dc-m16", comp: ipl, format: domain.FormatT20, stage: "Match 16", teams: [2]string{"rr", "dc"}, venue: "sawai-mansingh-stadium", offset: -4 * day,
		inns: []inn{i(0, 168, 9, 120), i(1, 159, 8, 120)}},
	{id: "ipl-rcb-pbks-m17", comp: ipl, format: domain.FormatT20, stage: "Match 17", teams: [2]string{"rcb", "pbks"}, venue: "m-chinnaswamy-stadium", offset: -3*day - 12*hr,
		inns: []inn{i(1, 175, 8, 120), i(0, 176, 3, 106)}, topBat: []string{"rcb-virat-kohli"}, topRuns: 102, potm: "rcb-virat-kohli"},
	{id: "ipl-dc-mi-m18", comp: ipl, format: domain.FormatT20, stage: "Match 18", teams: [2]string{"dc", "mi"}, venue: "arun-jaitley-stadium", offset: -2*day - 12*hr,
		inns: []inn{i(1, 151, 10, 116), i(0, 152, 2, 85)}},
	{id: "ipl-kkr-srh-m20", comp: ipl, format: domain.FormatT20, stage: "Match 20", teams: [2]string{"kkr", "srh"}, venue: "eden-gardens", offset: -(200 + 2*hr),
		inns: []inn{i(1, 231, 4, 120), i(0, 162, 10, 99)}, topBat: []string{"srh-abhishek-sharma"}, topRuns: 84, potm: "srh-abhishek-sharma", fanFav: "srh-heinrich-klaasen"},
	{id: "ipl-mi-csk-m21", comp: ipl, format: domain.FormatT20, stage: "Match 21", teams: [2]string{"mi", "csk"}, venue: "wankhede-stadium", offset: 25,
		inns: []inn{i(0, 196, 5, 120), i(1, 197, 6, 119)}},
	{id: "ipl-rcb-kkr-m24", comp: ipl, format: domain.FormatT20, stage: "Match 24", teams: [2]string{"rcb", "kkr"}, venue: "m-chinnaswamy-stadium", offset: 5 * day,
		inns: []inn{i(0, 205, 4, 120), i(1, 179, 9, 120)}},

	// WPL
	{id: "wpl-mi-w-gg-w-m8", comp: wpl, format: domain.FormatT20, stage: "Match 8", teams: [2]string{"mi-w", "gg-w"}, venue: "brabourne-stadium", offset: -6 * day,
		inns: []inn{i(1, 164, 6, 120), i(0, 165, 5, 118)}},
	{id: "wpl-rcb-w-dc-w-m11", comp: wpl, format: domain.FormatT20, stage: "Match 11", teams: [2]string{"rcb-w", "dc-w"}, venue: "dy-patil-stadium", offset: 3 * hr,
		inns: []inn{i(1, 148, 7, 120), i(0, 149, 4, 112)}},

	// BBL
	{id: "bbl-thunder-strikers-m6", comp: bbl, format: domain.FormatT20, stage: "Match 6", teams: [2]string{"thunder", "strikers"}, venue: "sydney-showground-stadium", offset: -2*day - 19*hr,
		inns: []inn{i(0, 94, 2, 66)}, noResult: true},
	{id: "bbl-sixers-scorchers-m9", comp: bbl, format: domain.FormatT20, stage: "Match 9", teams: [2]string{"sixers", "scorchers"}, venue: "sydney-cricket-ground", offset: -160,
		inns: []inn{i(1, 172, 6, 120), i(0, 166, 8, 120)}},

	// SA20
	{id: "sa20-pc-sec-m5", comp: sa20, format: domain.FormatT20, stage: "Match 5", teams: [2]string{"pc", "sec"}, venue: "supersport-park", offset: -3*day - 5*hr,
		inns: []inn{i(1, 210, 5, 120), i(0, 214, 6, 113)}},
	{id: "sa20-mict-jsk-m7", comp: sa20, format: domain.FormatT20, stage: "Match 7", teams: [2]string{"mict", "jsk"}, venue: "newlands", offset: 27 * hr,
		inns: []inn{i(0, 177, 7, 120), i(1, 153, 10, 115)}},

	// The Hundred
	{id: "hundred-rockets-spirit-m26", comp: hundred, format: domain.FormatHundred, stage: "Match 26", teams: [2]string{"rockets", "spirit"}, venue: "trent-bridge", offset: -(165 + 7*hr),
		inns: []inn{i(1, 162, 5, 100), i(0, 158, 8, 100)}},

	// Sri Lanka v England Tests
	{id: "sl-eng-test-1", comp: intl, format: domain.FormatTest, stage: "1st Test", teams: [2]string{"sl", "eng"}, venue: "galle-international-stadium", offset: -9*day - 12*hr,
		inns: []inn{i(0, 368, 10, 712), i(1, 331, 10, 608), i(0, 214, 10, 435), i(1, 252, 6, 425)}},
	{id: "sl-eng-test-2", comp: intl, format: domain.FormatTest, stage: "2nd Test", teams: [2]string{"sl", "eng"}, venue: "sinhalese-sports-club-ground", offset: -(day + 330),
		inns: []inn{i(1, 412, 10, 793), i(0, 298, 10, 592), i(1, 245, 7, 486), i(0, 203, 10, 464)}},

	// India Women v South Africa Women ODIs
	{id: "ind-w-sa-w-odi-1", comp: intl, format: domain.FormatODI, stage: "1st ODI", teams: [2]string{"ind-w", "sa-w"}, venue: "county-ground-bristol", offset: -9*day - 10*hr,
		inns: []inn{i(0, 281, 6, 300), i(1, 254, 10, 289)}, topBat: []string{"ind-w-smriti-mandhana"}, topRuns: 117},
	{id: "ind-w-sa-w-odi-2", comp: intl, format: domain.FormatODI, stage: "2nd ODI", teams: [2]string{"ind-w", "sa-w"}, venue: "county-ground-taunton", offset: -8 * day,
		inns: []inn{i(1, 248, 8, 300), i(0, 251, 5, 277)}},

	// ICC Women's T20 World Cup
	{id: "wc-aus-w-sa-w-m9", comp: wc, format: domain.FormatT20, stage: "Match 9", teams: [2]string{"aus-w", "sa-w"}, venue: "headingley", offset: -7 * day,
		inns: []inn{i(0, 158, 6, 120), i(1, 147, 9, 120)}},
	{id: "wc-ind-w-pak-w-m10", comp: wc, format: domain.FormatT20, stage: "Match 10", teams: [2]string{"ind-w", "pak-w"}, venue: "old-trafford", offset: -6*day - 12*hr,
		inns: []inn{i(1, 112, 10, 98), i(0, 113, 3, 86)}},
	{id: "wc-eng-w-nz-w-m11", comp: wc, format: domain.FormatT20, stage: "Match 11", teams: [2]string{"eng-w", "nz-w"}, venue: "the-rose-bowl", offset: -6 * day,
		inns: []inn{i(1, 139, 8, 120), i(0, 140, 4, 107)}},
	{id: "wc-wi-w-sl-w-m12", comp: wc, format: domain.FormatT20, stage: "Match 12", teams: [2]string{"wi-w", "sl-w"}, venue: "county-ground-bristol", offset: -5*day - 12*hr,
		inns: []inn{i(1, 151, 5, 120), i(0, 132, 9, 120)}, topBat: []string{"sl-w-chamari-athapaththu"}, topRuns: 71, upset: true},
	{id: "wc-nz-w-wi-w-m13", comp: wc, format: domain.FormatT20, stage: "Match 13", teams: [2]string{"nz-w", "wi-w"}, venue: "county-ground-taunton", offset: -4*day - 12*hr,
		inns: []inn{i(0, 171, 4, 120), i(1, 129, 8, 120)}},
	{id: "wc-ind-w-eng-w-m14", comp: wc, format: domain.FormatT20, stage: "Match 14", teams: [2]string{"ind-w", "eng-w"}, venue: "edgbaston", offset: -(200 + 20*hr),
		inns: []inn{i(1, 167, 6, 120), i(0, 168, 8, 120)}, potm: "ind-w-deepti-sharma", fanFav: "ind-w-richa-ghosh"},
	{id: "wc-aus-w-ind-w-sf1", comp: wc, format: domain.FormatT20, stage: "Semi-final 1", teams: [2]string{"aus-w", "ind-w"}, venue: "the-oval", offset: 20 * hr,
		inns: []inn{i(0, 149, 7, 120), i(1, 150, 6, 117)}},
	{id: "wc-eng-w-sa-w-sf2", comp: wc, format: domain.FormatT20, stage: "Semi-final 2", teams: [2]string{"eng-w", "sa-w"}, venue: "the-oval", offset: 44 * hr,
		inns: []inn{i(1, 141, 8, 120), i(0, 136, 9, 120)}},
}

// fixture is a compiled fixtureDef: the full scripted result, independent of the anchor.
type fixture struct {
	fixtureDef
	gender    domain.Gender
	dur       int // minutes from first ball to completion
	maxBalls  int // per innings (limited overs)
	winner    int // team index, -1 = none
	text      string
	topBatIDs []string
	topBowlID []string
	topWkts   int
	potmID    string // "" = none
	fav       string // fans' player of the match
	close     float64
	vibe      string
	margin    string
	pop       float64
	xi        map[string][]string
}

func (f *fixture) winnerID() string {
	if f.winner < 0 {
		return ""
	}
	return f.teams[f.winner]
}

func (f *fixture) loserID() string {
	if f.winner < 0 {
		return ""
	}
	return f.teams[1-f.winner]
}

func (f *fixture) inXI(pid string) bool {
	return slices.Contains(f.xi[f.teams[0]], pid) || slices.Contains(f.xi[f.teams[1]], pid)
}

func durationMin(f domain.Format) int {
	switch f {
	case domain.FormatHundred:
		return 165
	case domain.FormatODI:
		return 450
	case domain.FormatTest:
		return 4*day + 12*hr
	}
	return 200
}

func maxBallsFor(f domain.Format) int {
	switch f {
	case domain.FormatHundred:
		return 100
	case domain.FormatODI:
		return 300
	case domain.FormatTest:
		return 0
	}
	return 120
}

func oversText(format domain.Format, balls int) string {
	if format == domain.FormatHundred {
		return fmt.Sprintf("%d balls", balls)
	}
	if balls%6 == 0 {
		return fmt.Sprintf("%d", balls/6)
	}
	return fmt.Sprintf("%d.%d", balls/6, balls%6)
}

func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

var fixtures, fixtureByID = compileFixtures()

func compileFixtures() ([]*fixture, map[string]*fixture) {
	list := []*fixture{}
	byID := map[string]*fixture{}
	for _, d := range fixtureDefs {
		f := &fixture{fixtureDef: d, gender: teamTable[d.teams[0]].Gender, dur: durationMin(d.format), maxBalls: maxBallsFor(d.format), winner: -1}
		f.xi = map[string][]string{d.teams[0]: xiIDs(d.teams[0]), d.teams[1]: xiIDs(d.teams[1])}
		f.decide()
		f.performers()
		f.classify()
		list = append(list, f)
		byID[f.id] = f
	}
	sort.SliceStable(list, func(a, b int) bool { return list[a].offset < list[b].offset })
	return list, byID
}

// decide derives winner, text, margin and closeness from the scripted innings.
func (f *fixture) decide() {
	if f.noResult || len(f.inns) == 0 {
		f.text, f.close, f.margin = "No result", 0.3, ""
		return
	}
	last := f.inns[len(f.inns)-1]
	tot := [2]int{}
	for _, in := range f.inns {
		tot[in.t] += in.r
	}
	x, y := last.t, 1-last.t
	switch {
	case tot[x] > tot[y]:
		wl := 10 - last.w
		f.winner, f.margin = x, plural(wl, "wicket")
		if f.format == domain.FormatTest {
			f.close = clampF(1-0.08*float64(wl-1), 0, 1)
		} else {
			left := f.maxBalls - last.b
			f.close = clampF(1-0.08*float64(wl-1)-float64(left)/(0.25*float64(f.maxBalls))*0.5, 0, 1)
		}
	case tot[x] < tot[y] && (last.w == 10 || (f.format != domain.FormatTest && last.b >= f.maxBalls)):
		d := tot[y] - tot[x]
		f.winner, f.margin = y, plural(d, "run")
		scale := map[domain.Format]float64{domain.FormatT20: 40, domain.FormatHundred: 35, domain.FormatODI: 80, domain.FormatTest: 250}[f.format]
		f.close = clampF(1-float64(d)/scale, 0, 1)
	case tot[x] == tot[y]:
		f.text, f.close = "Match tied", 1
		return
	default:
		f.text, f.close = "Match drawn", 0.5
		return
	}
	f.text = teamName(f.teams[f.winner]) + " won by " + f.margin
}

func batters(xi []string) []string { return xi[:7] }
func bowlers(xi []string) []string { return xi[6:] }

func (f *fixture) performers() {
	r := newRNG("perf", f.id)
	side := func(prefWinner float64) int {
		if f.winner >= 0 {
			if r.chance(prefWinner) {
				return f.winner
			}
			return 1 - f.winner
		}
		return r.intn(2)
	}
	// top run scorer(s)
	f.topBatIDs = []string{}
	for _, id := range f.topBat {
		if f.inXI(id) {
			f.topBatIDs = append(f.topBatIDs, id)
		}
	}
	if len(f.topBatIDs) == 0 {
		t := side(0.7)
		if f.noResult {
			t = f.inns[0].t
		}
		f.topBatIDs = []string{pick(r, batters(f.xi[f.teams[t]])[:5])}
	}
	if f.topRuns == 0 {
		maxR := 0
		for _, in := range f.inns {
			if in.r > maxR {
				maxR = in.r
			}
		}
		f.topRuns = int(float64(maxR) * r.frange(0.3, 0.42))
		limit := map[domain.Format]int{domain.FormatT20: 92, domain.FormatHundred: 85, domain.FormatODI: 98, domain.FormatTest: 160}[f.format]
		f.topRuns = clampI(f.topRuns, 22, limit)
	}
	// top wicket-taker(s)
	f.topBowlID = []string{}
	for _, id := range f.topBowl {
		if f.inXI(id) {
			f.topBowlID = append(f.topBowlID, id)
		}
	}
	if len(f.topBowlID) == 0 {
		t := side(0.65)
		if f.noResult {
			t = 1 - f.inns[0].t
		}
		taken := 0
		for _, in := range f.inns {
			if in.t != t {
				taken += in.w
			}
		}
		if taken > 0 {
			f.topBowlID = []string{pick(r, bowlers(f.xi[f.teams[t]]))}
			if f.format == domain.FormatTest {
				f.topWkts = min(taken, r.between(5, 8))
			} else {
				f.topWkts = min(taken, r.between(2, 4))
			}
		}
	}
	// official player of the match
	teamOf := func(pid string) string { return playerTable[pid].TeamID }
	switch {
	case f.potm != "" && f.inXI(f.potm):
		f.potmID = f.potm
	case f.winner < 0:
		f.potmID = ""
	case teamOf(f.topBatIDs[0]) == f.winnerID() && (len(f.topBowlID) == 0 || teamOf(f.topBowlID[0]) != f.winnerID() || r.chance(0.6)):
		f.potmID = f.topBatIDs[0]
	case len(f.topBowlID) > 0 && teamOf(f.topBowlID[0]) == f.winnerID():
		f.potmID = f.topBowlID[0]
	default:
		f.potmID = batters(f.xi[f.winnerID()])[r.intn(4)]
	}
	f.fav = f.potmID
	if f.fanFav != "" && f.inXI(f.fanFav) {
		f.fav = f.fanFav
	}
	if f.fav == "" {
		f.fav = f.topBatIDs[0]
	}
	switch {
	case slices.Contains(f.teams[:], "ind") || slices.Contains(f.teams[:], "ind-w"):
		f.pop = 1
	case f.comp == ipl:
		f.pop = 0.85
	case f.comp == wc:
		f.pop = 0.7
	default:
		f.pop = 0.4
	}
}

func (f *fixture) classify() {
	big := map[domain.Format]int{domain.FormatT20: 190, domain.FormatHundred: 160, domain.FormatODI: 280, domain.FormatTest: 250}[f.format]
	switch {
	case f.noResult:
		f.vibe = "rain"
	case f.format != domain.FormatTest && (f.winner < 0 || f.close >= 0.72):
		f.vibe = "thriller" // thriller templates talk about last overs, so Tests use the other pools
	case f.topRuns >= 100:
		f.vibe = "century"
	}
	if f.vibe != "" {
		return
	}
	last := f.inns[len(f.inns)-1]
	chaseWon := f.winner == last.t
	var loserLast inn
	for _, in := range f.inns {
		if in.t != f.winner {
			loserLast = in
		}
	}
	switch {
	case chaseWon && f.inns[len(f.inns)-2].r >= big:
		f.vibe = "bigchase"
	case f.format != domain.FormatTest && loserLast.w == 10 && float64(loserLast.b) < 0.9*float64(f.maxBalls):
		f.vibe = "collapse"
	case f.close < 0.35:
		f.vibe = "thrashing"
	default:
		f.vibe = "routine"
	}
}

func (w *World) start(f *fixture) time.Time {
	return w.Anchor.Add(time.Duration(f.offset) * time.Minute)
}
func (w *World) end(f *fixture) time.Time {
	return w.Anchor.Add(time.Duration(f.offset+f.dur) * time.Minute)
}

func (w *World) stateAt(f *fixture, now time.Time) domain.MatchState {
	switch {
	case now.Before(w.start(f)):
		return domain.Upcoming
	case now.Before(w.end(f)):
		return domain.Live
	}
	return domain.Completed
}

func (f *fixture) result() *domain.MatchResult {
	res := &domain.MatchResult{Text: f.text, Innings: f.inningsScores(f.inns), TopRunScorerIDs: append([]string{}, f.topBatIDs...), TopWicketTakerIDs: append([]string{}, f.topBowlID...)}
	if f.winner >= 0 {
		res.WinnerID = domain.Ptr(f.teams[f.winner])
	}
	if f.potmID != "" {
		res.OfficialPotmID = domain.Ptr(f.potmID)
	}
	return res
}

func (f *fixture) inningsScores(ins []inn) []domain.InningsScore {
	out := []domain.InningsScore{}
	for _, in := range ins {
		out = append(out, domain.InningsScore{TeamID: f.teams[in.t], Runs: in.r, Wickets: in.w, Overs: oversText(f.format, in.b)})
	}
	return out
}

func (w *World) fixtureMatch(f *fixture, now time.Time) domain.Match {
	xi := map[string][]string{}
	for k, v := range f.xi {
		xi[k] = append([]string{}, v...)
	}
	m := domain.Match{
		ID: f.id, Kind: "fixture", CompetitionID: f.comp, Format: f.format, Gender: f.gender, Stage: f.stage,
		TeamIDs: f.teams, Venue: venue(f.venue), StartTime: w.start(f).UTC().Format(time.RFC3339),
		State: w.stateAt(f, now), XI: xi,
	}
	switch m.State {
	case domain.Live:
		m.Live = w.liveScore(f, now)
	case domain.Completed:
		m.Result = f.result()
	}
	return m
}
