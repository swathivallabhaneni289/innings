package sim

import (
	"math"
	"slices"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

func emptyCrowd() Crowd {
	return Crowd{RatingBuckets: make([]int, 10), PlayerTallies: map[string]Tally{}, Watch: map[domain.WatchPlace]int{}, Picks: map[string]int{}}
}

// buckets spreads n ratings over the 10 half-ball steps around mean mu.
func buckets(n int, mu, sd float64) []int {
	out := make([]int, 10)
	weights := make([]float64, 10)
	sum := 0.0
	for k := range weights {
		v := float64(k+1) * 0.5
		weights[k] = math.Exp(-(v - mu) * (v - mu) / (2 * sd * sd))
		sum += weights[k]
	}
	for k := range out {
		out[k] = int(math.Round(float64(n) * weights[k] / sum))
	}
	return out
}

func watchCounts(r *rng, n int, classic bool) map[domain.WatchPlace]int {
	shares := map[domain.WatchPlace]float64{
		domain.WatchHome: r.frange(0.46, 0.55), domain.WatchFriends: r.frange(0.15, 0.2), domain.WatchStadium: r.frange(0.06, 0.1),
		domain.WatchHostel: r.frange(0.06, 0.09), domain.WatchCafe: r.frange(0.04, 0.07), domain.WatchOffice: r.frange(0.03, 0.05),
		domain.WatchOnTheGo: r.frange(0.02, 0.04),
	}
	if classic {
		shares[domain.WatchStadium] = r.frange(0.02, 0.04)
	}
	out := map[domain.WatchPlace]int{}
	for _, p := range domain.WatchPlaces {
		out[p] = int(math.Round(float64(n) * shares[p]))
	}
	return out
}

func (w *World) crowd(matchID string) Crowd {
	if f, ok := fixtureByID[matchID]; ok {
		return fixtureCrowd(f)
	}
	if c, ok := classicByID[matchID]; ok {
		r := newRNG("crowd", matchID)
		cr := emptyCrowd()
		n := clampI(600+c.seed.Hype*450+r.between(-150, 150), 150, 3000)
		cr.RatingBuckets = buckets(n, 1.6+0.65*float64(c.seed.Hype), 0.7)
		cr.Watch = watchCounts(r, n, true)
		return cr
	}
	return emptyCrowd()
}

// fixtureCrowd is computed for every fixture; the API should only show ratings/tallies once it is completed.
func fixtureCrowd(f *fixture) Crowd {
	r := newRNG("crowd", f.id)
	cr := emptyCrowd()
	n := clampI(150+int(float64(r.between(500, 2850))*f.pop), 150, 3000)
	mu := 1.5 + 3.2*f.close
	switch f.vibe {
	case "century":
		mu += 0.6
	case "bigchase":
		mu += 0.7
	case "rain":
		mu = 2
	}
	cr.RatingBuckets = buckets(n, clampF(mu, 0.8, 4.9), 0.75)
	cr.Watch = watchCounts(r, n, false)

	for _, team := range f.teams {
		for _, id := range f.xi[team] {
			var avg, share float64
			top := slices.Contains(f.topBatIDs, id) || slices.Contains(f.topBowlID, id)
			switch {
			case id == f.fav:
				avg, share = r.frange(8.9, 9.3), r.frange(0.4, 0.55)
			case id == f.potmID:
				avg, share = r.frange(8.3, 8.6), r.frange(0.35, 0.5)
			case top:
				avg, share = r.frange(7.4, 8.1), r.frange(0.3, 0.45)
			case f.winner >= 0 && team == f.winnerID():
				avg, share = r.frange(5.6, 7.2), r.frange(0.08, 0.3)
			default:
				avg, share = r.frange(4.6, 6.6), r.frange(0.08, 0.3)
			}
			cnt := max(3, int(float64(n)*share))
			cr.PlayerTallies[id] = Tally{Sum: int(math.Round(avg * float64(cnt))), Count: cnt}
		}
	}

	picks := int(float64(n) * r.frange(0.7, 1.2))
	var share0 float64 // share for teams[0]
	switch {
	case f.winner < 0:
		share0 = r.frange(0.42, 0.58)
	case f.upset:
		ws := r.frange(0.2, 0.24)
		share0 = ws
		if f.winner == 1 {
			share0 = 1 - ws
		}
	default:
		ws := r.frange(0.42, 0.7)
		share0 = ws
		if f.winner == 1 {
			share0 = 1 - ws
		}
	}
	a := int(math.Round(float64(picks) * share0))
	cr.Picks[f.teams[0]], cr.Picks[f.teams[1]] = a, picks-a
	return cr
}

func (w *World) friendsWatching(matchID string, now time.Time) []string {
	out := []string{}
	f, ok := fixtureByID[matchID]
	if !ok {
		return out
	}
	var n int
	switch {
	case w.stateAt(f, now) == domain.Live:
		n = 2 + newRNG("watchn", f.id).intn(3)
	case now.Before(w.start(f)) && w.start(f).Sub(now) <= 2*time.Hour:
		n = 1 + newRNG("watchn", f.id).intn(2)
	default:
		return out
	}
	// fans of the playing teams first, then a deterministic shuffle of the rest
	r := newRNG("watching", f.id)
	rest := []string{}
	for _, fr := range shuffled(r, friendRows) {
		if slices.Contains(f.teams[:], fr.team) {
			out = append(out, fr.id)
		} else {
			rest = append(rest, fr.id)
		}
	}
	out = append(out, rest...)
	return out[:n]
}
