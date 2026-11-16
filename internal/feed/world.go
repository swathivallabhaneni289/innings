package feed

import (
	"maps"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/sim"
)

// World is what the API serves when the feed is on: real matches from the Source, and the simulation for
// everything that isn't a match (the simulated fans, their reviews of Vault classics, the Vault itself). The
// simulation's made-up fixtures, and its fans' logs, predictions and crowds for them, are left out, so no
// invented score or opinion ever sits next to a real match.
type World struct {
	*sim.World
	src     *Source
	classic map[string]bool
}

func NewWorld(w *sim.World, src *Source) *World {
	classic := map[string]bool{}
	for _, m := range w.Classics() {
		classic[m.ID] = true
	}
	return &World{World: w, src: src, classic: classic}
}

func (w *World) Fixtures(time.Time) []domain.Match { return w.src.Fixtures() }

// Match finds a real match, a Vault classic or a past match from the archive.
func (w *World) Match(id string, now time.Time) (domain.Match, bool) {
	if w.classic[id] || w.World.Archived(id) {
		return w.World.Match(id, now)
	}
	for _, m := range w.src.Fixtures() {
		if m.ID == id {
			return m, true
		}
	}
	return domain.Match{}, false
}

func (w *World) Teams() map[string]domain.Team {
	out := w.World.Teams()
	maps.Copy(out, w.src.Teams())
	return out
}

func (w *World) Players() map[string]domain.Player {
	out := w.World.Players()
	maps.Copy(out, w.src.Players())
	return out
}

// Logs keeps the simulated fans' logs of Vault classics only.
func (w *World) Logs(now time.Time) []domain.MatchLog {
	out := []domain.MatchLog{}
	for _, l := range w.World.Logs(now) {
		if w.classic[l.MatchID] {
			out = append(out, l)
		}
	}
	return out
}

// Predictions: the simulated fans only ever predicted made-up fixtures.
func (w *World) Predictions(time.Time) []domain.Prediction { return []domain.Prediction{} }

// Crowd: the simulation's crowd for a classic, and nobody for a real match (real fans' ratings are added on top).
func (w *World) Crowd(matchID string) sim.Crowd {
	if w.classic[matchID] {
		return w.World.Crowd(matchID)
	}
	return sim.Crowd{RatingBuckets: make([]int, 10), PlayerTallies: map[string]sim.Tally{}, Watch: map[domain.WatchPlace]int{}, Picks: map[string]int{}}
}

func (w *World) Calls(matchID string) []domain.LiveCall { return w.src.Calls(matchID) }

func (w *World) Call(id string) (domain.LiveCall, bool) { return w.src.Call(id) }

func (w *World) PostCall(spec domain.CallSpec, now time.Time) (domain.LiveCall, error) {
	if w.classic[spec.MatchID] {
		return domain.LiveCall{}, domain.ErrCallsOff
	}
	return w.src.PostCall(spec, now)
}

// FriendsWatching: simulated fans don't watch real matches.
func (w *World) FriendsWatching(string, time.Time) []string { return []string{} }
