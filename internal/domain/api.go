package domain

// Response shapes — mirror src/domain/api.ts exactly. Never encode a nil slice or map.

type PlayerAverage struct {
	PlayerID string  `json:"playerId"`
	Avg      float64 `json:"avg"`
	Count    int     `json:"count"`
}

type FeedItem struct {
	Log       MatchLog `json:"log"`
	User      User     `json:"user"`
	Match     Match    `json:"match"`
	Likes     int      `json:"likes"`
	Liked     bool     `json:"liked"`
	ThereToo  int      `json:"thereToo"`
	IWasThere bool     `json:"iWasThere"`
	Replies   []Reply  `json:"replies"`
}

type WatchShare struct {
	Place WatchPlace `json:"place"`
	Share float64    `json:"share"`
}

type WatchCount struct {
	Place WatchPlace `json:"place"`
	Count int        `json:"count"`
}

type FanSummary struct {
	RatingAvg      float64         `json:"ratingAvg"`
	RatingCount    int             `json:"ratingCount"`
	Buckets        []int           `json:"buckets"` // always 10
	Players        []PlayerAverage `json:"players"`
	FansPotmID     *string         `json:"fansPotmId"`
	OfficialPotmID *string         `json:"officialPotmId"`
	FansDisagree   bool            `json:"fansDisagree"`
	Watch          []WatchShare    `json:"watch"`
	Reviews        []FeedItem      `json:"reviews"`
	FriendReviews  []FeedItem      `json:"friendReviews"`
}

type FriendPick struct {
	User   User   `json:"user"`
	TeamID string `json:"teamId"`
}

type PredictionSplit struct {
	Total        int                `json:"total"`
	Shares       map[string]float64 `json:"shares"`
	FriendsTotal int                `json:"friendsTotal"`
	FriendShares map[string]float64 `json:"friendShares"`
	FriendPicks  []FriendPick       `json:"friendPicks"`
}

type PredictionResultRow struct {
	User       User              `json:"user"`
	Prediction Prediction        `json:"prediction"`
	Outcome    PredictionOutcome `json:"outcome"`
}

type FanStats struct {
	Logged         int          `json:"logged"`
	AvgRatingGiven *float64     `json:"avgRatingGiven"`
	Accuracy       *float64     `json:"accuracy"`
	Points         int          `json:"points"`
	Streak         int          `json:"streak"`
	Rank           int          `json:"rank"`
	RankOf         int          `json:"rankOf"`
	Badges         []BadgeDef   `json:"badges"`
	Watch          []WatchCount `json:"watch"`
	Stadiums       []Venue      `json:"stadiums"`
}

type RatedMatch struct {
	Match       Match   `json:"match"`
	RatingAvg   float64 `json:"ratingAvg"`
	RatingCount int     `json:"ratingCount"`
}

type JustFinished struct {
	Match       Match     `json:"match"`
	RatingAvg   float64   `json:"ratingAvg"`
	RatingCount int       `json:"ratingCount"`
	TopReview   *FeedItem `json:"topReview"`
}

type BoardRow struct {
	User   User `json:"user"`
	Points int  `json:"points"`
	IsMe   bool `json:"isMe"`
}

