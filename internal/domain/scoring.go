package domain

import (
	"math"
	"slices"
	"time"
)

// Bragging points (docs/PLAN.md → "Build decisions").
const (
	PointsWinner     = 10
	PointsTopScorer  = 15
	PointsTopWicket  = 15
	PointsUpsetBonus = 10
	UpsetThreshold   = 0.30 // the winning pick had < 30% of fan picks at lock time
	StadiumRegularAt = 3    // stadium logs needed for the Stadium Regular badge
)

// Badges, in display order.
var (
	BadgeFirstCall      = BadgeDef{ID: "first-call", Title: "First Call", Description: "Called your first winner."}
	BadgeCalledUpset    = BadgeDef{ID: "called-the-upset", Title: "Called the Upset", Description: "Backed a winner that fewer than 30% of fans picked."}
	BadgeFiveStreak     = BadgeDef{ID: "five-match-streak", Title: "5-Match Streak", Description: "Called five winners in a row."}
	BadgeSharpshooter   = BadgeDef{ID: "sharpshooter", Title: "Sharpshooter", Description: "Called the winner, top scorer and top wicket-taker of one match."}
	BadgeStadiumRegular = BadgeDef{ID: "stadium-regular", Title: "Stadium Regular", Description: "Logged three matches from the stands."}
)

// IsLocked reports whether predictions for m are closed: from the first ball on (or once it isn't upcoming).
func IsLocked(m Match, now time.Time) bool {
	if m.State != Upcoming {
		return true
	}
	start, err := time.Parse(time.RFC3339, m.StartTime)
	if err != nil {
		return true
	}
	return !now.Before(start)
}

// StreakMultiplier: 1–2 → ×1, 3–4 → ×1.5, 5+ → ×2.
func StreakMultiplier(streak int) float64 {
	switch {
	case streak >= 5:
		return 2
	case streak >= 3:
		return 1.5
	default:
		return 1
	}
}

// ScoreOne settles one prediction given the streak before this match and the winner's share of fan picks at lock
// time (ok=false when nobody predicted, which means no upset). Unsettled matches return Settled=false, 0 points.
func ScoreOne(p Prediction, m Match, streakBefore int, winnerShare float64, shareKnown bool) PredictionOutcome {
	o := PredictionOutcome{MatchID: p.MatchID, UserID: p.UserID, Multiplier: 1, Streak: streakBefore}
	if m.State != Completed || m.Result == nil {
		return o
	}
	r := m.Result
	o.Settled = true
	o.TopScorerCorrect = slices.Contains(r.TopRunScorerIDs, p.TopRunScorerID)
	o.TopWicketCorrect = slices.Contains(r.TopWicketTakerIDs, p.TopWicketTakerID)
	if r.WinnerID != nil {
		o.WinnerCorrect = *r.WinnerID == p.WinnerID
		if o.WinnerCorrect {
			o.Streak = streakBefore + 1
			o.Upset = shareKnown && winnerShare < UpsetThreshold
		} else {
			o.Streak = 0
		}
	}
	if o.WinnerCorrect {
		o.Base += PointsWinner
		o.Multiplier = StreakMultiplier(o.Streak)
	}
	if o.Upset {
		o.Base += PointsUpsetBonus
	}
	if o.TopScorerCorrect {
		o.Base += PointsTopScorer
	}
	if o.TopWicketCorrect {
		o.Base += PointsTopWicket
	}
	o.Points = int(math.Round(float64(o.Base) * o.Multiplier))
	return o
}

// SettleHistory scores one user's whole prediction history in match start order (streaks carry across matches).
// lookup finds a match; winnerShare returns the winner's share of all fan picks at lock time (ok=false if none).
// Predictions whose match is unknown are skipped. The result is in match start order.
func SettleHistory(preds []Prediction, lookup func(id string) (Match, bool), winnerShare func(Match) (float64, bool)) []PredictionOutcome {
	type pm struct {
		p Prediction
		m Match
	}
	items := make([]pm, 0, len(preds))
	for _, p := range preds {
		if m, ok := lookup(p.MatchID); ok {
			items = append(items, pm{p, m})
		}
	}
	slices.SortStableFunc(items, func(a, b pm) int {
		ta, _ := time.Parse(time.RFC3339, a.m.StartTime)
		tb, _ := time.Parse(time.RFC3339, b.m.StartTime)
		return ta.Compare(tb)
	})
	out := make([]PredictionOutcome, 0, len(items))
	streak := 0
	for _, it := range items {
		share, known := 0.0, false
		if it.m.State == Completed && it.m.Result != nil && it.m.Result.WinnerID != nil && winnerShare != nil {
			share, known = winnerShare(it.m)
		}
		o := ScoreOne(it.p, it.m, streak, share, known)
		if o.Settled {
			streak = o.Streak
		}
		out = append(out, o)
	}
	return out
}

// HistoryStats summarises settled outcomes (in match order) for a profile.
type HistoryStats struct {
	Points    int
	Streak    int      // current correct-winner streak
	Accuracy  *float64 // correct winners / settled matches with a winner; nil when none
	MaxStreak int
	Upsets    int
	Perfect   int // matches with all three picks right
	Correct   int // correct winners
}

// ties holds the ids of settled matches without a winner (they don't count towards accuracy); may be nil.
func SummariseHistory(outcomes []PredictionOutcome, ties map[string]bool) HistoryStats {
	var s HistoryStats
	decided := 0
	for _, o := range outcomes {
		if !o.Settled {
			continue
		}
		s.Points += o.Points
		s.Streak = o.Streak
		s.MaxStreak = max(s.MaxStreak, o.Streak)
		if o.Upset {
			s.Upsets++
		}
		if o.WinnerCorrect {
			s.Correct++
			decided++
		} else if !ties[o.MatchID] {
			decided++
		}
		if o.WinnerCorrect && o.TopScorerCorrect && o.TopWicketCorrect {
			s.Perfect++
		}
	}
	if decided > 0 {
		a := float64(s.Correct) / float64(decided)
		s.Accuracy = &a
	}
	return s
}

// Badges earned from prediction history and the number of logs watched from the stadium. Never nil.
func Badges(s HistoryStats, stadiumLogs int) []BadgeDef {
	out := []BadgeDef{}
	if s.Correct > 0 {
		out = append(out, BadgeFirstCall)
	}
	if s.Upsets > 0 {
		out = append(out, BadgeCalledUpset)
	}
	if s.MaxStreak >= 5 {
		out = append(out, BadgeFiveStreak)
	}
	if s.Perfect > 0 {
		out = append(out, BadgeSharpshooter)
	}
	if stadiumLogs >= StadiumRegularAt {
		out = append(out, BadgeStadiumRegular)
	}
	return out
}
