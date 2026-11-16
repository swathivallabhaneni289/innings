// Package feed brings real cricket into Innings: fixtures, live scores, squads, XIs and results from ESPN's public
// scores feed (espn.go). One goroutine polls it for the whole server, keeps what it learns in SQLite, and serves it
// from memory, so the app's requests never reach ESPN. It is on unless INNINGS_MATCHES=sim, which runs the
// simulated world instead.
package feed

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

const (
	busyEvery    = time.Minute     // scoreboard poll while a match is on or about to start
	idleEvery    = 5 * time.Minute // and otherwise
	aheadDays    = 7               // the next week's fixtures, one scoreboard call per day…
	sweepEvery   = 6 * time.Hour   // …every six hours
	backDays     = 14              // after time off, the most days caught up (ESPN only answers for recent days)
	liveSummary  = 2 * time.Minute // a live match's summary (XIs, per-innings lines)
	squadSummary = 6 * time.Hour   // an upcoming match's summary (announced squads), all week
	squadAhead   = aheadDays * 24 * time.Hour
	tossWindow   = 2 * time.Hour    // from this close to the start…
	tossSummary  = 5 * time.Minute  // …check every five minutes for the toss and the playing XIs
	finalEvery   = 10 * time.Minute // a finished match's summary, until it has the leaders and player of the match
	finalTries   = 6
	staleAfter   = 30 * time.Minute // a live match missing from the scoreboard this long is checked directly
	perPoll      = 12               // summaries fetched per poll at most
)

// record is one real match and what's been fetched for it.
type record struct {
	Match     domain.Match      `json:"match"`
	Event     esEvent           `json:"event"` // the scoreboard's last word on it: the feed's team and league ids
	SeenAt    time.Time         `json:"seenAt"`
	SummaryAt time.Time         `json:"summaryAt"`
	Final     bool              `json:"final"` // a finished match's summary is in, leaders and all
	Tries     int               `json:"tries"` // summary fetches since it finished
	Calls     []domain.LiveCall `json:"calls"` // predictions during the match (calls.go), oldest first

	card *battingCard // the latest batting card, for checking a fan's call is still fair (not saved)
}

func (r *record) espnTeams() [2]string {
	return [2]string{r.Event.Competitors[0].ID, r.Event.Competitors[1].ID}
}

// Source is the real match feed. Its read methods are safe for concurrent use while Run polls.
type Source struct {
	c   *client
	st  *store.Store
	log *slog.Logger
	now func() time.Time

	mu        sync.RWMutex
	cat       *catalog
	recs      map[string]*record
	players   map[string]domain.Player
	feedTeams map[string]domain.Team // teams the feed added (the rest come from the simulation)
	dirty     map[string]bool        // match ids to save
	newTeams  map[string]bool
	newPlayer map[string]bool
	lastSweep time.Time // last look at the week ahead (a restart looks again straight away)
	seen      time.Time // last scoreboard read; after a restart, the newest match's SeenAt stands in for it
}

// FromEnv returns the real feed unless INNINGS_MATCHES is "sim" (then nil: run the simulation). known is the
// simulation's team table, so real matches between those sides keep their ids and colours.
func FromEnv(getenv func(string) string, st *store.Store, known map[string]domain.Team, logger *slog.Logger) *Source {
	if strings.EqualFold(strings.TrimSpace(getenv("INNINGS_MATCHES")), "sim") {
		return nil
	}
	return newSource(newClient(), st, known, time.Now, logger)
}

func newSource(c *client, st *store.Store, known map[string]domain.Team, now func() time.Time, logger *slog.Logger) *Source {
	return &Source{
		c: c, st: st, log: logger, now: now, cat: newCatalog(known),
		recs: map[string]*record{}, players: map[string]domain.Player{}, feedTeams: map[string]domain.Team{},
		dirty: map[string]bool{}, newTeams: map[string]bool{}, newPlayer: map[string]bool{},
	}
}

