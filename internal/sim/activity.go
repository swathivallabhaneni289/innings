package sim

import (
	"maps"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// state is everything per-World that depends on the anchor: every simulated log, prediction and reply that
// will ever exist, with timestamps. Methods filter it by `now`.
type state struct {
	logs      []timedLog // sorted by createdAt ascending
	logByID   map[string]int
	preds     []timedPred
	replies   map[string][]timedReply
	watchByID map[string]domain.WatchPlace
}

type timedLog struct {
	at  time.Time
	log domain.MatchLog
}

type timedPred struct {
	at   time.Time
	pred domain.Prediction
}

type timedReply struct {
	at    time.Time
	reply domain.Reply
}

func ts(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func (w *World) st() *state {
	if w.s != nil {
		return w.s
	}
	return buildState(w.Anchor)
}

var watchWeights = []struct {
	p domain.WatchPlace
	w int
}{
	{domain.WatchHome, 42}, {domain.WatchFriends, 18}, {domain.WatchStadium, 9}, {domain.WatchHostel, 11},
	{domain.WatchCafe, 9}, {domain.WatchOffice, 6}, {domain.WatchOnTheGo, 5},
}

func pickPlace(r *rng, classic bool) domain.WatchPlace {
	total := 0
	for _, x := range watchWeights {
		total += x.w
	}
	n := r.intn(total)
	for _, x := range watchWeights {
		if n < x.w {
			if classic && x.p == domain.WatchStadium && r.chance(0.6) {
				return domain.WatchHome
			}
			return x.p
		}
		n -= x.w
	}
	return domain.WatchHome
}

// fixtureVals are the template values for a fixture.
func fixtureVals(f *fixture) map[string]string {
	v := map[string]string{
		"A": teamName(f.teams[f.inns[0].t]), "Z": teamName(f.teams[1-f.inns[0].t]),
		"T": playerNick(f.topBatIDs[0]), "R": strconv.Itoa(f.topRuns), "V": venueShort(f.venue), "M": f.margin,
	}
	if f.fav != f.potmID {
		v["F"] = playerNick(f.fav)
	}
	if f.winner >= 0 {
		v["W"], v["L"] = teamName(f.winnerID()), teamName(f.loserID())
	}
	if len(f.topBowlID) > 0 && f.topWkts > 0 {
		v["B"], v["K"] = playerNick(f.topBowlID[0]), strconv.Itoa(f.topWkts)
	}
	if f.potmID != "" {
		v["P"] = playerNick(f.potmID)
	}
	return v
}

func classicVals(c *classicEntry) map[string]string {
	s := c.seed
	v := map[string]string{"A": teamName(s.TeamIDs[0]), "Z": teamName(s.TeamIDs[1]), "H": s.Headline, "V": s.Venue.Name}
	if len(s.Date) >= 4 {
		v["Y"] = s.Date[:4]
	}
	if w, l := classicWinnerLoser(s); w != "" {
		v["W"], v["L"] = teamName(w), teamName(l)
	}
	if s.PotmName != "" {
		v["N"] = nick(s.PotmName)
	}
	return v
}

func (f *fixture) reviewVibe() string {
	if f.winner < 0 && !f.noResult {
		return "level"
	}
	return f.vibe
}

func fixtureRating(r *rng, f *fixture, favTeam string) float64 {
	base := 1.5 + 3.2*f.close
	switch f.vibe {
	case "century":
		base += 0.6
	case "bigchase":
		base += 0.7
	case "rain":
		base = 2
	}
	if f.winner >= 0 {
		if favTeam == f.winnerID() {
			base += 0.5
		} else if favTeam == f.loserID() {
			base -= 0.75
		}
	}
	return halfStep(base + r.frange(-0.75, 0.75))
}

func classicRating(r *rng, hype int) float64 {
	return halfStep(1.6 + 0.65*float64(hype) + r.frange(-0.6, 0.6))
}

// playerRatings rates 6–10 XI players, top performers higher; returns ratings and the fan's POTM.
func playerRatings(r *rng, f *fixture) (map[string]int, *string) {
	order := []string{f.fav}
	add := func(ids ...string) {
		for _, id := range ids {
			if id != "" && !slices.Contains(order, id) {
				order = append(order, id)
			}
		}
	}
	add(f.potmID)
	add(f.topBatIDs...)
	add(f.topBowlID...)
	all := append(append([]string{}, f.xi[f.teams[0]]...), f.xi[f.teams[1]]...)
	n := r.between(6, 10)
	for _, id := range shuffled(r, all) {
		if len(order) >= n {
			break
		}
		add(id)
	}
	out := map[string]int{}
	for _, id := range order {
		team := playerTable[id].TeamID
		var v int
		switch {
		case id == f.fav:
			v = r.between(9, 10)
		case id == f.potmID:
			v = r.between(7, 8)
		case slices.Contains(f.topBatIDs, id) || slices.Contains(f.topBowlID, id):
			v = r.between(7, 8)
		case f.winner >= 0 && team == f.winnerID():
			v = r.between(5, 8)
		default:
			v = r.between(3, 7)
		}
		out[id] = v
	}
	best := order[0]
	for _, id := range order {
		if out[id] > out[best] {
			best = id
		}
	}
	return out, domain.Ptr(best)
}

func companionsFor(r *rng, p domain.WatchPlace) *int {
	switch p {
	case domain.WatchFriends, domain.WatchStadium, domain.WatchHostel, domain.WatchCafe, domain.WatchOffice:
		return domain.Ptr(r.between(1, 14))
	case domain.WatchHome:
		if r.chance(0.5) {
			return domain.Ptr(r.between(1, 4))
		}
	}
	return nil
}

func noteFor(r *rng, pools map[string][]string, p domain.WatchPlace, venueName string) string {
	if !r.chance(0.75) {
		return ""
	}
	s, _ := fill(pick(r, pools[string(p)]), map[string]string{"V": venueName})
	return s
}

func buildState(anchor time.Time) *state {
	w := &World{Anchor: anchor}
	st := &state{logByID: map[string]int{}, replies: map[string][]timedReply{}, watchByID: map[string]domain.WatchPlace{}}
	logs := []timedLog{}

	// Fixture logs.
	for _, f := range fixtures {
		end := w.end(f)
		loggers := []string{}
		for _, fr := range friendRows {
			if newRNG("flog", fr.id, f.id).chance(0.55) {
				loggers = append(loggers, fr.id)
			}
		}
		pr := newRNG("plog", f.id)
		for _, p := range shuffled(pr, publicRows)[:pr.between(4, 8)] {
			loggers = append(loggers, p.id)
		}
		vals := fixtureVals(f)
		for _, uid := range loggers {
			r := newRNG("log", uid, f.id)
			l := domain.MatchLog{ID: "l-" + uid + "-" + f.id, UserID: uid, MatchID: f.id, WatchedWith: []string{}}
			l.Rating = fixtureRating(r, f, personByID[uid].team)
			l.InstantClassic = l.Rating >= 4.5 && r.chance(0.6)
			l.PlayerRatings, l.FanPotmID = playerRatings(r, f)
			l.WatchedAt = pickPlace(r, false)
			l.PlaceNote = noteFor(r, placeNotes, l.WatchedAt, venueShort(f.venue))
			l.Companions = companionsFor(r, l.WatchedAt)
			if isFriend(uid) && l.WatchedAt != domain.WatchHome && l.WatchedAt != domain.WatchOnTheGo && r.chance(0.45) {
				for _, o := range loggers {
					if o != uid && isFriend(o) {
						l.WatchedWith = append(l.WatchedWith, o)
						break
					}
				}
			}
			if !isFriend(uid) || r.chance(0.75) {
				l.Review = fillFrom(r, reviewPool[f.reviewVibe()], vals)
			}
			at := end.Add(time.Duration(r.between(30, 720)) * time.Minute)
			logs = append(logs, timedLog{at: at, log: l})
		}
	}

	// Classic logs (created in the weeks before the anchor).
	for k := range classicEntries {
		c := &classicEntries[k]
		loggers := []string{}
		for _, fr := range friendRows {
			if newRNG("fclog", fr.id, c.seed.ID).chance(0.2) {
				loggers = append(loggers, fr.id)
			}
		}
		pr := newRNG("pclog", c.seed.ID)
		for _, p := range shuffled(pr, publicRows)[:pr.between(4, 8)] {
			loggers = append(loggers, p.id)
		}
		vals := classicVals(c)
		for _, uid := range loggers {
			r := newRNG("clog", uid, c.seed.ID)
			l := domain.MatchLog{ID: "l-" + uid + "-" + c.seed.ID, UserID: uid, MatchID: c.seed.ID, WatchedWith: []string{}, PlayerRatings: map[string]int{}}
			l.Rating = classicRating(r, c.seed.Hype)
			l.InstantClassic = l.Rating >= 4.5 && r.chance(0.7)
			l.WatchedAt = pickPlace(r, true)
			l.PlaceNote = noteFor(r, classicPlaceNotes, l.WatchedAt, c.seed.Venue.Name)
			l.Companions = companionsFor(r, l.WatchedAt)
			if !isFriend(uid) || r.chance(0.75) {
				l.Review = fillFrom(r, reviewPool["classic"], vals)
			}
			at := anchor.Add(-time.Duration(r.between(60, 40*day)) * time.Minute)
			logs = append(logs, timedLog{at: at, log: l})
		}
	}

	sort.SliceStable(logs, func(a, b int) bool {
		if !logs[a].at.Equal(logs[b].at) {
			return logs[a].at.Before(logs[b].at)
		}
		return logs[a].log.ID < logs[b].log.ID
	})
	stub := map[string]int{}
	for k := range logs {
		l := &logs[k].log
		stub[l.UserID]++
		l.StubNo = stub[l.UserID]
		l.CreatedAt = ts(logs[k].at)
		st.logByID[l.ID] = k
		st.watchByID[l.ID] = l.WatchedAt
	}
	st.logs = logs

	// Replies.
	everyone := append(append([]personRow{}, friendRows...), publicRows...)
	for _, tl := range logs {
		r := newRNG("replies", tl.log.ID)
		n := []int{0, 0, 1, 1, 2, 3}[r.intn(6)]
		if tl.log.Review == "" {
			n = min(n, 1)
		}
		at := tl.at
		used := []string{}
		for k := 0; k < n; k++ {
			who := pick(r, everyone)
			if who.id == tl.log.UserID {
				continue
			}
			text := pick(r, replyPool)
			if slices.Contains(used, text) {
				continue
			}
			used = append(used, text)
			at = at.Add(time.Duration(r.between(10, 360)) * time.Minute)
			id := tl.log.ID + "-r" + strconv.Itoa(k+1)
			st.replies[tl.log.ID] = append(st.replies[tl.log.ID], timedReply{at: at, reply: domain.Reply{ID: id, LogID: tl.log.ID, UserID: who.id, Text: text, CreatedAt: ts(at)}})
		}
	}

	// Friends' predictions.
	for _, f := range fixtures {
		for _, fr := range friendRows {
			r := newRNG("pred", fr.id, f.id)
			if !r.chance(0.7) {
				continue
			}
			p := predictFor(r, f, fr.team)
			p.UserID = fr.id
			at := w.start(f).Add(-time.Duration(r.between(60, 30*hr)) * time.Minute)
			p.CreatedAt = ts(at)
			st.preds = append(st.preds, timedPred{at: at, pred: p})
		}
	}
	sort.SliceStable(st.preds, func(a, b int) bool {
		if !st.preds[a].at.Equal(st.preds[b].at) {
			return st.preds[a].at.Before(st.preds[b].at)
		}
		return st.preds[a].pred.UserID+st.preds[a].pred.MatchID < st.preds[b].pred.UserID+st.preds[b].pred.MatchID
	})
	return st
}

func predictFor(r *rng, f *fixture, favTeam string) domain.Prediction {
	team := f.teams[r.intn(2)]
	switch {
	case slices.Contains(f.teams[:], favTeam):
		team = favTeam
	case f.winner >= 0 && r.chance(0.6):
		team = f.winnerID()
	case f.winner >= 0:
		team = f.loserID()
	}
	other := f.teams[0]
	if other == team {
		other = f.teams[1]
	}
	bat := pick(r, batters(f.xi[team])[:5])
	if r.chance(0.25) {
		bat = f.topBatIDs[0]
	}
	bowl := pick(r, bowlers(f.xi[other]))
	if len(f.topBowlID) > 0 && r.chance(0.2) {
		bowl = f.topBowlID[0]
	}
	return domain.Prediction{MatchID: f.id, WinnerID: team, TopRunScorerID: bat, TopWicketTakerID: bowl}
}

func cloneLog(l domain.MatchLog) domain.MatchLog {
	c := l
	c.PlayerRatings = maps.Clone(l.PlayerRatings)
	if c.PlayerRatings == nil {
		c.PlayerRatings = map[string]int{}
	}
	c.WatchedWith = append([]string{}, l.WatchedWith...)
	if l.FanPotmID != nil {
		c.FanPotmID = domain.Ptr(*l.FanPotmID)
	}
	if l.Companions != nil {
		c.Companions = domain.Ptr(*l.Companions)
	}
	return c
}

func (w *World) logs(now time.Time) []domain.MatchLog {
	st := w.st()
	out := []domain.MatchLog{}
	for k := len(st.logs) - 1; k >= 0; k-- {
		if !st.logs[k].at.After(now) {
			out = append(out, cloneLog(st.logs[k].log))
		}
	}
	return out
}

func (w *World) predictions(now time.Time) []domain.Prediction {
	out := []domain.Prediction{}
	for _, p := range w.st().preds {
		if !p.at.After(now) {
			out = append(out, p.pred)
		}
	}
	return out
}

func baseLikes(logID string) int {
	r := newRNG("likes", logID)
	switch {
	case len(logID) > 4 && logID[:4] == "l-p-":
		f := r.float()
		return 12 + int(f*f*380)
	case len(logID) > 4 && logID[:4] == "l-f-":
		return 1 + int(r.float()*r.float()*30)
	}
	return 0
}

func (w *World) baseThereToo(logID string) int {
	place, ok := w.st().watchByID[logID]
	if !ok {
		return 0
	}
	r := newRNG("theretoo", logID)
	if place == domain.WatchStadium {
		return 3 + r.intn(58)
	}
	return r.intn(3)
}

func (w *World) replies(logID string, now time.Time) []domain.Reply {
	out := []domain.Reply{}
	for _, tr := range w.st().replies[logID] {
		if !tr.at.After(now) {
			out = append(out, tr.reply)
		}
	}
	return out
}
