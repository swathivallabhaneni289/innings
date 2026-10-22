package httpapi

import (
	"fmt"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/sim"
)

// fakeWorld is a small populated world (live, upcoming, just-finished and older fixtures, a classic, friends and
// their logs/predictions) so the merge logic is exercised even before the real sim is filled in.
type fakeWorld struct {
	now      time.Time
	teams    map[string]domain.Team
	players  map[string]domain.Player
	fixtures []domain.Match
	classics []domain.Match
	friends  []domain.User
	public   []domain.User
	logs     []domain.MatchLog
	preds    []domain.Prediction
}

func xi(team string) []string {
	out := make([]string, 11)
	for i := range out {
		out[i] = fmt.Sprintf("%s-p%d", team, i)
	}
	return out
}

func newFakeWorld(now time.Time) *fakeWorld {
	w := &fakeWorld{now: now, teams: map[string]domain.Team{}, players: map[string]domain.Player{}}
	for _, t := range []domain.Team{
		{ID: "ind", Name: "India", Short: "IND", Colour: "#1f4fa0", Gender: domain.Men, Kind: "international"},
		{ID: "aus", Name: "Australia", Short: "AUS", Colour: "#f2c200", Gender: domain.Men, Kind: "international"},
		{ID: "csk", Name: "Chennai Super Kings", Short: "CSK", Colour: "#f9cd05", Gender: domain.Men, Kind: "franchise"},
		{ID: "mi", Name: "Mumbai Indians", Short: "MI", Colour: "#004ba0", Gender: domain.Men, Kind: "franchise"},
	} {
		w.teams[t.ID] = t
		for _, id := range xi(t.ID) {
			w.players[id] = domain.Player{ID: id, Name: id, TeamID: t.ID, Role: "BAT"}
		}
	}
	venue := domain.Venue{ID: "wankhede", Name: "Wankhede Stadium", City: "Mumbai", Country: "India"}
	fixture := func(id string, a, b string, comp domain.CompetitionID, start time.Time, state domain.MatchState) domain.Match {
		m := domain.Match{ID: id, Kind: "fixture", CompetitionID: comp, Format: domain.FormatT20, Gender: domain.Men, Stage: "Match " + id,
			TeamIDs: [2]string{a, b}, Venue: venue, StartTime: start.UTC().Format(time.RFC3339), State: state,
			XI: map[string][]string{a: xi(a), b: xi(b)}}
		switch state {
		case domain.Live:
			m.Live = &domain.LiveScore{BattingTeamID: a, StatusText: "IND 120/3"} // nil Innings on purpose: normMatch must fix it
		case domain.Completed:
			m.Result = &domain.MatchResult{WinnerID: domain.Ptr(a), Text: a + " won", TopRunScorerIDs: []string{a + "-p0"},
				TopWicketTakerIDs: []string{b + "-p10", a + "-p10"}, OfficialPotmID: domain.Ptr(a + "-p0")}
		}
		return m
	}
	w.fixtures = []domain.Match{
		fixture("old", "ind", "aus", "international", now.Add(-72*time.Hour), domain.Completed),
		fixture("done", "csk", "mi", "ipl", now.Add(-5*time.Hour), domain.Completed),
		fixture("live", "ind", "aus", "international", now.Add(-time.Hour), domain.Live),
		fixture("soon", "mi", "csk", "ipl", now.Add(time.Hour), domain.Upcoming),
		fixture("big", "ind", "aus", "world-cup", now.Add(20*time.Hour), domain.Upcoming),
	}
	w.classics = []domain.Match{{
		ID: "wc-2011-final", Kind: "classic", CompetitionID: "world-cup", Format: domain.FormatODI, Gender: domain.Men, Stage: "Final",
		TeamIDs: [2]string{"ind", "aus"}, Venue: venue, StartTime: "2011-04-02T09:00:00Z", State: domain.Completed,
		Result:  &domain.MatchResult{WinnerID: domain.Ptr("ind"), Text: "India won by 6 wickets"},
		Classic: &domain.ClassicInfo{Date: "2011-04-02", CompetitionName: "ICC Cricket World Cup 2011", Headline: "Dhoni finishes it", Collections: []domain.CollectionID{"world-cup-finals"}, Era: "2010s"},
	}}
	w.friends = []domain.User{
		{ID: "f-aditi", Name: "Aditi Rao", Handle: "aditi", AvatarColour: "#123", FavouriteTeamID: "ind", Bio: "Dhoni finishes, I cry"},
		{ID: "f-kabir", Name: "Kabir", Handle: "kabir", AvatarColour: "#456", FavouriteTeamID: "mi"},
	}
	w.public = []domain.User{{ID: "p-critic", Name: "The Critic", Handle: "critic", AvatarColour: "#789", FavouriteTeamID: "aus"}}
	at := func(d time.Duration) string { return now.Add(d).UTC().Format(time.RFC3339) }
	w.logs = []domain.MatchLog{
		{ID: "sl-1", UserID: "f-aditi", MatchID: "done", Rating: 4.5, PlayerRatings: map[string]int{"csk-p0": 9, "mi-p1": 6}, WatchedAt: domain.WatchStadium, Review: "Unreal finish", CreatedAt: at(-time.Hour), StubNo: 3},
		{ID: "sl-2", UserID: "f-kabir", MatchID: "done", Rating: 3, PlayerRatings: map[string]int{"csk-p0": 8}, WatchedAt: domain.WatchHome, CreatedAt: at(-50 * time.Minute), StubNo: 7}, // nil WatchedWith on purpose
		{ID: "sl-3", UserID: "p-critic", MatchID: "done", Rating: 2.5, PlayerRatings: map[string]int{"csk-p0": 7}, WatchedAt: domain.WatchCafe, Review: "Overrated", CreatedAt: at(-40 * time.Minute), StubNo: 1, WatchedWith: []string{}},
		{ID: "sl-4", UserID: "f-aditi", MatchID: "wc-2011-final", Rating: 5, WatchedAt: domain.WatchFriends, Review: "Never forget", CreatedAt: at(-48 * time.Hour), StubNo: 1},
	}
	w.preds = []domain.Prediction{
		{UserID: "f-aditi", MatchID: "done", WinnerID: "csk", TopRunScorerID: "csk-p0", TopWicketTakerID: "mi-p10", CreatedAt: at(-6 * time.Hour)},
		{UserID: "f-kabir", MatchID: "done", WinnerID: "mi", TopRunScorerID: "mi-p0", TopWicketTakerID: "csk-p10", CreatedAt: at(-6 * time.Hour)},
		{UserID: "f-aditi", MatchID: "big", WinnerID: "ind", TopRunScorerID: "ind-p0", TopWicketTakerID: "aus-p10", CreatedAt: at(-time.Hour)},
	}
	return w
}

