package httpapi

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// view is one request's merged picture of the world: simulated data + real users' rows, loaded lazily and cached
// for the request. Store failures panic with storeErr (→ 500). Data volumes are demo-sized, so whole tables are
// loaded; at scale these would become scoped queries.
type view struct {
	s   *Server
	ctx context.Context
	now time.Time
	me  domain.User

	fixtureList []domain.Match
	classicList []domain.Match
	matchIdx    map[string]domain.Match
	teamMap     map[string]domain.Team

	realUsers   map[string]domain.User
	realOrder   []domain.User
	listedOrder []domain.User       // real users who finished sign-up (Store.ListedUsers)
	listedIDs   map[string]bool     // ids of listedOrder
	follows     map[string][]string // real user id → ids they follow

	logList     []domain.MatchLog // sim + real, newest first
	logIdx      map[string]domain.MatchLog
	realLog     map[string]bool
	logsByMatch map[string][]domain.MatchLog
	logsByUser  map[string][]domain.MatchLog

	predsByMatch map[string][]domain.Prediction
	predsByUser  map[string][]domain.Prediction

	reactions map[string][]store.Reaction // table → rows
	replyIdx  map[string][]domain.Reply   // real replies by log id

	shareCache   map[string]shareInfo
	outcomeCache map[string][]domain.PredictionOutcome
}

type shareInfo struct {
	picks map[string]int
}

func (s *Server) newView(ctx context.Context, me domain.User) *view {
	return &view{s: s, ctx: ctx, now: s.Now(), me: me, shareCache: map[string]shareInfo{}, outcomeCache: map[string][]domain.PredictionOutcome{}}
}

// ---- matches ----

func (v *view) fixtures() []domain.Match {
	if v.fixtureList == nil {
		v.fixtureList = []domain.Match{}
		for _, m := range v.s.World.Fixtures(v.now) {
			v.fixtureList = append(v.fixtureList, normMatch(m))
		}
	}
	return v.fixtureList
}

func (v *view) classics() []domain.Match {
	if v.classicList == nil {
		v.classicList = []domain.Match{}
		for _, m := range v.s.World.Classics() {
			v.classicList = append(v.classicList, normMatch(m))
		}
	}
	return v.classicList
}

func (v *view) match(id string) (domain.Match, bool) {
	if v.matchIdx == nil {
		v.matchIdx = map[string]domain.Match{}
		for _, m := range v.classics() {
			v.matchIdx[m.ID] = m
		}
		for _, m := range v.fixtures() {
			v.matchIdx[m.ID] = m
		}
	}
	m, ok := v.matchIdx[id]
	return m, ok
}

func (v *view) teams() map[string]domain.Team {
	if v.teamMap == nil {
		v.teamMap = nonNilMap(v.s.World.Teams())
	}
	return v.teamMap
}

func (v *view) teamShort(id string) string {
	if t, ok := v.teams()[id]; ok && t.Short != "" {
		return t.Short
	}
	return strings.ToUpper(id)
}

