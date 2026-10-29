package feed

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// Live calls are the predictions fans make while a match is on. They're opened, locked and settled here, from the
// same readings that drive the live score, so every poll moves them on:
//
//   - "Where will {team} finish?": the first innings of a limited-overs match, locked at 10 overs of a T20 (25 of
//     an ODI), settled when the innings ends.
//   - "Will {team} chase {target}?": the second innings, locked 5 overs in (10 in an ODI), settled by the result.
//   - "Will {batter} make 50?" (or 100): asked once a batter on the card reaches 25 (75), locked at 40 (90),
//     settled when they get there or get out. Only where ESPN publishes a batting card.
//
// A call is created once (its id says what it's about), its state only moves forward, and one first seen past its
// lock point is created locked, so a restart mid-innings or ESPN's two caches disagreeing for a minute never
// reopens a call. ESPN runs a minute or two behind television; that's the slack in every lock.

type band struct {
	id, label string
	lo, hi    int // runs in [lo, hi); hi 0 = no top
}

func bands(cuts [3]int) []band {
	return []band{
		{"lt" + strconv.Itoa(cuts[0]), fmt.Sprintf("Under %d", cuts[0]), 0, cuts[0]},
		{strconv.Itoa(cuts[0]), fmt.Sprintf("%d–%d", cuts[0], cuts[1]-1), cuts[0], cuts[1]},
		{strconv.Itoa(cuts[1]), fmt.Sprintf("%d–%d", cuts[1], cuts[2]-1), cuts[1], cuts[2]},
		{strconv.Itoa(cuts[2]) + "plus", fmt.Sprintf("%d or more", cuts[2]), cuts[2], 0},
	}
}

// totalBands are four ranges around a usual first-innings score for the format.
func totalBands(f domain.Format, g domain.Gender) []band {
	w := g == domain.Women
	switch {
	case f == domain.FormatODI && w:
		return bands([3]int{230, 260, 290})
	case f == domain.FormatODI:
		return bands([3]int{260, 290, 320})
	case f == domain.FormatHundred && w:
		return bands([3]int{120, 140, 160})
	case f == domain.FormatHundred:
		return bands([3]int{140, 160, 180})
	case w:
		return bands([3]int{130, 150, 170})
	}
	return bands([3]int{150, 170, 190})
}

// lockBalls is when a total (first innings) and a chase (second) call lock, in balls bowled in that innings.
func lockBalls(f domain.Format) (total, chase int) {
	switch f {
	case domain.FormatODI:
		return 25 * 6, 10 * 6
	case domain.FormatHundred:
		return 50, 25
	}
	return 10 * 6, 5 * 6
}

var targetText = regexp.MustCompile(`target (\d+)`)

func yesNo() []domain.CallOption {
	return []domain.CallOption{{ID: "yes", Label: "Yes"}, {ID: "no", Label: "No"}}
}

// battingCard is the summary's batting card for the innings in progress: who's batted, their runs, how they're out.
type battingCard struct {
	innings string
	batters []cardBatter
}

type cardBatter struct {
	id, name  string
	runs      int
	hasBatted bool
	out       bool
}

// card picks the latest batting card out of a summary, or nil when ESPN gives none.
func card(s esSummary) *battingCard {
	var best *battingCard
	bestN := -1
	for _, mc := range s.Matchcards {
		if !strings.EqualFold(mc.Headline, "Batting") {
			continue
		}
		n, _ := strconv.Atoi(string(mc.InningsNumber))
		if n < bestN {
			continue
		}
		c := &battingCard{innings: string(mc.InningsNumber)}
		for _, p := range mc.PlayerDetails {
			if p.PlayerID == "" {
				continue
			}
			b := cardBatter{id: playerPrefix + string(p.PlayerID), name: p.PlayerName}
			if r, err := strconv.Atoi(strings.TrimSpace(string(p.Runs))); err == nil {
				b.runs, b.hasBatted = r, true
			}
			d := strings.ToLower(strings.TrimSpace(p.Dismissal))
			b.out = b.hasBatted && d != "" && d != "not out"
			c.batters = append(c.batters, b)
		}
		best, bestN = c, n
	}
	return best
}

func (r *record) call(id string) *domain.LiveCall {
	for k := range r.Calls {
		if r.Calls[k].ID == id {
			return &r.Calls[k]
		}
	}
	return nil
}

