package sim

import (
	"sort"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// classicEntry keeps the seed next to its converted match (Hype drives simulated ratings).
type classicEntry struct {
	seed  ClassicSeed
	match domain.Match
}

var classicEntries, classicByID = loadClassics()

func loadClassics() ([]classicEntry, map[string]*classicEntry) {
	out := []classicEntry{}
	seeds, err := LoadClassicSeeds()
	if err != nil {
		return out, map[string]*classicEntry{}
	}
	for _, s := range seeds {
		if s.ID == "" {
			continue
		}
		if s.Innings == nil {
			s.Innings = []domain.InningsScore{}
		}
		if s.Collections == nil {
			s.Collections = []domain.CollectionID{}
		}
		s.Hype = clampI(s.Hype, 1, 5)
		var winner *string
		if s.WinnerID != nil {
			winner = domain.Ptr(*s.WinnerID)
		}
		m := domain.Match{
			ID: s.ID, Kind: "classic", CompetitionID: s.CompetitionID, Format: s.Format, Gender: s.Gender, Stage: s.Stage,
			TeamIDs: s.TeamIDs, Venue: s.Venue, StartTime: s.Date + "T12:00:00Z", State: domain.Completed,
			Result: &domain.MatchResult{
				WinnerID: winner, Text: s.ResultText, Innings: append([]domain.InningsScore{}, s.Innings...),
				TopRunScorerIDs: []string{}, TopWicketTakerIDs: []string{}, OfficialPotmID: nil,
			},
			Classic: &domain.ClassicInfo{
				Date: s.Date, CompetitionName: s.CompetitionName, Headline: s.Headline, Scoreline: s.Scoreline,
				PotmName: s.PotmName, Collections: append([]domain.CollectionID{}, s.Collections...), Era: s.Era,
			},
		}
		out = append(out, classicEntry{seed: s, match: m})
	}
	sort.SliceStable(out, func(a, b int) bool {
		if out[a].seed.Date != out[b].seed.Date {
			return out[a].seed.Date < out[b].seed.Date
		}
		return out[a].seed.ID < out[b].seed.ID
	})
	byID := map[string]*classicEntry{}
	for k := range out {
		byID[out[k].seed.ID] = &out[k]
	}
	return out, byID
}

// cloneClassic returns a deep-enough copy so callers can't mutate shared state.
func cloneClassic(m domain.Match) domain.Match {
	c := m
	r := *m.Result
	r.Innings = append([]domain.InningsScore{}, r.Innings...)
	r.TopRunScorerIDs, r.TopWicketTakerIDs = []string{}, []string{}
	if r.WinnerID != nil {
		r.WinnerID = domain.Ptr(*r.WinnerID)
	}
	c.Result = &r
	ci := *m.Classic
	ci.Collections = append([]domain.CollectionID{}, ci.Collections...)
	c.Classic = &ci
	return c
}

func classicWinnerLoser(s ClassicSeed) (string, string) {
	if s.WinnerID == nil {
		return "", ""
	}
	w := *s.WinnerID
	l := s.TeamIDs[0]
	if l == w {
		l = s.TeamIDs[1]
	}
	return w, l
}
