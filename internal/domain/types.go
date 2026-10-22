// Package domain mirrors src/domain/types.ts and src/domain/api.ts — that TypeScript is the JSON contract.
// Rules: camelCase tags; TS `x?` fields are pointers with omitempty; TS `X | null` fields are pointers WITHOUT
// omitempty; slices and maps must never be nil when encoded (use the New*/Ensure helpers or literal {}).
package domain

import "slices"

type Format string

const (
	FormatT20     Format = "T20"
	FormatODI     Format = "ODI"
	FormatTest    Format = "Test"
	FormatHundred Format = "The Hundred"
)

type Gender string

const (
	Men   Gender = "men"
	Women Gender = "women"
)

// CompetitionID values: international, ipl, wpl, world-cup, bbl, sa20, the-hundred.
type CompetitionID string

type Team struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Short  string `json:"short"`
	Colour string `json:"colour"`
	Gender Gender `json:"gender"`
	Kind   string `json:"kind"` // "international" | "franchise"
}

// PlayerRole values: BAT, BOWL, AR, WK.
type Player struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	TeamID string `json:"teamId"`
	Role   string `json:"role"`
}

type Venue struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type MatchState string

const (
	Upcoming  MatchState = "upcoming"
	Live      MatchState = "live"
	Completed MatchState = "completed"
)

type InningsScore struct {
	TeamID  string `json:"teamId"`
	Runs    int    `json:"runs"`
	Wickets int    `json:"wickets"`
	Overs   string `json:"overs"`
}

type LiveScore struct {
	Innings       []InningsScore `json:"innings"`
	BattingTeamID string         `json:"battingTeamId"`
	StatusText    string         `json:"statusText"`
}

type MatchResult struct {
	WinnerID          *string        `json:"winnerId"` // null = tie / draw / no result
	Text              string         `json:"text"`
	Innings           []InningsScore `json:"innings"`
	TopRunScorerIDs   []string       `json:"topRunScorerIds"`
	TopWicketTakerIDs []string       `json:"topWicketTakerIds"`
	OfficialPotmID    *string        `json:"officialPotmId"`
}

// CollectionID values: world-cup-finals, greatest-chases, test-thrillers, super-over-madness, womens-classics.
type CollectionID string

type Collection struct {
	ID    CollectionID `json:"id"`
	Title string       `json:"title"`
	Blurb string       `json:"blurb"`
}

type ClassicInfo struct {
	Date            string         `json:"date"` // YYYY-MM-DD
	CompetitionName string         `json:"competitionName"`
	Headline        string         `json:"headline"`
	Scoreline       string         `json:"scoreline"`
	PotmName        string         `json:"potmName,omitempty"`
	Collections     []CollectionID `json:"collections"`
	Era             string         `json:"era"` // "2010s"
}

type Match struct {
	ID            string              `json:"id"`
	Kind          string              `json:"kind"` // "fixture" | "classic"
	CompetitionID CompetitionID       `json:"competitionId"`
	Format        Format              `json:"format"`
	Gender        Gender              `json:"gender"`
	Stage         string              `json:"stage"`
	TeamIDs       [2]string           `json:"teamIds"`
	Venue         Venue               `json:"venue"`
	StartTime     string              `json:"startTime"` // RFC 3339; predictions lock here
	State         MatchState          `json:"state"`
	Live          *LiveScore          `json:"live,omitempty"`
	Result        *MatchResult        `json:"result,omitempty"`
	XI            map[string][]string `json:"xi,omitempty"` // fixtures only
	Classic       *ClassicInfo        `json:"classic,omitempty"`
}

type WatchPlace string

const (
	WatchStadium WatchPlace = "stadium"
	WatchHome    WatchPlace = "home"
	WatchFriends WatchPlace = "friends"
	WatchHostel  WatchPlace = "hostel"
	WatchOffice  WatchPlace = "office"
	WatchCafe    WatchPlace = "cafe"
	WatchOnTheGo WatchPlace = "on-the-go"
)

var WatchPlaces = []WatchPlace{WatchStadium, WatchHome, WatchFriends, WatchHostel, WatchOffice, WatchCafe, WatchOnTheGo}

func ValidWatchPlace(p WatchPlace) bool { return slices.Contains(WatchPlaces, p) }

type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Handle          string `json:"handle"`
	AvatarColour    string `json:"avatarColour"`
	FavouriteTeamID string `json:"favouriteTeamId"`
	Bio             string `json:"bio"`
	PhotoURI        string `json:"photoUri,omitempty"`
}

type MatchLog struct {
	ID             string         `json:"id"`
	UserID         string         `json:"userId"`
	MatchID        string         `json:"matchId"`
	Rating         float64        `json:"rating"` // 0.5–5 in 0.5 steps
	InstantClassic bool           `json:"instantClassic"`
	PlayerRatings  map[string]int `json:"playerRatings"` // playerId → 1..10
	FanPotmID      *string        `json:"fanPotmId"`
	WatchedAt      WatchPlace     `json:"watchedAt"`
	PlaceNote      string         `json:"placeNote,omitempty"`
	WatchedWith    []string       `json:"watchedWith"`
	Companions     *int           `json:"companions,omitempty"`
	PhotoURI       string         `json:"photoUri,omitempty"`
	Review         string         `json:"review,omitempty"`
	CreatedAt      string         `json:"createdAt"`
	StubNo         int            `json:"stubNo"`
}

type Prediction struct {
	UserID           string `json:"userId"`
	MatchID          string `json:"matchId"`
	WinnerID         string `json:"winnerId"`
	TopRunScorerID   string `json:"topRunScorerId"`
	TopWicketTakerID string `json:"topWicketTakerId"`
	CreatedAt        string `json:"createdAt"`
}

type PredictionOutcome struct {
	MatchID          string  `json:"matchId"`
	UserID           string  `json:"userId"`
	WinnerCorrect    bool    `json:"winnerCorrect"`
	TopScorerCorrect bool    `json:"topScorerCorrect"`
	TopWicketCorrect bool    `json:"topWicketCorrect"`
	Upset            bool    `json:"upset"`
	Base             int     `json:"base"`
	Multiplier       float64 `json:"multiplier"`
	Points           int     `json:"points"`
	Streak           int     `json:"streak"`
	Settled          bool    `json:"settled"`
}

type Reply struct {
	ID        string `json:"id"`
	LogID     string `json:"logId"`
	UserID    string `json:"userId"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

// BadgeDef IDs: called-the-upset, five-match-streak, sharpshooter, first-call, stadium-regular.
type BadgeDef struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Ptr returns a pointer to v (handy for nullable JSON fields).
func Ptr[T any](v T) *T { return &v }