func (r *record) open(c domain.LiveCall, locked bool, now time.Time) *domain.LiveCall {
	c.MatchID, c.State, c.OpenedAt = r.Match.ID, domain.CallOpen, now.UTC().Format(time.RFC3339)
	if locked {
		c.State = domain.CallLocked
	}
	r.Calls = append(r.Calls, c)
	return &r.Calls[len(r.Calls)-1]
}

func lock(c *domain.LiveCall) {
	if c.State == domain.CallOpen {
		c.State = domain.CallLocked
	}
}

func settle(c *domain.LiveCall, answer, result string) {
	if c.State == domain.CallOpen || c.State == domain.CallLocked {
		c.State, c.AnswerID, c.Result = domain.CallSettled, &answer, result
	}
}

func void(c *domain.LiveCall, why string) {
	if c.State == domain.CallOpen || c.State == domain.CallLocked {
		c.State, c.Result = domain.CallVoid, why
	}
}

// updateCalls moves a match's calls on from its latest reading. bc is the current batting card when this reading
// came with a summary (nil from the scoreboard alone); name gives a player's display name.
func updateCalls(r *record, bc *battingCard, name func(id string) string, now time.Time) {
	m := r.Match
	if m.Kind != "fixture" || m.State == domain.Upcoming {
		return
	}
	var innings []domain.InningsScore
	switch {
	case m.Live != nil:
		innings = m.Live.Innings
	case m.Result != nil:
		innings = m.Result.Innings
	}
	done := m.State == domain.Completed
	team := func(id string) string { return teamName(id, r) }

	if m.Format != domain.FormatTest && len(innings) > 0 {
		totalLock, chaseLock := lockBalls(m.Format)
		first := innings[0]

		// Where will the side batting first finish?
		id := m.ID + ":total"
		c := r.call(id)
		if c == nil && !done && len(innings) == 1 {
			opts := []domain.CallOption{}
			for _, b := range totalBands(m.Format, m.Gender) {
				opts = append(opts, domain.CallOption{ID: b.id, Label: b.label})
			}
			c = r.open(domain.LiveCall{ID: id, Kind: "total", Question: fmt.Sprintf("Where will %s finish?", team(first.TeamID)), Options: opts,
				LocksAt: fmt.Sprintf("at %s overs", overs(float64(totalLock/6)))}, balls(first.Overs) >= totalLock, now)
		}
		if c != nil {
			if balls(first.Overs) >= totalLock || len(innings) > 1 || done {
				lock(c)
			}
			if len(innings) > 1 || done {
				for _, b := range totalBands(m.Format, m.Gender) {
					if first.Runs >= b.lo && (b.hi == 0 || first.Runs < b.hi) {
						settle(c, b.id, fmt.Sprintf("%s made %d", team(first.TeamID), first.Runs))
					}
				}
			}
		}

		// Will the side batting second get there?
		if len(innings) > 1 && innings[1].TeamID != first.TeamID {
			chasing := innings[1]
			id := m.ID + ":chase"
			c := r.call(id)
			if c == nil && !done {
				target := first.Runs + 1
				for _, ec := range r.Event.Competitors {
					if g := targetText.FindStringSubmatch(ec.Score); g != nil {
						target, _ = strconv.Atoi(g[1]) // a rain-reduced target too
					}
				}
				c = r.open(domain.LiveCall{ID: id, Kind: "chase", Question: fmt.Sprintf("Will %s chase %d?", team(chasing.TeamID), target), Options: yesNo(),
					LocksAt: fmt.Sprintf("%s overs into the chase", overs(float64(chaseLock/6)))}, balls(chasing.Overs) >= chaseLock, now)
			}
			if c != nil {
				if balls(chasing.Overs) >= chaseLock || done {
					lock(c)
				}
				if done && m.Result != nil {
					switch w := m.Result.WinnerID; {
					case w == nil:
						void(c, cmpOr(m.Result.Text, "No result"))
					case *w == chasing.TeamID:
						settle(c, "yes", m.Result.Text)
					default:
						settle(c, "no", m.Result.Text)
					}
				}
			}
		}
	}

	// Will a batter on the card make 50 (or 100)?
	if bc != nil {
		prefix := m.ID + ":ms:" + bc.innings + ":"
		for _, b := range bc.batters {
			if !b.hasBatted {
				continue
			}
			for _, target := range []int{50, 100} {
				id := prefix + strings.TrimPrefix(b.id, playerPrefix) + ":" + strconv.Itoa(target)
				c := r.call(id)
				if c == nil && !done && !b.out && b.runs >= target-25 && b.runs < target-10 {
					who := cmpOr(name(b.id), b.name)
					c = r.open(domain.LiveCall{ID: id, Kind: "milestone", Question: fmt.Sprintf("Will %s make %d?", who, target), Options: yesNo(),
						LocksAt: fmt.Sprintf("at %d runs", target-10)}, false, now)
				}
				if c == nil {
					continue
				}
				who := cmpOr(name(b.id), b.name)
				switch {
				case b.runs >= target:
					settle(c, "yes", fmt.Sprintf("%s got to %d", who, target))
				case b.out:
					settle(c, "no", fmt.Sprintf("%s out for %d", who, b.runs))
				case b.runs >= target-10:
					lock(c)
				}
			}
		}
	}
	updateFanCalls(r, bc, innings, done)
	// A batter's call from an innings that has ended without an answer, or any left when the match is over: void.
	for k := range r.Calls {
		c := &r.Calls[k]
		if c.Kind != "milestone" || (c.State != domain.CallOpen && c.State != domain.CallLocked) {
			continue
		}
		if done || (bc != nil && !strings.HasPrefix(c.ID, m.ID+":ms:"+bc.innings+":")) {
			void(c, "The innings ended before it was settled")
		}
	}
}

