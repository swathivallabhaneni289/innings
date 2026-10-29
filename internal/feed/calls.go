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
