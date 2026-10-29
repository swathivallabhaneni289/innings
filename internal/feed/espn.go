package feed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ESPN's public cricket feed: the JSON behind espn.in's scores pages (the same data as ESPNcricinfo). It needs no
// key or account. It isn't an official API, so it can change without notice; everything here reads it defensively
// and a failed call just waits for the next poll.
const defaultBase = "https://site.web.api.espn.com/apis/"

const userAgent = "Innings/0.1 (personal cricket diary)"

type client struct {
	base string
	http *http.Client
}

func newClient() *client {
	return &client{base: defaultBase, http: &http.Client{Timeout: 20 * time.Second}}
}

func (c *client) get(ctx context.Context, path string, params url.Values, out any) error {
	params.Set("lang", "en")
	params.Set("region", "in")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path+"?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("espn %s: %w", path, err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		if ue, ok := errors.AsType[*url.Error](err); ok {
			err = ue.Err
		}
		return fmt.Errorf("espn %s: %w", path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("espn %s: HTTP %d", path, res.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 16<<20)).Decode(out); err != nil {
		return fmt.Errorf("espn %s: %w", path, err)
	}
	return nil
}

// ---- the scoreboard: every current, recent and next-day match in one call, or one day's ----

type esScoreboard struct {
	Sports []struct {
		Leagues []struct {
			ID     string    `json:"id"`
			Name   string    `json:"name"` // "Pakistan ODI Tri-Series", "India tour of New Zealand 2026/27"
			Events []esEvent `json:"events"`
		} `json:"leagues"`
	} `json:"sports"`
}

type esEvent struct {
	ID       string `json:"id"`
	Date     string `json:"date"`  // RFC 3339, the scheduled first ball
	Title    string `json:"title"` // "6th Match (D/N)"
	Location string `json:"location"`
	Status   string `json:"status"` // "pre" | "in" | "post"
	Class    struct {
		Card string `json:"generalClassCard"` // "T20I", "ODI", "Test", "Women's T20", "Twenty20", "First-class", …
	} `json:"class"`
	FullStatus struct {
		LongSummary   string `json:"longSummary"` // "Sri Lanka require another 257 runs with 8 wickets and 28.4 overs remaining"
		Summary       string `json:"summary"`
		BattingTeamID int    `json:"battingTeamId"`
	} `json:"fullStatus"`
	Notes []struct {
		Type string `json:"type"` // "toss" says who chose what
		Text string `json:"text"`
	} `json:"notes"`
	Competitors []esCompetitor `json:"competitors"`

	LeagueID string `json:"leagueId"` // filled from the enclosing league (a match's summary is fetched under it)
	League   string `json:"league"`
}

// flexString reads a JSON string or number as text: ESPN sends ids and runs either way.
type flexString string

func (s *flexString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = flexString(str)
		return nil
	}
	*s = flexString(strings.TrimSpace(string(data)))
	return nil
}

// flexBool reads true, false, "true" or "false": the scoreboard sends a winner as a JSON bool, a summary as a string.
type flexBool bool

func (b *flexBool) UnmarshalJSON(data []byte) error {
	switch strings.Trim(string(data), `"`) {
	case "true":
		*b = true
	case "false", "null", "":
		*b = false
	default:
		return fmt.Errorf("not a bool: %s", data)
	}
	return nil
}

type esCompetitor struct {
	ID           string   `json:"id"`
	DisplayName  string   `json:"displayName"`
	Abbreviation string   `json:"abbreviation"`
	Color        string   `json:"color"`
	Winner       flexBool `json:"winner"`
	Score        string   `json:"score"` // "386/5", "123/2 (21/50 ov, target 387)", "329 & 63/2 (19.4 ov)"
}

// scoreboard returns the current, recent and next-day matches, or with day set ("20261102") that day's matches.
func (c *client) scoreboard(ctx context.Context, day string) ([]esEvent, error) {
	params := url.Values{"sport": {"cricket"}}
	if day != "" {
		params.Set("dates", day)
	}
	var sb esScoreboard
	if err := c.get(ctx, "v2/scoreboard/header", params, &sb); err != nil {
		return nil, err
	}
	var out []esEvent
	for _, sp := range sb.Sports {
		for _, lg := range sp.Leagues {
			for _, e := range lg.Events {
				e.LeagueID, e.League = lg.ID, lg.Name
				out = append(out, e)
			}
		}
	}
	return out, nil
}

// ---- one match's summary: squads, XIs, innings, leaders, player of the match ----

type esAthlete struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Position    struct {
		Name string `json:"name"` // "Top-order batter", "Wicketkeeper batter", "Bowling allrounder", "Bowler", …
	} `json:"position"`
	Keeper flexBool `json:"keeper"`
}

type esTeamRef struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type esSummary struct {
	// The batting and bowling cards for the innings in progress ("Batting" / "Bowling" / "Partnerships").
	Matchcards []struct {
		Headline      string     `json:"headline"`
		InningsNumber flexString `json:"inningsNumber"`
		PlayerDetails []struct {
			PlayerID   flexString `json:"playerID"`
			PlayerName string     `json:"playerName"`
			Runs       flexString `json:"runs"`      // "" for a batter who hasn't batted
			Dismissal  string     `json:"dismissal"` // "not out" at the crease, "" not batted yet
		} `json:"playerDetails"`
	} `json:"matchcards"`
	Squads []struct {
		Team     esTeamRef   `json:"team"`
		Athletes []esAthlete `json:"athletes"`
	} `json:"squads"`
	Rosters []struct {
		Team   esTeamRef `json:"team"`
		Roster []struct {
			Starter flexBool  `json:"starter"`
			Athlete esAthlete `json:"athlete"`
		} `json:"roster"`
	} `json:"rosters"`
	Leaders []struct {
		Team       esTeamRef `json:"team"`
		Linescores []struct {
			IsBatting flexBool `json:"isBatting"`
			Leaders   []struct {
				Name    string `json:"name"` // "runs" | "wickets"
				Leaders []struct {
					Value   string    `json:"value"`
					Athlete esAthlete `json:"athlete"`
				} `json:"leaders"`
			} `json:"leaders"`
		} `json:"linescores"`
	} `json:"leaders"`
	GameInfo struct {
		Venue struct {
			FullName string `json:"fullName"`
			Address  struct {
				City    string `json:"city"`
				Country string `json:"country"`
			} `json:"address"`
		} `json:"venue"`
	} `json:"gameInfo"`
	Header struct {
		Competitions []struct {
			Status struct {
				Type struct {
					State string `json:"state"`
				} `json:"type"`
				FeaturedAthletes []struct {
					Name    string    `json:"name"` // "playerOfTheMatch"
					Athlete esAthlete `json:"athlete"`
				} `json:"featuredAthletes"`
			} `json:"status"`
			Competitors []struct {
				ID         string   `json:"id"`
				Winner     flexBool `json:"winner"`
				Linescores []struct {
					Period    int      `json:"period"` // the innings' place in the match: 1, 2, 3, 4
					Runs      int      `json:"runs"`
					Wickets   int      `json:"wickets"`
					Overs     float64  `json:"overs"`
					IsBatting flexBool `json:"isBatting"`
				} `json:"linescores"`
			} `json:"competitors"`
		} `json:"competitions"`
	} `json:"header"`
}

func (c *client) summary(ctx context.Context, leagueID, eventID string) (esSummary, error) {
	var s esSummary
	err := c.get(ctx, "site/v2/sports/cricket/"+url.PathEscape(leagueID)+"/summary", url.Values{"event": {eventID}}, &s)
	return s, err
}
