package feed

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// Real matches get "es-" ids (players "esp-"), so they can never collide with the simulation's or the Vault's.
const (
	matchPrefix  = "es-"
	playerPrefix = "esp-"
	newTeamInk   = "#3A4250" // kit colour for a team the app has no colours for and the feed gives none
)

// Countries whose senior sides count as internationals. A, U19 and domestic sides don't match these names exactly,
// so they're left out; the women's sides are these names + " Women".
var nations = []string{
	"India", "Australia", "England", "South Africa", "New Zealand", "Pakistan", "Sri Lanka", "West Indies",
	"Bangladesh", "Afghanistan", "Zimbabwe", "Ireland", "Netherlands", "Scotland", "United States", "United States of America",
	"USA", "Kenya", "Namibia", "United Arab Emirates", "UAE", "Nepal", "Oman", "Canada", "Hong Kong", "Papua New Guinea",
	"Uganda", "Italy", "Jersey", "Thailand",
}

// Leagues the app knows, found in the feed's league (series) name.
var leagueNames = []struct {
	re   *regexp.Regexp
	comp domain.CompetitionID
}{
	{regexp.MustCompile(`(?i)women'?s premier league|\bWPL\b`), "wpl"},
	{regexp.MustCompile(`(?i)indian premier league|\bIPL\b`), "ipl"},
	{regexp.MustCompile(`(?i)big bash|\bW?BBL\b`), "bbl"},
	{regexp.MustCompile(`(?i)\bSA20\b`), "sa20"},
	{regexp.MustCompile(`(?i)\bhundred\b`), "the-hundred"},
}

var (
	worldCup    = regexp.MustCompile(`(?i)world cup`)
	notWorldCup = regexp.MustCompile(`(?i)league|qualifier|challenge|sub.?regional`) // "ICC Men's Cricket World Cup League 2"
	slugBad     = regexp.MustCompile(`[^a-z0-9]+`)
	dayNight    = regexp.MustCompile(`\s*\((D/N|N)\)\s*$`)
	// One innings in a scoreboard line: "386/5", "429", "250/7d", "123/2 (21/50 ov, target 387)", "63/2 (19.4 ov)".
	inningsText = regexp.MustCompile(`^(\d+)(?:/(\d+))?(d)?(?:\s*\(([\d.]+)(?:/\d+)?\s*ov)?`)
	tossText    = regexp.MustCompile(`(?i)^(.+?)\s*,\s*elected to (bat|field)`)
)

// aliases are other names the feed may use for sides in the simulation's team table.
var aliases = map[string]string{
	"royal challengers bangalore": "rcb",
	"kings xi punjab":             "pbks",
	"united states of america":    "usa",
	"durban super giants":         "dsg",
}

// catalog is everything mapping needs to know about teams: the simulation's (with their colours) plus any the feed
// has added, both looked up by lower-case name.
type catalog struct {
	teams  map[string]domain.Team // id → team
	byName map[string]string      // lower-case name → id
}

func newCatalog(known map[string]domain.Team) *catalog {
	c := &catalog{teams: map[string]domain.Team{}, byName: map[string]string{}}
	for _, t := range known {
		c.add(t)
	}
	for name, id := range aliases {
		if _, ok := c.teams[id]; ok {
			c.byName[name] = id
		}
	}
	return c
}

func (c *catalog) add(t domain.Team) {
	c.teams[t.ID] = t
	c.byName[strings.ToLower(t.Name)] = t.ID
}

// freeID is id, or id with a number on the end if another team already has it.
func (c *catalog) freeID(id string) string {
	out := id
	for n := 2; ; n++ {
		if _, taken := c.teams[out]; !taken {
			return out
		}
		out = id + "-" + strconv.Itoa(n)
	}
}

