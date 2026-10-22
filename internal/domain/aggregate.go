package domain

import (
	"cmp"
	"math"
	"slices"
)

// MinPotmRatings is how many player ratings a player needs to be the Fans' Player of the Match.
const MinPotmRatings = 3

// Tally is a running sum of 1–10 player ratings (mirrors sim.Tally; domain can't import sim).
type Tally struct {
	Sum   int
	Count int
}

// BucketIndex maps a 0.5–5 rating to its half-ball bucket 0..9 (-1 if out of range).
func BucketIndex(rating float64) int {
	i := int(math.Round(rating*2)) - 1
	if i < 0 || i > 9 {
		return -1
	}
	return i
}

// ValidRating: 0.5–5 in 0.5 steps.
func ValidRating(r float64) bool {
	return r >= 0.5 && r <= 5 && r*2 == math.Trunc(r*2)
}

// AddRating counts rating into buckets (len 10).
func AddRating(buckets []int, rating float64) {
	if i := BucketIndex(rating); i >= 0 && i < len(buckets) {
		buckets[i]++
	}
}

// BucketAverage returns the mean rating and count of 10 half-ball buckets (0, 0 when empty). Rounded to 2 dp.
func BucketAverage(buckets []int) (avg float64, count int) {
	sum := 0.0
	for i, n := range buckets {
		sum += float64(n) * float64(i+1) / 2
		count += n
	}
	if count == 0 {
		return 0, 0
	}
	return Round2(sum / float64(count)), count
}

// PlayerAverages turns tallies into averages, best first (ties: more ratings first, then id). Never nil.
func PlayerAverages(tallies map[string]Tally) []PlayerAverage {
	out := make([]PlayerAverage, 0, len(tallies))
	for id, t := range tallies {
		if t.Count <= 0 {
			continue
		}
		out = append(out, PlayerAverage{PlayerID: id, Avg: Round2(float64(t.Sum) / float64(t.Count)), Count: t.Count})
	}
	slices.SortFunc(out, func(a, b PlayerAverage) int {
		if c := cmp.Compare(b.Avg, a.Avg); c != 0 {
			return c
		}
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(a.PlayerID, b.PlayerID)
	})
	return out
}

// FansPotm is the best-averaged player with at least MinPotmRatings ratings (players must be sorted best first).
func FansPotm(players []PlayerAverage) *string {
	for _, p := range players {
		if p.Count >= MinPotmRatings {
			return Ptr(p.PlayerID)
		}
	}
	return nil
}

// WatchShares converts counts to shares, sorted desc (ties in WatchPlaces order). Zero counts are dropped. Never nil.
func WatchShares(counts map[WatchPlace]int) []WatchShare {
	total := 0
	for _, n := range counts {
		total += n
	}
	out := []WatchShare{}
	if total == 0 {
		return out
	}
	for _, p := range WatchPlaces {
		if n := counts[p]; n > 0 {
			out = append(out, WatchShare{Place: p, Share: float64(n) / float64(total)})
		}
	}
	slices.SortStableFunc(out, func(a, b WatchShare) int { return cmp.Compare(b.Share, a.Share) })
	return out
}

// WatchCounts lists non-zero counts, sorted desc (ties in WatchPlaces order). Never nil.
func WatchCounts(counts map[WatchPlace]int) []WatchCount {
	out := []WatchCount{}
	for _, p := range WatchPlaces {
		if n := counts[p]; n > 0 {
			out = append(out, WatchCount{Place: p, Count: n})
		}
	}
	slices.SortStableFunc(out, func(a, b WatchCount) int { return cmp.Compare(b.Count, a.Count) })
	return out
}

// Shares splits picks between the two teams; both ids are always present (0 when nobody picked).
func Shares(teamIDs [2]string, picks map[string]int) (map[string]float64, int) {
	total := picks[teamIDs[0]] + picks[teamIDs[1]]
	out := map[string]float64{teamIDs[0]: 0, teamIDs[1]: 0}
	if total > 0 {
		out[teamIDs[0]] = float64(picks[teamIDs[0]]) / float64(total)
		out[teamIDs[1]] = float64(picks[teamIDs[1]]) / float64(total)
	}
	return out, total
}

func Round2(f float64) float64 { return math.Round(f*100) / 100 }