// Load reads what earlier runs saved. Call it before serving.
func (s *Source) Load(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	teams, err := s.st.FeedRows(ctx, store.FeedTeam)
	if err != nil {
		return err
	}
	for id, body := range teams {
		var t domain.Team
		if err := json.Unmarshal(body, &t); err != nil {
			return fmt.Errorf("feed team %s: %w", id, err)
		}
		s.feedTeams[t.ID] = t
		s.cat.add(t)
	}
	players, err := s.st.FeedRows(ctx, store.FeedPlayer)
	if err != nil {
		return err
	}
	for id, body := range players {
		var p domain.Player
		if err := json.Unmarshal(body, &p); err != nil {
			return fmt.Errorf("feed player %s: %w", id, err)
		}
		s.players[p.ID] = p
	}
	matches, err := s.st.FeedRows(ctx, store.FeedMatch)
	if err != nil {
		return err
	}
	for id, body := range matches {
		var r record
		if err := json.Unmarshal(body, &r); err != nil {
			return fmt.Errorf("feed match %s: %w", id, err)
		}
		if len(r.Event.Competitors) == 2 {
			s.recs[r.Match.ID] = &r
			if r.SeenAt.After(s.seen) {
				s.seen = r.SeenAt
			}
		}
	}
	s.log.Info("feed: loaded saved matches", "matches", len(s.recs), "teams", len(s.feedTeams), "players", len(s.players))
	return nil
}

// ---- reads (what the API serves) ----

// Fixtures returns every real match the server knows, by start time. Matches stay after the feed stops listing
// them, so fans' logs and predictions always find theirs.
func (s *Source) Fixtures() []domain.Match {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Match, 0, len(s.recs))
	for _, r := range s.recs {
		out = append(out, cloneMatch(r.Match))
	}
	slices.SortStableFunc(out, func(a, b domain.Match) int {
		return cmp.Or(strings.Compare(a.StartTime, b.StartTime), strings.Compare(a.ID, b.ID))
	})
	return out
}

// Teams returns the teams the feed added (the simulation's table covers the rest).
func (s *Source) Teams() map[string]domain.Team {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.feedTeams)
}

// Players returns every player in a real match's squads, XIs or leaders.
func (s *Source) Players() map[string]domain.Player {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.players)
}

func cloneMatch(m domain.Match) domain.Match {
	if m.XI != nil {
		xi := make(map[string][]string, len(m.XI))
		for k, v := range m.XI {
			xi[k] = slices.Clone(v)
		}
		m.XI = xi
	}
	if m.Live != nil {
		l := *m.Live
		l.Innings = slices.Clone(l.Innings)
		m.Live = &l
	}
	if m.Result != nil {
		r := *m.Result
		r.Innings = slices.Clone(r.Innings)
		r.TopRunScorerIDs = slices.Clone(r.TopRunScorerIDs)
		r.TopWicketTakerIDs = slices.Clone(r.TopWicketTakerIDs)
		m.Result = &r
	}
	return m
}

// ---- polling ----

// Run polls until ctx ends: once a minute while a match is on or about to start, every five minutes otherwise.
func (s *Source) Run(ctx context.Context) {
	for {
		wait := s.poll(ctx)
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
	}
}

// poll does one round and returns how long to wait before the next.
func (s *Source) poll(ctx context.Context) time.Duration {
	now := s.now()
	events, err := s.c.scoreboard(ctx, "")
	if err != nil {
		s.log.Warn("feed: scoreboard failed", "err", err)
	} else {
		s.merge(events, now)
		s.catchUp(ctx, now)
		s.seen = now
	}
	if now.Sub(s.lastSweep) >= sweepEvery {
		// The week ahead, so fixtures (and their squads, once announced) show up days before the first ball.
		ok := true
		for d := 1; d <= aheadDays; d++ {
			day, err := s.c.scoreboard(ctx, now.UTC().AddDate(0, 0, d).Format("20060102"))
			if err != nil {
				s.log.Warn("feed: week-ahead scoreboard failed", "day", d, "err", err)
				ok = false
				continue
			}
			s.merge(day, now)
		}
		if ok {
			s.lastSweep = now
		}
	}
	for _, id := range s.wantSummary(now) {
		s.fetchSummary(ctx, id, now)
	}
	if err := s.save(ctx, now); err != nil {
		s.log.Error("feed: save", "err", err)
	}
	wait := idleEvery
	if s.busy(now) {
		wait = busyEvery
	}
	s.log.Info("feed: polled", "matches", s.count(), "next", wait)
	return wait
}

