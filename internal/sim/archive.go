package sim

import (
	_ "embed"
	"encoding/json"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// archive.json is Innings' own record of past tournament matches, built by scripts/archive.py with Wikipedia's
// tournament pages as the reference. Search and logging reach these matches; the Matches tab and Home don't list
// them. The teams it adds (sides that no longer play, smaller nations) join the team table.
//
//go:embed archive.json
var archiveJSON []byte

// archiveSeed is one match of archive.json.
type archiveSeed struct {
	ID              string                `json:"id"` // "ar-1384403": the ESPNcricinfo scorecard's number
	Date            string                `json:"date"`
	CompetitionID   domain.CompetitionID  `json:"competitionId"`
	CompetitionName string                `json:"competitionName"` // "ICC Cricket World Cup 2023"
	Stage           string                `json:"stage"`           // "Group stage", "Asia Cup · Final"
	Format          domain.Format         `json:"format"`
	Gender          domain.Gender         `json:"gender"`
	TeamIDs         [2]string             `json:"teamIds"` // the side that batted first first
	Venue           domain.Venue          `json:"venue"`
	WinnerID        *string               `json:"winnerId"` // null for a tie, no result or abandonment
	ResultText      string                `json:"resultText"`
	Innings         []domain.InningsScore `json:"innings"`
}

var archiveTeams, archiveList, archiveByID = loadArchive()

func loadArchive() ([]domain.Team, []domain.Match, map[string]int) {
	var f struct {
		Teams   []domain.Team `json:"teams"`
		Matches []archiveSeed `json:"matches"`
	}
	if err := json.Unmarshal(archiveJSON, &f); err != nil {
		return nil, nil, map[string]int{}
	}
	list, byID := make([]domain.Match, 0, len(f.Matches)), map[string]int{}
	for _, s := range f.Matches {
		var winner *string
		if s.WinnerID != nil {
			winner = domain.Ptr(*s.WinnerID)
		}
		byID[s.ID] = len(list)
		list = append(list, domain.Match{
			ID: s.ID, Kind: "fixture", CompetitionID: s.CompetitionID, Format: s.Format, Gender: s.Gender, Stage: s.Stage,
			TeamIDs: s.TeamIDs, Venue: s.Venue, StartTime: s.Date + "T12:00:00Z", State: domain.Completed, Archived: true,
			Result: &domain.MatchResult{
				WinnerID: winner, Text: s.ResultText, Innings: append([]domain.InningsScore{}, s.Innings...),
				TopRunScorerIDs: []string{}, TopWicketTakerIDs: []string{},
			},
		})
	}
	return f.Teams, list, byID
}

// cloneArchived copies what a caller could change.
func cloneArchived(m domain.Match) domain.Match {
	r := *m.Result
	r.Innings = append([]domain.InningsScore{}, r.Innings...)
	r.TopRunScorerIDs, r.TopWicketTakerIDs = []string{}, []string{}
	if r.WinnerID != nil {
		r.WinnerID = domain.Ptr(*r.WinnerID)
	}
	m.Result = &r
	return m
}

// Archive returns the past tournament matches of archive.json, oldest first.
func (w *World) Archive() []domain.Match {
	out := make([]domain.Match, len(archiveList))
	for i, m := range archiveList {
		out[i] = cloneArchived(m)
	}
	return out
}

// Archived reports whether id is an archive match.
func (w *World) Archived(id string) bool {
	_, ok := archiveByID[id]
	return ok
}
