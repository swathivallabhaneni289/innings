package feed

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

func TestBattingCardFromRealSummary(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "summary-engsl.json"))
	if err != nil {
		t.Fatal(err)
	}
	var s esSummary
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	bc := card(s)
	if bc == nil || bc.innings != "2" || len(bc.batters) != 11 {
		t.Fatalf("card = %+v", bc)
	}
	by := map[string]cardBatter{}
	for _, x := range bc.batters {
		by[x.name] = x
	}
	if x := by["BKG Mendis"]; !x.hasBatted || x.runs != 89 || !x.out {
		t.Fatalf("Kusal Mendis = %+v", x)
	}
	if x := by["P Rathnayake"]; x.runs != 57 || x.out {
		t.Fatalf("Rathnayake = %+v", x)
	}
	if x := by["D Madushanka"]; x.hasBatted || x.out {
		t.Fatalf("yet to bat = %+v", x)
	}
}

// A T20 followed reading by reading: the total call opens, locks at 10 overs and settles when the innings ends; the
// chase call opens with the second innings, locks five overs in and settles on the result; batters' calls open at
// 25, lock at 40 and settle at 50 or a dismissal.
func TestLiveCallsFollowTheMatch(t *testing.T) {
	at := now
	r := &record{
		Match: domain.Match{ID: "es-9", Kind: "fixture", Format: domain.FormatT20, Gender: domain.Men, TeamIDs: [2]string{"afg", "zim"}, State: domain.Live},
		Event: esEvent{Competitors: []esCompetitor{{ID: "40", DisplayName: "Afghanistan"}, {ID: "9", DisplayName: "Zimbabwe"}}},
	}
	names := map[string]string{"esp-1": "Rahmanullah Gurbaz", "esp-2": "Ibrahim Zadran"}
	name := func(id string) string { return names[id] }
	read := func(bc *battingCard, innings ...domain.InningsScore) {
		r.Match.Live = &domain.LiveScore{Innings: innings}
		updateCalls(r, bc, name, at)
		at = at.Add(time.Minute)
	}
	state := func(id string) string {
		if c := r.call(id); c != nil {
			return c.State
		}
		return ""
	}
	afg := func(runs, wkts int, overs string) domain.InningsScore {
		return domain.InningsScore{TeamID: "afg", Runs: runs, Wickets: wkts, Overs: overs}
	}
	zim := func(runs, wkts int, overs string) domain.InningsScore {
		return domain.InningsScore{TeamID: "zim", Runs: runs, Wickets: wkts, Overs: overs}
	}
	bat := func(inn string, bs ...cardBatter) *battingCard { return &battingCard{innings: inn, batters: bs} }

	read(bat("1", cardBatter{id: "esp-1", runs: 27, hasBatted: true}, cardBatter{id: "esp-2", runs: 12, hasBatted: true}), afg(31, 0, "4.0"))
	if c := r.call("es-9:total"); c == nil || c.State != domain.CallOpen || c.Question != "Where will Afghanistan finish?" || len(c.Options) != 4 {
		t.Fatalf("total = %+v", c)
	}
	if c := r.call("es-9:ms:1:1:50"); c == nil || c.State != domain.CallOpen || c.Question != "Will Rahmanullah Gurbaz make 50?" {
		t.Fatalf("milestone = %+v", c)
	}
	if r.call("es-9:ms:1:2:50") != nil {
		t.Fatal("a batter on 12 got a call")
	}

	read(bat("1", cardBatter{id: "esp-1", runs: 41, hasBatted: true}, cardBatter{id: "esp-2", runs: 30, hasBatted: true}), afg(76, 0, "10.0"))
	if state("es-9:total") != domain.CallLocked || state("es-9:ms:1:1:50") != domain.CallLocked || state("es-9:ms:1:2:50") != domain.CallOpen {
		t.Fatalf("at 10 overs: %+v", r.Calls)
	}
	// A stale reading from before the lock never reopens a call.
	read(nil, afg(70, 0, "9.3"))
	if state("es-9:total") != domain.CallLocked {
		t.Fatal("a stale reading reopened the total call")
	}

	read(bat("1", cardBatter{id: "esp-1", runs: 52, hasBatted: true}, cardBatter{id: "esp-2", runs: 33, hasBatted: true, out: true}), afg(120, 1, "14.2"))
	if c := r.call("es-9:ms:1:1:50"); c.State != domain.CallSettled || *c.AnswerID != "yes" || c.Result != "Rahmanullah Gurbaz got to 50" {
		t.Fatalf("Gurbaz = %+v", c)
	}
	if c := r.call("es-9:ms:1:2:50"); c.State != domain.CallSettled || *c.AnswerID != "no" || c.Result != "Ibrahim Zadran out for 33" {
		t.Fatalf("Zadran = %+v", c)
	}

	// The chase begins: Afghanistan's 172 settles the total call; the chase call opens for 173.
	read(bat("2", cardBatter{id: "esp-5", runs: 28, hasBatted: true}), afg(172, 6, "20"), zim(8, 0, "1.0"))
	if c := r.call("es-9:total"); c.State != domain.CallSettled || *c.AnswerID != "170" || c.Result != "Afghanistan made 172" {
		t.Fatalf("total settled = %+v", c)
	}
	if c := r.call("es-9:chase"); c == nil || c.State != domain.CallOpen || c.Question != "Will Zimbabwe chase 173?" {
		t.Fatalf("chase = %+v", c)
	}
	read(nil, afg(172, 6, "20"), zim(39, 1, "5.0"))
	if state("es-9:chase") != domain.CallLocked {
		t.Fatal("chase not locked five overs in")
	}

	// Zimbabwe win. The chase call is settled; a batter's call left open when it ended is void.
	r.Match.State, r.Match.Live = domain.Completed, nil
	r.Match.Result = &domain.MatchResult{WinnerID: domain.Ptr("zim"), Text: "Zimbabwe won by 4 wickets", Innings: []domain.InningsScore{afg(172, 6, "20"), zim(175, 6, "19.3")}}
	updateCalls(r, nil, name, at)
	if c := r.call("es-9:chase"); c.State != domain.CallSettled || *c.AnswerID != "yes" {
		t.Fatalf("chase settled = %+v", c)
	}
	if c := r.call("es-9:ms:2:5:50"); c == nil || c.State != domain.CallVoid {
		t.Fatalf("unresolved milestone = %+v", c)
	}
	n := len(r.Calls)
	updateCalls(r, nil, name, at)
	if len(r.Calls) != n {
		t.Fatal("calls were created twice")
	}
}