func slug(s string) string {
	return strings.Trim(slugBad.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func isNation(name string) bool {
	return slices.ContainsFunc(nations, func(n string) bool {
		return strings.EqualFold(name, n) || strings.EqualFold(name, n+" Women")
	})
}

func womenName(name string) bool { return strings.HasSuffix(strings.ToLower(name), " women") }

// team finds or makes the team for a feed competitor. New teams are returned with added=true so the caller can keep
// them; nothing is added for a match the app doesn't show.
func (c *catalog) team(ec esCompetitor) (t domain.Team, added bool) {
	name := ec.DisplayName
	if id, ok := c.byName[strings.ToLower(name)]; ok {
		return c.teams[id], false
	}
	short := strings.ToUpper(ec.Abbreviation)
	if short == "" {
		short = strings.ToUpper(name[:min(3, len(name))])
	}
	colour := newTeamInk
	if len(ec.Color) == 7 && ec.Color[0] == '#' {
		colour = strings.ToUpper(ec.Color)
	}
	t = domain.Team{ID: slug(name), Name: name, Short: short, Colour: colour, Gender: domain.Men, Kind: "franchise"}
	if womenName(name) {
		t.Gender = domain.Women
		// A women's side of a team the app knows wears its colours: "Sydney Sixers Women" → sixers-w, SIX-W.
		if id, ok := c.byName[strings.ToLower(name[:len(name)-len(" Women")])]; ok {
			base := c.teams[id]
			t.ID, t.Short, t.Colour = base.ID+"-w", base.Short+"-W", base.Colour
		}
	}
	if isNation(name) {
		t.Kind = "international"
	}
	t.ID = c.freeID(t.ID)
	return t, true
}

// competition says which of the app's competitions a match belongs to, or ok=false for one it doesn't show
// (domestic, A sides, youth, other leagues).
func competition(league string, a, b domain.Team) (domain.CompetitionID, bool) {
	if a.Kind == "international" && b.Kind == "international" {
		if worldCup.MatchString(league) && !notWorldCup.MatchString(league) {
			return "world-cup", true
		}
		return "international", true
	}
	for _, l := range leagueNames {
		if l.re.MatchString(league) {
			return l.comp, true
		}
	}
	return "", false
}

func format(card string, comp domain.CompetitionID) (domain.Format, bool) {
	if comp == "the-hundred" {
		return domain.FormatHundred, true
	}
	c := strings.ToLower(card)
	switch {
	case strings.Contains(c, "first-class"):
		return "", false
	case strings.Contains(c, "test"):
		return domain.FormatTest, true
	case strings.Contains(c, "odi"):
		return domain.FormatODI, true
	case strings.Contains(c, "t20"), strings.Contains(c, "twenty20"):
		return domain.FormatT20, true
	}
	return "", false
}

func venue(location string) domain.Venue {
	name, city, _ := strings.Cut(location, ",")
	name, city = strings.TrimSpace(name), strings.TrimSpace(city)
	return domain.Venue{ID: slug(name), Name: name, City: city}
}

func overs(o float64) string { return strconv.FormatFloat(o, 'f', -1, 64) }

// teamInnings reads one side's scoreboard line ("329 & 63/2 (19.4 ov)") as its innings in order. A total with no
// wickets is all out; overs are only shown for the innings in progress or just finished.
func teamInnings(score, teamID string) []domain.InningsScore {
	out := []domain.InningsScore{}
	if strings.TrimSpace(score) == "" {
		return out
	}
	for _, part := range strings.Split(score, "&") {
		g := inningsText.FindStringSubmatch(strings.TrimSpace(part))
		if g == nil {
			continue
		}
		runs, _ := strconv.Atoi(g[1])
		wkts := 10
		if g[2] != "" {
			wkts, _ = strconv.Atoi(g[2])
		}
		out = append(out, domain.InningsScore{TeamID: teamID, Runs: runs, Wickets: wkts, Overs: g[4]})
	}
	return out
}

// battedFirst is the index (0/1) of the side that batted first: the one not chasing a target, else from the toss.
func battedFirst(e esEvent) int {
	for k, c := range e.Competitors {
		if strings.Contains(c.Score, "target") {
			return 1 - k
		}
	}
	for _, n := range e.Notes {
		if n.Type != "toss" {
			continue
		}
		if g := tossText.FindStringSubmatch(n.Text); g != nil {
			k := slices.IndexFunc(e.Competitors, func(c esCompetitor) bool { return strings.EqualFold(strings.TrimSpace(g[1]), c.DisplayName) })
			if k >= 0 {
				if strings.EqualFold(g[2], "bat") {
					return k
				}
				return 1 - k
			}
		}
	}
	return 0
}

// scoreboardInnings interleaves both sides' innings in batting order (first side, second side, first, second).
func scoreboardInnings(e esEvent, ids [2]string) []domain.InningsScore {
	first := battedFirst(e)
	a := teamInnings(e.Competitors[first].Score, ids[first])
	b := teamInnings(e.Competitors[1-first].Score, ids[1-first])
	out := []domain.InningsScore{}
	for k := 0; k < max(len(a), len(b)); k++ {
		if k < len(a) {
			out = append(out, a[k])
		}
		if k < len(b) {
			out = append(out, b[k])
		}
	}
	return out
}

// toMatch maps one scoreboard event, or ok=false when it isn't one the app shows. New teams it needed come back in
// added (kept only for a match that's kept). The squads, XIs, leaders and player of the match come from the
// match's summary, fetched separately (applySummary).
func (c *catalog) toMatch(e esEvent) (out domain.Match, added []domain.Team, ok bool) {
	if len(e.Competitors) != 2 {
		return domain.Match{}, nil, false
	}
	start, err := time.Parse(time.RFC3339, e.Date)
	if err != nil {
		if start, err = time.Parse("2006-01-02T15:04Z07:00", e.Date); err != nil {
			return domain.Match{}, nil, false
		}
	}
	var teams [2]domain.Team
	for k, ec := range e.Competitors {
		t, isNew := c.team(ec)
		teams[k] = t
		if isNew {
			added = append(added, t)
		}
	}
	if teams[0].ID == teams[1].ID {
		return domain.Match{}, nil, false
	}
	comp, ok := competition(e.League, teams[0], teams[1])
	if !ok {
		return domain.Match{}, nil, false
	}
	f, ok := format(e.Class.Card, comp)
	if !ok {
		return domain.Match{}, nil, false
	}
	ids := [2]string{teams[0].ID, teams[1].ID}
	out = domain.Match{
		ID: matchPrefix + e.ID, Kind: "fixture", CompetitionID: comp, Format: f, Gender: teams[0].Gender,
		Stage: dayNight.ReplaceAllString(e.Title, ""), TeamIDs: ids, Venue: venue(e.Location),
		StartTime: start.UTC().Format(time.RFC3339), State: domain.Upcoming, XI: map[string][]string{},
	}
	status := cmpOr(e.FullStatus.LongSummary, e.FullStatus.Summary)
	switch e.Status {
	case "post":
		out.State = domain.Completed
		out.Result = &domain.MatchResult{Text: status, Innings: scoreboardInnings(e, ids), TopRunScorerIDs: []string{}, TopWicketTakerIDs: []string{}}
		for k, ec := range e.Competitors {
			if ec.Winner {
				out.Result.WinnerID = &ids[k]
			}
		}
	case "in":
		out.State = domain.Live
		out.Live = &domain.LiveScore{Innings: scoreboardInnings(e, ids), StatusText: status}
		for k, ec := range e.Competitors {
			if ec.ID == strconv.Itoa(e.FullStatus.BattingTeamID) {
				out.Live.BattingTeamID = ids[k]
			}
		}
		if out.Live.BattingTeamID == "" && len(out.Live.Innings) > 0 {
			out.Live.BattingTeamID = out.Live.Innings[len(out.Live.Innings)-1].TeamID
		}
	}
	return out, added, true
}

func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func role(a esAthlete) string {
	r := strings.ToLower(a.Position.Name)
	switch {
	case bool(a.Keeper) || strings.Contains(r, "wicketkeeper"):
		return "WK"
	case strings.Contains(r, "allrounder"):
		return "AR"
	case strings.Contains(r, "bowler"):
		return "BOWL"
	}
	return "BAT"
}

// applySummary folds a match's summary into it: the playing XIs once the toss is done (else the announced squads),
// innings from the per-innings lines, the venue's country, and for a finished match the top run-scorers,
// wicket-takers and player of the match. espnTeams are the feed's ids for the match's two sides. It returns the
// players it met.
func applySummary(m *domain.Match, s esSummary, espnTeams [2]string) []domain.Player {
	side := func(espnID string) (string, bool) {
		for k, id := range espnTeams {
			if id == espnID {
				return m.TeamIDs[k], true
			}
		}
		return "", false
	}
	players := []domain.Player{}
	seen := map[string]bool{}
	note := func(a esAthlete, team string) string {
		id := playerPrefix + a.ID
		if !seen[id] {
			seen[id] = true
			players = append(players, domain.Player{ID: id, Name: a.DisplayName, TeamID: team, Role: role(a)})
		}
		return id
	}

	xi := map[string][]string{}
	for _, r := range s.Rosters {
		if team, ok := side(r.Team.ID); ok {
			for _, p := range r.Roster {
				if p.Starter && p.Athlete.ID != "" {
					xi[team] = append(xi[team], note(p.Athlete, team))
				}
			}
		}
	}
	if len(xi) < 2 {
		xi = map[string][]string{}
		for _, sq := range s.Squads {
			if team, ok := side(sq.Team.ID); ok {
				for _, a := range sq.Athletes {
					if a.ID != "" {
						xi[team] = append(xi[team], note(a, team))
					}
				}
			}
		}
	}
	if len(xi) > 0 {
		m.XI = xi
	}

	if v := s.GameInfo.Venue; v.FullName != "" {
		m.Venue = venue(v.FullName)
		if v.Address.City != "" {
			m.Venue.City = v.Address.City
		}
		m.Venue.Country = v.Address.Country
	}

	if len(s.Header.Competitions) == 0 {
		return players
	}
	comp := s.Header.Competitions[0]
	type inn struct {
		period int
		score  domain.InningsScore
	}
	var all []inn
	for _, c := range comp.Competitors {
		team, ok := side(c.ID)
		if !ok {
			continue
		}
		for _, l := range c.Linescores {
			if l.IsBatting && (l.Runs > 0 || l.Overs > 0 || l.Wickets > 0) {
				all = append(all, inn{l.Period, domain.InningsScore{TeamID: team, Runs: l.Runs, Wickets: l.Wickets, Overs: overs(l.Overs)}})
			}
		}
	}
	slices.SortFunc(all, func(a, b inn) int { return a.period - b.period })
	if len(all) > 0 {
		innings := make([]domain.InningsScore, len(all))
		for k, x := range all {
			innings[k] = x.score
		}
		switch {
		case m.Live != nil:
			m.Live.Innings = innings
		case m.Result != nil:
			m.Result.Innings = innings
		}
	}

	if m.Result != nil {
		runs, wkts := map[string]int{}, map[string]int{}
		for _, t := range s.Leaders {
			team, _ := side(t.Team.ID)
			for _, ls := range t.Linescores {
				for _, cat := range ls.Leaders {
					for _, l := range cat.Leaders {
						n, err := strconv.Atoi(l.Value)
						if err != nil || l.Athlete.ID == "" {
							continue
						}
						switch {
						case cat.Name == "runs" && bool(ls.IsBatting):
							runs[note(l.Athlete, team)] += n
						case cat.Name == "wickets" && !bool(ls.IsBatting):
							wkts[note(l.Athlete, team)] += n
						}
					}
				}
			}
		}
		m.Result.TopRunScorerIDs, m.Result.TopWicketTakerIDs = leaders(runs), leaders(wkts)
		for _, fa := range comp.Status.FeaturedAthletes {
			if fa.Name == "playerOfTheMatch" && fa.Athlete.ID != "" {
				id := note(fa.Athlete, "")
				m.Result.OfficialPotmID = &id
			}
		}
	}
	return players
}

// leaders are the players with the most (ties share it); none when nobody scored.
func leaders(by map[string]int) []string {
	best := 0
	for _, n := range by {
		best = max(best, n)
	}
	out := []string{}
	if best == 0 {
		return out
	}
	for id, n := range by {
		if n == best {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out
}