type AuthStartResponse struct {
	Phone            string `json:"phone"`
	ResendInSeconds  int    `json:"resendInSeconds"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
	DevCode          string `json:"devCode,omitempty"`
}

type AuthVerifyResponse struct {
	Token   string  `json:"token"`
	Me      User    `json:"me"`
	Account Account `json:"account"`
	IsNew   bool    `json:"isNew"`
}

type HandleCheckResponse struct {
	Handle    string  `json:"handle"`
	Available bool    `json:"available"`
	Reason    *string `json:"reason"`  // "invalid" | "reserved" | "taken" | null
	Message   *string `json:"message"` // human text when not available
}

type ContactsMatchResponse struct {
	Matches []FriendRow `json:"matches"`
}

type BootstrapResponse struct {
	Token      string            `json:"token"`
	Me         User              `json:"me"`
	Account    Account           `json:"account"`
	Teams      map[string]Team   `json:"teams"`
	Players    map[string]Player `json:"players"`
	ServerTime string            `json:"serverTime"`
}

type HomeResponse struct {
	OnNow         []Match          `json:"onNow"`
	YourCall      *Match           `json:"yourCall"`
	YourCallSplit *PredictionSplit `json:"yourCallSplit"`
	YourCallMine  *Prediction      `json:"yourCallMine"`
	Feed          []FeedItem       `json:"feed"`
	JustFinished  []JustFinished   `json:"justFinished"`
	Board         []BoardRow       `json:"board"`
	Unread        int              `json:"unread"`
}

type FanRating struct {
	Avg   float64 `json:"avg"`
	Count int     `json:"count"`
}

type MatchesResponse struct {
	Matches    []Match              `json:"matches"`
	FanRatings map[string]FanRating `json:"fanRatings"`
}

type MatchMine struct {
	Prediction *Prediction        `json:"prediction"`
	Outcome    *PredictionOutcome `json:"outcome"`
	Log        *MatchLog          `json:"log"`
	Watching   *WatchPlace        `json:"watching"`
}

type MatchPageResponse struct {
	Match           Match                 `json:"match"`
	Locked          bool                  `json:"locked"`
	Split           PredictionSplit       `json:"split"`
	Mine            MatchMine             `json:"mine"`
	FriendsWatching []User                `json:"friendsWatching"`
	Fans            *FanSummary           `json:"fans"`
	Results         []PredictionResultRow `json:"results"`
}

type CollectionShelf struct {
	Collection Collection   `json:"collection"`
	Matches    []RatedMatch `json:"matches"`
}

type OnThisDay struct {
	Match    Match `json:"match"`
	YearsAgo int   `json:"yearsAgo"`
	Exact    bool  `json:"exact"`
}

type VaultResponse struct {
	Collections []CollectionShelf    `json:"collections"`
	OnThisDay   *OnThisDay           `json:"onThisDay"`
	Top5        []Match              `json:"top5"`
	FanRatings  map[string]FanRating `json:"fanRatings"` // for every top5 match
	Eras        []string             `json:"eras"`
}

type RankingsResponse struct {
	Matches []RatedMatch `json:"matches"`
}

type UserResponse struct {
	User      User       `json:"user"`
	IsMe      bool       `json:"isMe"`
	Following bool       `json:"following"`
	Stats     FanStats   `json:"stats"`
	Logs      []FeedItem `json:"logs"`
	Top5      []Match    `json:"top5"`
}

type FriendRow struct {
	User      User `json:"user"`
	Following bool `json:"following"`
}

type FriendsResponse struct {
	Friends []FriendRow `json:"friends"`
}

type SearchResponse struct {
	Matches []Match `json:"matches"`
	Users   []User  `json:"users"`
}

type LogSuggestionsResponse struct {
	Suggested []Match `json:"suggested"`
	Recent    []Match `json:"recent"`
}

type Notification struct {
	ID        string  `json:"id"`
	Text      string  `json:"text"`
	MatchID   *string `json:"matchId"`
	LogID     *string `json:"logId"`
	CreatedAt string  `json:"createdAt"`
}

type NotificationsResponse struct {
	Items []Notification `json:"items"`
}

// ChatResponse is GET /v1/matches/{id}/chat: the room's messages and who wrote them.
type ChatResponse struct {
	Messages []ChatMessage   `json:"messages"` // oldest first; with ?after=, only those from then on
	Users    map[string]User `json:"users"`    // every author in Messages
	Open     bool            `json:"open"`     // fans can post now
	OpensAt  string          `json:"opensAt"`  // when a closed room opens ("" once it has closed, and for classics)
}