func TestCallsFirstSeenPastTheirLockStartLocked(t *testing.T) {
	r := &record{
		Match: domain.Match{ID: "es-8", Kind: "fixture", Format: domain.FormatODI, TeamIDs: [2]string{"eng", "sl"}, State: domain.Live,
			Live: &domain.LiveScore{Innings: []domain.InningsScore{{TeamID: "eng", Runs: 386, Wickets: 5, Overs: "50"}, {TeamID: "sl", Runs: 130, Wickets: 2, Overs: "21.2"}}}},
		Event: esEvent{Competitors: []esCompetitor{{ID: "1", DisplayName: "England", Score: "386/5"}, {ID: "8", DisplayName: "Sri Lanka", Score: "130/2 (21.2/50 ov, target 387)"}}},
	}
	updateCalls(r, nil, func(string) string { return "" }, now)
	if r.call("es-8:total") != nil {
		t.Fatal("a total call for an innings that's over")
	}
	if c := r.call("es-8:chase"); c == nil || c.State != domain.CallLocked || c.Question != "Will Sri Lanka chase 387?" {
		t.Fatalf("chase = %+v", c)
	}
}

func TestPollMakesCallsForLiveMatches(t *testing.T) {
	f := &espnFake{calls: map[string]int{}}
	clock := now
	s := testSource(t, f, openStore(t), &clock)
	s.poll(context.Background())
	got := s.Calls("es-1550718")
	if len(got) == 0 || got[len(got)-1].ID != "es-1550718:chase" || got[len(got)-1].State != domain.CallLocked {
		t.Fatalf("ENG v SL calls = %+v", got)
	}
	if c, ok := s.Call("es-1550718:chase"); !ok || c.MatchID != "es-1550718" {
		t.Fatalf("Call = %+v %v", c, ok)
	}
	if len(s.Calls("es-1543945")) != 0 {
		t.Fatal("a finished match got calls")
	}
	if len(s.Calls("es-1525661")) != 0 {
		t.Fatal("a Test got a total or chase call")
	}
}