func (w *fakeWorld) Teams() map[string]domain.Team             { return w.teams }
func (w *fakeWorld) Players() map[string]domain.Player         { return w.players }
func (w *fakeWorld) Fixtures(time.Time) []domain.Match         { return w.fixtures }
func (w *fakeWorld) Classics() []domain.Match                  { return w.classics }
func (w *fakeWorld) Friends() []domain.User                    { return w.friends }
func (w *fakeWorld) PublicFans() []domain.User                 { return w.public }
func (w *fakeWorld) Logs(time.Time) []domain.MatchLog          { return w.logs }
func (w *fakeWorld) Predictions(time.Time) []domain.Prediction { return w.preds }
func (w *fakeWorld) BaseLikes(id string) int {
	if id == "sl-1" {
		return 12
	}
	return 0
}
func (w *fakeWorld) BaseThereToo(string) int { return 2 }
func (w *fakeWorld) Replies(id string, _ time.Time) []domain.Reply {
	if id == "sl-1" {
		return []domain.Reply{{ID: "sr-1", LogID: id, UserID: "f-kabir", Text: "Same!", CreatedAt: w.now.Add(-30 * time.Minute).UTC().Format(time.RFC3339)}}
	}
	return []domain.Reply{}
}

func (w *fakeWorld) User(id string) (domain.User, bool) {
	for _, u := range append(append([]domain.User{}, w.friends...), w.public...) {
		if u.ID == id {
			return u, true
		}
	}
	return domain.User{}, false
}

func (w *fakeWorld) Crowd(matchID string) sim.Crowd {
	c := sim.Crowd{RatingBuckets: make([]int, 10), PlayerTallies: map[string]sim.Tally{}, Watch: map[domain.WatchPlace]int{}, Picks: map[string]int{}}
	if matchID == "done" {
		c.RatingBuckets[7] = 10 // ten 4-ball ratings
		c.PlayerTallies["mi-p1"] = sim.Tally{Sum: 30, Count: 3}
		c.Watch[domain.WatchHome] = 10
		c.Picks["csk"], c.Picks["mi"] = 2, 8 // CSK is the upset winner
	}
	return c
}

func (w *fakeWorld) FriendsWatching(matchID string, _ time.Time) []string {
	if matchID == "live" {
		return []string{"f-aditi", "f-kabir"}
	}
	return []string{}
}

func (w *fakeWorld) DemoHistory(userID string, now time.Time) ([]domain.MatchLog, []domain.Prediction, []string) {
	at := now.Add(-70 * time.Hour).UTC().Format(time.RFC3339)
	logs := []domain.MatchLog{
		{ID: "demo", MatchID: "old", Rating: 4, WatchedAt: domain.WatchStadium, Review: "Loud night", CreatedAt: at},
		{ID: "demo", MatchID: "wc-2011-final", Rating: 5, WatchedAt: domain.WatchHome, CreatedAt: at},
	}
	preds := []domain.Prediction{{MatchID: "old", WinnerID: "ind", TopRunScorerID: "ind-p0", TopWicketTakerID: "aus-p10", CreatedAt: now.Add(-73 * time.Hour).UTC().Format(time.RFC3339)}}
	return logs, preds, []string{"f-aditi"}
}