func teamName(id string, r *record) string {
	for k, ec := range r.Event.Competitors {
		if k < 2 && r.Match.TeamIDs[k] == id {
			return ec.DisplayName
		}
	}
	return strings.ToUpper(id)
}

// callsOf returns a copy of a record's calls, newest first.
func callsOf(r *record) []domain.LiveCall {
	out := make([]domain.LiveCall, len(r.Calls))
	for k, c := range r.Calls {
		c.Options = slices.Clone(c.Options)
		out[len(r.Calls)-1-k] = c
	}
	return out
}

// ---- fans' own calls ----

func liveInnings(m domain.Match) []domain.InningsScore {
	switch {
	case m.Live != nil:
		return m.Live.Innings
	case m.Result != nil:
		return m.Result.Innings
	}
	return nil
}

func inningsOf(innings []domain.InningsScore, team string) int {
	return slices.IndexFunc(innings, func(i domain.InningsScore) bool { return i.TeamID == team })
}

func inXI(m domain.Match, playerID string) bool {
	for _, ids := range m.XI {
		if slices.Contains(ids, playerID) {
			return true
		}
	}
	return false
}

// PostCall makes a fan's own call on a real match, or says why it can't be made: the same calls the app makes, on
// the fan's choice of batter, team and target, and only while it's still a real question (not for a batter
// already within 10 of it, a total once the innings is past its lock, a winner once the chase is under way). The
// same call asked twice is one card, so fans pile onto it.
func (s *Source) PostCall(spec domain.CallSpec, now time.Time) (domain.LiveCall, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.recs[spec.MatchID]
	if r == nil || len(r.Event.Competitors) < 2 {
		return domain.LiveCall{}, domain.ErrCallsOff
	}
	m := r.Match
	if m.State == domain.Completed {
		return domain.LiveCall{}, domain.ErrCallTooLate
	}
	innings := liveInnings(m)
	totalLock, chaseLock := lockBalls(m.Format)
	limited := m.Format != domain.FormatTest
	team := func(id string) string { return teamName(id, r) }
	isTeam := spec.TeamID == m.TeamIDs[0] || spec.TeamID == m.TeamIDs[1]

	c := domain.LiveCall{Kind: spec.Kind, Options: yesNo(), CreatedBy: spec.By, PlayerID: spec.PlayerID, TeamID: spec.TeamID, Target: spec.Target}
	switch spec.Kind {
	case "runs":
		if (spec.Target != 30 && spec.Target != 50 && spec.Target != 100) || !inXI(m, spec.PlayerID) {
			return domain.LiveCall{}, domain.ErrCallKind
		}
		if r.card != nil {
			for _, b := range r.card.batters {
				if b.id == spec.PlayerID && b.hasBatted && (b.out || b.runs >= spec.Target-10) {
					return domain.LiveCall{}, domain.ErrCallTooLate
				}
			}
		}
		c.ID = fmt.Sprintf("%s:fan:runs:%s:%d", m.ID, strings.TrimPrefix(spec.PlayerID, playerPrefix), spec.Target)
		c.Question = fmt.Sprintf("Will %s make %d?", cmpOr(s.players[spec.PlayerID].Name, "he"), spec.Target)
		c.LocksAt = fmt.Sprintf("at %d runs", spec.Target-10)
	case "teamTotal":
		if !limited || !isTeam || spec.Target < 50 || spec.Target > 500 {
			return domain.LiveCall{}, domain.ErrCallKind
		}
		if k := inningsOf(innings, spec.TeamID); k >= 0 && (k < len(innings)-1 || balls(innings[k].Overs) >= totalLock) {
			return domain.LiveCall{}, domain.ErrCallTooLate
		}
		c.ID = fmt.Sprintf("%s:fan:total:%s:%d", m.ID, spec.TeamID, spec.Target)
		c.Question = fmt.Sprintf("Will %s score %d or more?", team(spec.TeamID), spec.Target)
		c.LocksAt = fmt.Sprintf("at %s overs of their innings", overs(float64(totalLock/6)))
	case "win":
		if !isTeam {
			return domain.LiveCall{}, domain.ErrCallKind
		}
		if (limited && len(innings) >= 2 && balls(innings[1].Overs) >= chaseLock) || (!limited && len(innings) >= 3) {
			return domain.LiveCall{}, domain.ErrCallTooLate
		}
		c.ID = fmt.Sprintf("%s:fan:win:%s", m.ID, spec.TeamID)
		c.Question = fmt.Sprintf("Will %s win?", team(spec.TeamID))
		c.LocksAt = fmt.Sprintf("%s overs into the chase", overs(float64(chaseLock/6)))
		if !limited {
			c.LocksAt = "when the third innings starts"
		}
	default:
		return domain.LiveCall{}, domain.ErrCallKind
	}
	if old := r.call(c.ID); old != nil {
		out := *old
		out.Options = slices.Clone(old.Options)
		return out, nil
	}
	made := r.open(c, false, now)
	s.dirty[m.ID] = true
	out := *made
	out.Options = slices.Clone(made.Options)
	return out, nil
}

