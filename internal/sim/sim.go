// Package sim is the deterministic simulated cricket world: fixtures around server start, live scores,
// results, simulated friends and public fans, their logs/predictions, and the anonymous crowd.
// CONTRACT: the exported signatures in this file are used by internal/httpapi. Keep them; fill the bodies.
package sim

import (
	"maps"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// Tally is a running sum of 1–10 player ratings.
type Tally struct {
	Sum   int
	Count int
}

// Crowd is the anonymous (non-friend, non-public-reviewer) fan base for one match.
type Crowd struct {
	RatingBuckets []int                     // len 10: counts for 0.5, 1, … 5 balls
	PlayerTallies map[string]Tally          // playerId → ratings out of 10
	Watch         map[domain.WatchPlace]int // where the crowd watched
	Picks         map[string]int            // teamId → fans who predicted that winner (frozen at lock time)
}

// World is safe for concurrent reads; it never mutates after NewWorld.
type World struct {
	Anchor time.Time // server start, rounded down to the minute; every fixture time is relative to it
	s      *state    // precomputed activity (nil for a literal World; rebuilt on demand)
}

// NewWorld builds the world. Same anchor → same world (IDs never depend on the anchor).
func NewWorld(anchor time.Time) *World {
	w := &World{Anchor: anchor.Truncate(time.Minute)}
	w.s = buildState(w.Anchor)
	return w
}

// Teams returns every team any fixture or classic references, keyed by id.
func (w *World) Teams() map[string]domain.Team { return maps.Clone(teamTable) }

// Players returns every player in any fixture XI, keyed by id.
func (w *World) Players() map[string]domain.Player { return maps.Clone(playerTable) }

// Fixtures returns all simulated fixtures with State/Live/Result computed for `now`, sorted by StartTime.
func (w *World) Fixtures(now time.Time) []domain.Match {
	out := []domain.Match{}
	for _, f := range fixtures {
		out = append(out, w.fixtureMatch(f, now))
	}
	return out
}

// Classics returns the Vault archive (real historical matches), oldest first.
func (w *World) Classics() []domain.Match {
	out := []domain.Match{}
	for _, c := range classicEntries {
		out = append(out, cloneClassic(c.match))
	}
	return out
}

// Match finds a fixture (computed for now) or a classic by id.
func (w *World) Match(id string, now time.Time) (domain.Match, bool) {
	if f, ok := fixtureByID[id]; ok {
		return w.fixtureMatch(f, now), true
	}
	if c, ok := classicByID[id]; ok {
		return cloneClassic(c.match), true
	}
	if i, ok := archiveByID[id]; ok {
		return cloneArchived(archiveList[i]), true
	}
	return domain.Match{}, false
}

// Friends returns the 8 simulated friends a guest can follow.
func (w *World) Friends() []domain.User { return toUsers(friendRows) }

// PublicFans returns non-friend fans who write the public "popular reviews".
func (w *World) PublicFans() []domain.User { return toUsers(publicRows) }

// User finds a simulated friend or public fan by id.
func (w *World) User(id string) (domain.User, bool) {
	r, ok := personByID[id]
	if !ok {
		return domain.User{}, false
	}
	return toUsers([]personRow{r})[0], true
}

// Logs returns friends' and public fans' logs that exist by `now` (completed fixtures + classics), newest first.
func (w *World) Logs(now time.Time) []domain.MatchLog { return w.logs(now) }

// Predictions returns friends' predictions made by `now` (CreatedAt is always before the match start).
func (w *World) Predictions(now time.Time) []domain.Prediction { return w.predictions(now) }

// BaseLikes is the simulated like count on a simulated log (real users' likes are added on top by the API).
func (w *World) BaseLikes(logID string) int { return baseLikes(logID) }

// BaseThereToo is the simulated "I was there too" count on a simulated log.
func (w *World) BaseThereToo(logID string) int { return w.baseThereToo(logID) }

// Replies returns simulated replies on a simulated log that exist by `now`, oldest first.
func (w *World) Replies(logID string, now time.Time) []domain.Reply { return w.replies(logID, now) }

// Crowd returns the anonymous crowd for a match (zero-value maps/slices are non-nil).
// Ratings, tallies and watch counts are filled for every fixture (Picks are needed before the match), so callers
// should only show the rating side once the match is completed.
func (w *World) Crowd(matchID string) Crowd { return w.crowd(matchID) }

// Calls: the simulated world has no predictions during a match (those come with real matches, internal/feed).
func (w *World) Calls(matchID string) []domain.LiveCall { return []domain.LiveCall{} }

func (w *World) Call(id string) (domain.LiveCall, bool) { return domain.LiveCall{}, false }

func (w *World) PostCall(domain.CallSpec, time.Time) (domain.LiveCall, error) {
	return domain.LiveCall{}, domain.ErrCallsOff
}

// FriendsWatching returns ids of simulated friends watching a live (or about-to-start) match.
func (w *World) FriendsWatching(matchID string, now time.Time) []string {
	return w.friendsWatching(matchID, now)
}

// DemoHistory is what a brand-new guest starts with: logs of Vault classics and of fixtures that finished 2+
// days before `now`, settled predictions on such fixtures, and the friend ids they follow (5 of 8).
func (w *World) DemoHistory(userID string, now time.Time) (logs []domain.MatchLog, preds []domain.Prediction, follows []string) {
	return w.demoHistory(userID, now)
}