// Fans' own calls: refused once they're as good as decided, one card per question, settled like the app's.
func TestFanCallsAreFairAndSettle(t *testing.T) {
	clock := now
	s := testSource(t, &espnFake{calls: map[string]int{}}, openStore(t), &clock)
	r := &record{
		Match: domain.Match{ID: "es-7", Kind: "fixture", Format: domain.FormatT20, TeamIDs: [2]string{"afg", "zim"}, State: domain.Live,
			XI:   map[string][]string{"afg": {"esp-1", "esp-2"}, "zim": {"esp-5"}},
			Live: &domain.LiveScore{Innings: []domain.InningsScore{{TeamID: "afg", Runs: 44, Wickets: 1, Overs: "6.0"}}}},
		Event: esEvent{Competitors: []esCompetitor{{ID: "40", DisplayName: "Afghanistan"}, {ID: "9", DisplayName: "Zimbabwe"}}},
		card:  &battingCard{innings: "1", batters: []cardBatter{{id: "esp-1", runs: 42, hasBatted: true}, {id: "esp-2", runs: 12, hasBatted: true}}},
	}
	s.recs["es-7"] = r
	s.players["esp-2"] = domain.Player{ID: "esp-2", Name: "Ibrahim Zadran"}
	post := func(spec domain.CallSpec) (domain.LiveCall, error) {
		spec.MatchID, spec.By = "es-7", "u-1"
		return s.PostCall(spec, now)
	}

	if _, err := post(domain.CallSpec{Kind: "runs", PlayerID: "esp-1", Target: 50}); err != domain.ErrCallTooLate {
		t.Fatalf("a batter on 42 for 50: %v", err)
	}
	c, err := post(domain.CallSpec{Kind: "runs", PlayerID: "esp-2", Target: 50})
	if err != nil || c.Question != "Will Ibrahim Zadran make 50?" || c.State != domain.CallOpen || c.CreatedBy != "u-1" {
		t.Fatalf("Zadran 50: %+v %v", c, err)
	}
	if again, _ := post(domain.CallSpec{Kind: "runs", PlayerID: "esp-2", Target: 50}); again.ID != c.ID || len(r.Calls) != 1 {
		t.Fatal("the same call made twice")
	}
	for _, bad := range []domain.CallSpec{{Kind: "runs", PlayerID: "esp-9", Target: 50}, {Kind: "runs", PlayerID: "esp-2", Target: 45}, {Kind: "teamTotal", TeamID: "ind", Target: 180}, {Kind: "teamTotal", TeamID: "afg", Target: 20}, {Kind: "guess"}} {
		if _, err := post(bad); err != domain.ErrCallKind {
			t.Fatalf("%+v: %v", bad, err)
		}
	}
	total, err := post(domain.CallSpec{Kind: "teamTotal", TeamID: "afg", Target: 180})
	if err != nil || total.Question != "Will Afghanistan score 180 or more?" {
		t.Fatalf("total: %+v %v", total, err)
	}
	win, err := post(domain.CallSpec{Kind: "win", TeamID: "zim"})
	if err != nil || win.Question != "Will Zimbabwe win?" {
		t.Fatalf("win: %+v %v", win, err)
	}

	// Ten overs in, a total is too late to call; Zadran gets to 50; Afghanistan finish on 172; Zimbabwe win.
	r.Match.Live.Innings[0] = domain.InningsScore{TeamID: "afg", Runs: 80, Wickets: 1, Overs: "10.0"}
	updateCalls(r, &battingCard{innings: "1", batters: []cardBatter{{id: "esp-2", runs: 51, hasBatted: true}}}, s.playerName, now)
	if _, err := post(domain.CallSpec{Kind: "teamTotal", TeamID: "afg", Target: 200}); err != domain.ErrCallTooLate {
		t.Fatalf("total at 10 overs: %v", err)
	}
	if c := r.call(c.ID); c.State != domain.CallSettled || *c.AnswerID != "yes" {
		t.Fatalf("Zadran = %+v", c)
	}
	if c := r.call(total.ID); c.State != domain.CallLocked {
		t.Fatalf("total at 10 overs = %+v", c)
	}
	r.Match.Live.Innings = []domain.InningsScore{{TeamID: "afg", Runs: 172, Wickets: 6, Overs: "20"}, {TeamID: "zim", Runs: 40, Wickets: 0, Overs: "5.0"}}
	updateCalls(r, nil, s.playerName, now)
	if c := r.call(total.ID); c.State != domain.CallSettled || *c.AnswerID != "no" || c.Result != "Afghanistan made 172" {
		t.Fatalf("total settled = %+v", c)
	}
	if _, err := post(domain.CallSpec{Kind: "win", TeamID: "afg"}); err != domain.ErrCallTooLate {
		t.Fatalf("a winner five overs into the chase: %v", err)
	}
	r.Match.State, r.Match.Live = domain.Completed, nil
	r.Match.Result = &domain.MatchResult{WinnerID: domain.Ptr("zim"), Text: "Zimbabwe won by 4 wickets", Innings: []domain.InningsScore{{TeamID: "afg", Runs: 172, Wickets: 6, Overs: "20"}, {TeamID: "zim", Runs: 175, Wickets: 6, Overs: "19.3"}}}
	updateCalls(r, nil, s.playerName, now)
	if c := r.call(win.ID); c.State != domain.CallSettled || *c.AnswerID != "yes" {
		t.Fatalf("win settled = %+v", c)
	}
	if _, err := post(domain.CallSpec{Kind: "win", TeamID: "zim"}); err != domain.ErrCallTooLate {
		t.Fatalf("a finished match: %v", err)
	}
}