// updateFanCalls moves fans' own calls on, the same way as the app's: a batter's score from the batting card (it
// counts in the first innings they bat in after it was made), a team's total when that innings ends, the winner
// from the result.
func updateFanCalls(r *record, bc *battingCard, innings []domain.InningsScore, done bool) {
	m := r.Match
	totalLock, chaseLock := lockBalls(m.Format)
	limited := m.Format != domain.FormatTest
	for k := range r.Calls {
		c := &r.Calls[k]
		if c.CreatedBy == "" || (c.State != domain.CallOpen && c.State != domain.CallLocked) {
			continue
		}
		switch c.Kind {
		case "runs":
			if bc != nil {
				for _, b := range bc.batters {
					if b.id != c.PlayerID || !b.hasBatted {
						continue
					}
					if c.Innings == "" {
						c.Innings = bc.innings
					}
					if c.Innings != bc.innings {
						continue
					}
					switch {
					case b.runs >= c.Target:
						settle(c, "yes", fmt.Sprintf("He got to %d", c.Target))
					case b.out:
						settle(c, "no", fmt.Sprintf("Out for %d", b.runs))
					case b.runs >= c.Target-10:
						lock(c)
					}
				}
				if c.Innings != "" && c.Innings != bc.innings {
					void(c, "The innings ended before it was settled")
				}
			}
			if done {
				void(c, "The match ended before it was settled")
			}
		case "teamTotal":
			if i := inningsOf(innings, c.TeamID); i >= 0 {
				if balls(innings[i].Overs) >= totalLock {
					lock(c)
				}
				if i < len(innings)-1 || done {
					answer := "no"
					if innings[i].Runs >= c.Target {
						answer = "yes"
					}
					settle(c, answer, fmt.Sprintf("%s made %d", teamName(c.TeamID, r), innings[i].Runs))
				}
			} else if done {
				void(c, "They didn't bat")
			}
		case "win":
			if (limited && len(innings) >= 2 && balls(innings[1].Overs) >= chaseLock) || (!limited && len(innings) >= 3) {
				lock(c)
			}
			if done && m.Result != nil {
				switch w := m.Result.WinnerID; {
				case w == nil:
					void(c, cmpOr(m.Result.Text, "No result"))
				case *w == c.TeamID:
					settle(c, "yes", m.Result.Text)
				default:
					settle(c, "no", m.Result.Text)
				}
			}
		}
	}
}
