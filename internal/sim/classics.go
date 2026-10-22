package sim

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// classics.json is the hand-seeded, fact-checked Vault (owned by the Vault seed task; read-only for sim code).
//
//go:embed classics.json
var classicsJSON []byte

// ClassicSeed is one entry of classics.json.
type ClassicSeed struct {
	ID              string                `json:"id"`              // "wc-2011-final"
	Date            string                `json:"date"`            // "2011-04-02" (first day for Tests)
	CompetitionID   domain.CompetitionID  `json:"competitionId"`   // world-cup for any World Cup / T20 World Cup / WWC; otherwise international (or a franchise league)
	CompetitionName string                `json:"competitionName"` // "ICC Cricket World Cup 2011"
	Stage           string                `json:"stage"`           // "Final", "2nd Test", "Semi-final", "Group A"
	Format          domain.Format         `json:"format"`
	Gender          domain.Gender         `json:"gender"`
	TeamIDs         [2]string             `json:"teamIds"`    // canonical team ids (see sim team list)
	Venue           domain.Venue          `json:"venue"`      // id = kebab-case of the ground name
	WinnerID        *string               `json:"winnerId"`   // null for a tie or a draw
	ResultText      string                `json:"resultText"` // "India won by 6 wickets"
	Innings         []domain.InningsScore `json:"innings"`    // in batting order; Tests list all innings
	Scoreline       string                `json:"scoreline"`  // "SL 274/6 (50) · IND 277/4 (48.2)"
	PotmName        string                `json:"potmName,omitempty"`
	Headline        string                `json:"headline"` // one-line hook, <= 60 chars
	Collections     []domain.CollectionID `json:"collections"`
	Era             string                `json:"era"`  // "2010s"
	Hype            int                   `json:"hype"` // 1–5, how legendary; shapes simulated fan ratings
}

// LoadClassicSeeds parses the embedded Vault file.
func LoadClassicSeeds() ([]ClassicSeed, error) {
	var seeds []ClassicSeed
	if err := json.Unmarshal(classicsJSON, &seeds); err != nil {
		return nil, fmt.Errorf("classics.json: %w", err)
	}
	return seeds, nil
}