// catchUp asks for each day missed since the last scoreboard read (the server was stopped, or the Mac slept), from
// that day to yesterday: the scoreboard itself only lists today's and the latest matches. It looks back backDays at
// most, which is also what a fresh database gets. A day that fails is logged, not retried.
func (s *Source) catchUp(ctx context.Context, now time.Time) {
	today := now.UTC().Truncate(24 * time.Hour)
	from := today.AddDate(0, 0, -backDays)
	if last := s.seen.UTC().Truncate(24 * time.Hour); last.After(from) {
		from = last
	}
	for d := from; d.Before(today); d = d.AddDate(0, 0, 1) {
		events, err := s.c.scoreboard(ctx, d.Format("20060102"))
		if err != nil {
			s.log.Warn("feed: catch-up scoreboard failed", "day", d.Format(time.DateOnly), "err", err)
			continue
		}
		s.merge(events, now)
	}
}

func (s *Source) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.recs)
}

// busy: a match is live or starts within the hour.
func (s *Source) busy(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.recs {
		if r.Match.State == domain.Live {
			return true
		}
		if start := startOf(r.Match); r.Match.State == domain.Upcoming && start.After(now.Add(-time.Hour)) && start.Before(now.Add(time.Hour)) {
			return true
		}
	}
	return false
}

func startOf(m domain.Match) time.Time {
	t, _ := time.Parse(time.RFC3339, m.StartTime)
	return t
}

// merge folds the scoreboard into the records, keeping what summaries already added.
func (s *Source) merge(events []esEvent, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range events {
		m, added, ok := s.cat.toMatch(e)
		if !ok {
			continue
		}
		for _, t := range added {
			s.cat.add(t)
			s.feedTeams[t.ID] = t
			s.newTeams[t.ID] = true
		}
		r := s.recs[m.ID]
		if r == nil {
			r = &record{}
			s.recs[m.ID] = r
		}
		keep(&m, r.Match)
		r.Match, r.Event, r.SeenAt = m, e, now
		updateCalls(r, nil, s.playerName, now)
		s.dirty[m.ID] = true
	}
}

// keep carries what summaries added (XIs, the venue's country, overs per innings, leaders, player of the match)
// from the old version of a match onto a fresh one from the scoreboard.
func keep(m *domain.Match, old domain.Match) {
	if old.ID == "" {
		return
	}
	if len(m.XI) == 0 && len(old.XI) > 0 {
		m.XI = old.XI
	}
	if m.Venue.Country == "" {
		m.Venue.Country = old.Venue.Country
		if old.Venue.City != "" {
			m.Venue.City = old.Venue.City
		}
	}
	oldInnings := func() []domain.InningsScore {
		switch {
		case old.Live != nil:
			return old.Live.Innings
		case old.Result != nil:
			return old.Result.Innings
		}
		return nil
	}()
	fill := func(innings []domain.InningsScore) {
		for k := range innings {
			if innings[k].Overs == "" && k < len(oldInnings) && oldInnings[k].TeamID == innings[k].TeamID && oldInnings[k].Runs == innings[k].Runs {
				innings[k].Overs = oldInnings[k].Overs
			}
		}
	}
	if m.Live != nil && old.Live != nil {
		m.Live.Innings = furthest(m.Live.Innings, old.Live.Innings) // never wind a live score back
	}
	if m.Result != nil {
		fill(m.Result.Innings)
		if old.Result != nil {
			if len(old.Result.TopRunScorerIDs) > 0 || len(old.Result.TopWicketTakerIDs) > 0 {
				m.Result.TopRunScorerIDs, m.Result.TopWicketTakerIDs = old.Result.TopRunScorerIDs, old.Result.TopWicketTakerIDs
			}
			if m.Result.OfficialPotmID == nil {
				m.Result.OfficialPotmID = old.Result.OfficialPotmID
			}
			if m.Result.WinnerID == nil {
				m.Result.WinnerID = old.Result.WinnerID
			}
		}
	}
}

// wantSummary lists the matches whose summary is due, live ones first, then by start, at most perPoll.
func (s *Source) wantSummary(now time.Time) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ids []string
	for id, r := range s.recs {
		since := now.Sub(r.SummaryAt)
		due := false
		switch r.Match.State {
		case domain.Live:
			due = since >= liveSummary
		case domain.Upcoming:
			// Squads every six hours in the week before (hourly while none are out), then every five minutes
			// from two hours out, so the playing XIs land as soon as the toss is done.
			start := startOf(r.Match)
			switch {
			case start.Before(now.Add(tossWindow)):
				due = since >= tossSummary
			case start.Before(now.Add(squadAhead)):
				due = since >= squadSummary || (len(r.Match.XI) == 0 && since >= time.Hour)
			}
		case domain.Completed:
			due = !r.Final && r.Tries < finalTries && since >= finalEvery
		}
		if due {
			ids = append(ids, id)
		}
	}
	slices.SortFunc(ids, func(a, b string) int {
		ra, rb := s.recs[a].Match, s.recs[b].Match
		if la, lb := ra.State == domain.Live, rb.State == domain.Live; la != lb {
			if la {
				return -1
			}
			return 1
		}
		return strings.Compare(ra.StartTime, rb.StartTime)
	})
	return ids[:min(perPoll, len(ids))]
}