func (v *view) title(m domain.Match) string {
	return v.teamShort(m.TeamIDs[0]) + " v " + v.teamShort(m.TeamIDs[1])
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func startOf(m domain.Match) time.Time { return parseTime(m.StartTime) }

// endOf estimates when a match finished (Match has no end time): start + typical duration for the format, capped
// at now for completed matches.
func (v *view) endOf(m domain.Match) time.Time {
	d := 3*time.Hour + 30*time.Minute
	switch m.Format {
	case domain.FormatHundred:
		d = 2*time.Hour + 45*time.Minute
	case domain.FormatODI:
		d = 8 * time.Hour
	case domain.FormatTest:
		d = 5 * 24 * time.Hour
	}
	end := startOf(m).Add(d)
	if m.State == domain.Completed && end.After(v.now) {
		end = v.now
	}
	return end
}

// ---- users ----

func (v *view) loadUsers() {
	if v.realUsers != nil {
		return
	}
	v.realOrder = must(v.s.Store.Users(v.ctx))
	v.realUsers = map[string]domain.User{}
	for _, u := range v.realOrder {
		v.realUsers[u.ID] = u
	}
}

func (v *view) user(id string) (domain.User, bool) {
	if id == v.me.ID {
		return v.me, true
	}
	if u, ok := v.s.World.User(id); ok {
		return u, true
	}
	v.loadUsers()
	u, ok := v.realUsers[id]
	return u, ok
}

// listedUsers are the real users who finished sign-up: the only real users search, friend lists and contact
// matches show. Everyone else (half-signed-up accounts, pre-sign-in guests) still resolves through user() so their
// existing logs, replies and board rows keep working.
func (v *view) listedUsers() []domain.User {
	if v.listedIDs == nil {
		v.listedOrder = must(v.s.Store.ListedUsers(v.ctx))
		v.listedIDs = map[string]bool{}
		for _, u := range v.listedOrder {
			v.listedIDs[u.ID] = true
		}
	}
	return v.listedOrder
}

// listed reports whether id may appear in search and friend lists: simulated fans always, real users once onboarded.
func (v *view) listed(id string) bool {
	if _, ok := v.s.World.User(id); ok {
		return true
	}
	v.listedUsers()
	return v.listedIDs[id]
}

func (v *view) isReal(id string) bool {
	if id == v.me.ID {
		return true
	}
	v.loadUsers()
	_, ok := v.realUsers[id]
	return ok
}

// followsOf returns who userID follows (sim users follow nobody).
func (v *view) followsOf(userID string) []string {
	if v.follows == nil {
		v.follows = map[string][]string{}
	}
	if f, ok := v.follows[userID]; ok {
		return f
	}
	f := []string{}
	if v.isReal(userID) {
		f = must(v.s.Store.Follows(v.ctx, userID))
	}
	v.follows[userID] = f
	return f
}

func (v *view) iFollow(id string) bool { return slices.Contains(v.followsOf(v.me.ID), id) }

// circle is userID + everyone they follow who exists.
func (v *view) circle(userID string) []string {
	out := []string{userID}
	for _, id := range v.followsOf(userID) {
		if _, ok := v.user(id); ok && id != userID {
			out = append(out, id)
		}
	}
	return out
}

func toSet(ids []string) map[string]bool {
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// ---- logs ----

func (v *view) loadLogs() {
	if v.logIdx != nil {
		return
	}
	real := must(v.s.Store.Logs(v.ctx))
	v.realLog = map[string]bool{}
	all := make([]domain.MatchLog, 0, len(real))
	for _, l := range real {
		v.realLog[l.ID] = true
		all = append(all, l)
	}
	for _, l := range v.s.World.Logs(v.now) {
		all = append(all, normLog(l))
	}
	slices.SortStableFunc(all, func(a, b domain.MatchLog) int { return parseTime(b.CreatedAt).Compare(parseTime(a.CreatedAt)) })
	v.logList = all
	v.logIdx = map[string]domain.MatchLog{}
	v.logsByMatch = map[string][]domain.MatchLog{}
	v.logsByUser = map[string][]domain.MatchLog{}
	for _, l := range all {
		v.logIdx[l.ID] = l
		v.logsByMatch[l.MatchID] = append(v.logsByMatch[l.MatchID], l)
		v.logsByUser[l.UserID] = append(v.logsByUser[l.UserID], l)
	}
}

func (v *view) logs() []domain.MatchLog { v.loadLogs(); return v.logList }

func (v *view) logByID(id string) (domain.MatchLog, bool) {
	v.loadLogs()
	l, ok := v.logIdx[id]
	return l, ok
}

func (v *view) matchLogs(matchID string) []domain.MatchLog {
	v.loadLogs()
	return v.logsByMatch[matchID]
}
func (v *view) userLogs(userID string) []domain.MatchLog { v.loadLogs(); return v.logsByUser[userID] }

func (v *view) myLog(matchID string) *domain.MatchLog {
	for _, l := range v.userLogs(v.me.ID) {
		if l.MatchID == matchID {
			return &l
		}
	}
	return nil
}

func (v *view) reactionRows(table string) []store.Reaction {
	if v.reactions == nil {
		v.reactions = map[string][]store.Reaction{}
	}
	rows, ok := v.reactions[table]
	if !ok {
		rows = must(v.s.Store.Reactions(v.ctx, table))
		v.reactions[table] = rows
	}
	return rows
}

// reactionsOn returns the count (sim base + real) and whether I reacted.
func (v *view) reactionsOn(table, logID string) (int, bool) {
	n, mine := 0, false
	v.loadLogs()
	if !v.realLog[logID] {
		if table == store.Likes {
			n = v.s.World.BaseLikes(logID)
		} else {
			n = v.s.World.BaseThereToo(logID)
		}
	}
	for _, r := range v.reactionRows(table) {
		if r.LogID == logID {
			n++
			if r.UserID == v.me.ID {
				mine = true
			}
		}
	}
	return n, mine
}

func (v *view) realReplies() map[string][]domain.Reply {
	if v.replyIdx == nil {
		v.replyIdx = map[string][]domain.Reply{}
		for _, r := range must(v.s.Store.Replies(v.ctx)) {
			v.replyIdx[r.LogID] = append(v.replyIdx[r.LogID], r)
		}
	}
	return v.replyIdx
}

func (v *view) replies(logID string) []domain.Reply {
	out := []domain.Reply{}
	v.loadLogs()
	if !v.realLog[logID] {
		out = append(out, v.s.World.Replies(logID, v.now)...)
	}
	out = append(out, v.realReplies()[logID]...)
	slices.SortStableFunc(out, func(a, b domain.Reply) int { return parseTime(a.CreatedAt).Compare(parseTime(b.CreatedAt)) })
	return out
}

func (v *view) feedItem(l domain.MatchLog) (domain.FeedItem, bool) {
	u, ok := v.user(l.UserID)
	if !ok {
		return domain.FeedItem{}, false
	}
	m, ok := v.match(l.MatchID)
	if !ok {
		return domain.FeedItem{}, false
	}
	likes, liked := v.reactionsOn(store.Likes, l.ID)
	there, was := v.reactionsOn(store.ThereToo, l.ID)
	return domain.FeedItem{Log: normLog(l), User: u, Match: m, Likes: likes, Liked: liked, ThereToo: there, IWasThere: was, Replies: v.replies(l.ID)}, true
}

// feedItems converts logs (keeping their order), dropping any whose user or match is unknown; max ≤ 0 = no limit.
func (v *view) feedItems(logs []domain.MatchLog, max int) []domain.FeedItem {
	out := []domain.FeedItem{}
	for _, l := range logs {
		if max > 0 && len(out) >= max {
			break
		}
		if fi, ok := v.feedItem(l); ok {
			out = append(out, fi)
		}
	}
	return out
}

// ---- predictions ----

func (v *view) loadPreds() {
	if v.predsByMatch != nil {
		return
	}
	v.predsByMatch = map[string][]domain.Prediction{}
	v.predsByUser = map[string][]domain.Prediction{}
	add := func(p domain.Prediction) {
		v.predsByMatch[p.MatchID] = append(v.predsByMatch[p.MatchID], p)
		v.predsByUser[p.UserID] = append(v.predsByUser[p.UserID], p)
	}
	for _, p := range v.s.World.Predictions(v.now) {
		add(p)
	}
	for _, p := range must(v.s.Store.Predictions(v.ctx)) {
		add(p)
	}
}

func (v *view) myPrediction(matchID string) *domain.Prediction {
	v.loadPreds()
	for _, p := range v.predsByMatch[matchID] {
		if p.UserID == v.me.ID {
			return &p
		}
	}
	return nil
}

// picks is the all-fan winner pick count at lock time: crowd + sim friends + real users who predicted before start.
func (v *view) picks(m domain.Match) map[string]int {
	if si, ok := v.shareCache[m.ID]; ok {
		return si.picks
	}
	picks := map[string]int{}
	for team, n := range v.s.World.Crowd(m.ID).Picks {
		picks[team] += n
	}
	v.loadPreds()
	start := startOf(m)
	for _, p := range v.predsByMatch[m.ID] {
		if parseTime(p.CreatedAt).Before(start) {
			picks[p.WinnerID]++
		}
	}
	v.shareCache[m.ID] = shareInfo{picks: picks}
	return picks
}

// winnerShare is the winner's share of all fan picks at lock time (ok=false when nobody predicted).
func (v *view) winnerShare(m domain.Match) (float64, bool) {
	if m.Result == nil || m.Result.WinnerID == nil {
		return 0, false
	}
	shares, total := domain.Shares(m.TeamIDs, v.picks(m))
	if total == 0 {
		return 0, false
	}
	return shares[*m.Result.WinnerID], true
}

func (v *view) split(m domain.Match) domain.PredictionSplit {
	shares, total := domain.Shares(m.TeamIDs, v.picks(m))
	friendCounts := map[string]int{}
	picks := []domain.FriendPick{}
	following := toSet(v.followsOf(v.me.ID))
	v.loadPreds()
	for _, p := range v.predsByMatch[m.ID] {
		if p.UserID == v.me.ID || !following[p.UserID] {
			continue
		}
		u, ok := v.user(p.UserID)
		if !ok {
			continue
		}
		friendCounts[p.WinnerID]++
		picks = append(picks, domain.FriendPick{User: u, TeamID: p.WinnerID})
	}
	fShares, fTotal := domain.Shares(m.TeamIDs, friendCounts)
	return domain.PredictionSplit{Total: total, Shares: shares, FriendsTotal: fTotal, FriendShares: fShares, FriendPicks: picks}
}

// outcomes settles userID's whole prediction history (match start order).
func (v *view) outcomes(userID string) []domain.PredictionOutcome {
	if o, ok := v.outcomeCache[userID]; ok {
		return o
	}
	v.loadPreds()
	o := domain.SettleHistory(v.predsByUser[userID], v.match, v.winnerShare)
	v.outcomeCache[userID] = o
	return o
}

func (v *view) outcomeFor(userID, matchID string) *domain.PredictionOutcome {
	for _, o := range v.outcomes(userID) {
		if o.MatchID == matchID {
			return &o
		}
	}
	return nil
}

func (v *view) ties(userID string) map[string]bool {
	out := map[string]bool{}
	for _, o := range v.outcomes(userID) {
		if m, ok := v.match(o.MatchID); ok && o.Settled && m.Result != nil && m.Result.WinnerID == nil {
			out[o.MatchID] = true
		}
	}
	return out
}

// pointsSince sums settled points from matches that finished at or after since.
func (v *view) pointsSince(userID string, since time.Time) int {
	n := 0
	for _, o := range v.outcomes(userID) {
		if !o.Settled {
			continue
		}
		if m, ok := v.match(o.MatchID); ok && !v.endOf(m).Before(since) {
			n += o.Points
		}
	}
	return n
}

// ---- fan aggregates ----

func (v *view) buckets(m domain.Match) []int {
	b := make([]int, 10)
	copy(b, v.s.World.Crowd(m.ID).RatingBuckets)
	for _, l := range v.matchLogs(m.ID) {
		domain.AddRating(b, l.Rating)
	}
	return b
}

func (v *view) fanRating(m domain.Match) domain.FanRating {
	avg, n := domain.BucketAverage(v.buckets(m))
	return domain.FanRating{Avg: avg, Count: n}
}

func (v *view) rated(m domain.Match) domain.RatedMatch {
	r := v.fanRating(m)
	return domain.RatedMatch{Match: m, RatingAvg: r.Avg, RatingCount: r.Count}
}

// reviewItems are the match's reviews, most liked first (ties: newest first).
func (v *view) reviewItems(m domain.Match) []domain.FeedItem {
	var withReview []domain.MatchLog
	for _, l := range v.matchLogs(m.ID) {
		if strings.TrimSpace(l.Review) != "" {
			withReview = append(withReview, l)
		}
	}
	items := v.feedItems(withReview, 0)
	slices.SortStableFunc(items, func(a, b domain.FeedItem) int { return cmp.Compare(b.Likes, a.Likes) })
	return items
}

func (v *view) fanSummary(m domain.Match) domain.FanSummary {
	crowd := v.s.World.Crowd(m.ID)
	buckets := v.buckets(m)
	tallies := map[string]domain.Tally{}
	for id, t := range crowd.PlayerTallies {
		tallies[id] = domain.Tally{Sum: t.Sum, Count: t.Count}
	}
	watch := map[domain.WatchPlace]int{}
	for p, n := range crowd.Watch {
		watch[p] += n
	}
	for _, l := range v.matchLogs(m.ID) {
		for pid, r := range l.PlayerRatings {
			t := tallies[pid]
			t.Sum += r
			t.Count++
			tallies[pid] = t
		}
		watch[l.WatchedAt]++
	}
	avg, n := domain.BucketAverage(buckets)
	players := domain.PlayerAverages(tallies)
	fans := domain.FansPotm(players)
	var official *string
	if m.Result != nil {
		official = m.Result.OfficialPotmID
	}
	reviews := v.reviewItems(m)
	circle := toSet(v.circle(v.me.ID))
	friendReviews := []domain.FeedItem{}
	for _, fi := range reviews {
		if circle[fi.User.ID] {
			friendReviews = append(friendReviews, fi)
		}
	}
	slices.SortStableFunc(friendReviews, func(a, b domain.FeedItem) int {
		return parseTime(b.Log.CreatedAt).Compare(parseTime(a.Log.CreatedAt))
	})
	if len(reviews) > 10 {
		reviews = reviews[:10]
	}
	return domain.FanSummary{
		RatingAvg: avg, RatingCount: n, Buckets: buckets, Players: players,
		FansPotmID: fans, OfficialPotmID: official,
		FansDisagree: fans != nil && official != nil && *fans != *official,
		Watch:        domain.WatchShares(watch), Reviews: reviews, FriendReviews: friendReviews,
	}
}

// ---- normalisation: never encode nil slices/maps, whatever the sim returns ----

func normMatch(m domain.Match) domain.Match {
	if m.Result != nil {
		r := *m.Result
		r.Innings = nonNil(r.Innings)
		r.TopRunScorerIDs = nonNil(r.TopRunScorerIDs)
		r.TopWicketTakerIDs = nonNil(r.TopWicketTakerIDs)
		m.Result = &r
	}
	if m.Live != nil {
		l := *m.Live
		l.Innings = nonNil(l.Innings)
		m.Live = &l
	}
	if m.Classic != nil {
		c := *m.Classic
		c.Collections = nonNil(c.Collections)
		m.Classic = &c
	}
	if m.XI != nil {
		xi := make(map[string][]string, len(m.XI))
		for k, ids := range m.XI {
			xi[k] = nonNil(ids)
		}
		m.XI = xi
	}
	return m
}

func normLog(l domain.MatchLog) domain.MatchLog {
	if l.PlayerRatings == nil {
		l.PlayerRatings = map[string]int{}
	}
	l.WatchedWith = nonNil(l.WatchedWith)
	return l
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func nonNilMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return map[K]V{}
	}
	return m
}

func eraOf(m domain.Match) string {
	if m.Classic != nil && m.Classic.Era != "" {
		return m.Classic.Era
	}
	t := startOf(m)
	return fmt.Sprintf("%d0s", t.Year()/10)
}

// xiPlayers is the set of player ids in either XI.
func xiPlayers(m domain.Match) map[string]bool {
	out := map[string]bool{}
	for _, ids := range m.XI {
		for _, id := range ids {
			out[id] = true
		}
	}
	return out
}