func (s *Source) fetchSummary(ctx context.Context, id string, now time.Time) {
	s.mu.RLock()
	r := s.recs[id]
	leagueID, eventID := r.Event.LeagueID, r.Event.ID
	s.mu.RUnlock()
	sum, err := s.c.summary(ctx, leagueID, eventID)

	s.mu.Lock()
	defer s.mu.Unlock()
	r.SummaryAt = now
	s.dirty[id] = true
	if r.Match.State == domain.Completed {
		r.Tries++
	}
	if err != nil {
		s.log.Warn("feed: summary failed", "match", id, "err", err)
		return
	}
	// A live match the scoreboard stopped listing: the summary says whether it has finished.
	if r.Match.State == domain.Live && now.Sub(r.SeenAt) >= staleAfter && len(sum.Header.Competitions) > 0 && sum.Header.Competitions[0].Status.Type.State == "post" {
		res := &domain.MatchResult{Text: r.Match.Live.StatusText, Innings: r.Match.Live.Innings, TopRunScorerIDs: []string{}, TopWicketTakerIDs: []string{}}
		espn := r.espnTeams()
		for _, c := range sum.Header.Competitions[0].Competitors {
			for k := range espn {
				if c.Winner && c.ID == espn[k] {
					res.WinnerID = &r.Match.TeamIDs[k]
				}
			}
		}
		r.Match.State, r.Match.Live, r.Match.Result = domain.Completed, nil, res
	}
	for _, p := range applySummary(&r.Match, sum, r.espnTeams()) {
		// The first sighting wins, except a later one that knows the player's side.
		if old, known := s.players[p.ID]; !known || (old.TeamID == "" && p.TeamID != "") {
			s.players[p.ID] = p
			s.newPlayer[p.ID] = true
		}
	}
	if res := r.Match.Result; res != nil {
		r.Final = res.OfficialPotmID != nil && len(res.TopRunScorerIDs) > 0
	}
	r.card = card(sum)
	updateCalls(r, r.card, s.playerName, now)
}

// playerName is a player's display name ("" if unknown). Callers hold s.mu.
func (s *Source) playerName(id string) string { return s.players[id].Name }

// Calls returns a match's live calls, newest first.
func (s *Source) Calls(matchID string) []domain.LiveCall {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if r := s.recs[matchID]; r != nil {
		return callsOf(r)
	}
	return []domain.LiveCall{}
}

// Call finds one live call by id (a call's id starts with its match's).
func (s *Source) Call(id string) (domain.LiveCall, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	matchID, _, _ := strings.Cut(id, ":")
	if r := s.recs[matchID]; r != nil {
		if c := r.call(id); c != nil {
			out := *c
			out.Options = slices.Clone(c.Options)
			return out, true
		}
	}
	return domain.LiveCall{}, false
}

// save writes what changed since the last save.
func (s *Source) save(ctx context.Context, now time.Time) error {
	s.mu.Lock()
	matches, teams, players := map[string][]byte{}, map[string][]byte{}, map[string][]byte{}
	var err error
	put := func(dst map[string][]byte, id string, v any) {
		b, e := json.Marshal(v)
		if e != nil {
			err = e
			return
		}
		dst[id] = b
	}
	for id := range s.dirty {
		put(matches, id, s.recs[id])
	}
	for id := range s.newTeams {
		put(teams, id, s.feedTeams[id])
	}
	for id := range s.newPlayer {
		put(players, id, s.players[id])
	}
	clear(s.dirty)
	clear(s.newTeams)
	clear(s.newPlayer)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	for kind, rows := range map[string]map[string][]byte{store.FeedTeam: teams, store.FeedPlayer: players, store.FeedMatch: matches} {
		if err := s.st.SaveFeedRows(ctx, kind, rows, now); err != nil {
			return err
		}
	}
	return nil
}
