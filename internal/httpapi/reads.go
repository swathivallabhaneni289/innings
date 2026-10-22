package httpapi

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// collections mirrors COLLECTIONS in src/domain/constants.ts (order, titles, blurbs).
var collections = []domain.Collection{
	{ID: "world-cup-finals", Title: "World Cup Finals", Blurb: "Every crown, every heartbreak."},
	{ID: "greatest-chases", Title: "Greatest Chases", Blurb: "Targets that should have been too many."},
	{ID: "test-thrillers", Title: "Test Match Thrillers", Blurb: "Five days, decided in a session."},
	{ID: "super-over-madness", Title: "Super Over Madness", Blurb: "Tied after the full distance."},
	{ID: "womens-classics", Title: "Women's Cricket Classics", Blurb: "The games that changed the women’s game."},
}

// competitionNames mirrors COMPETITIONS in src/domain/constants.ts (used by search).
var competitionNames = map[domain.CompetitionID][]string{
	"international": {"International", "INTL"},
	"ipl":           {"Indian Premier League", "IPL"},
	"wpl":           {"Women's Premier League", "WPL"},
	"world-cup":     {"World Cup"},
	"bbl":           {"Big Bash League", "BBL"},
	"sa20":          {"SA20"},
	"the-hundred":   {"The Hundred"},
}

func byStart(a, b domain.Match) int     { return startOf(a).Compare(startOf(b)) }
func byStartDesc(a, b domain.Match) int { return startOf(b).Compare(startOf(a)) }

// ---- home ----

// profile ranks how big a fixture is for "Your Call": India > World Cup > other internationals > leagues.
func (v *view) profile(m domain.Match) int {
	p := 0
	for _, id := range m.TeamIDs {
		if id == "ind" || id == "ind-w" {
			p += 4
		}
	}
	if m.CompetitionID == "world-cup" {
		p += 3
	}
	if t, ok := v.teams()[m.TeamIDs[0]]; (ok && t.Kind == "international") || m.CompetitionID == "international" || m.CompetitionID == "world-cup" {
		p += 2
	}
	if m.Gender == domain.Men {
		p++
	}
	return p
}

func (v *view) yourCall() *domain.Match {
	var open []domain.Match
	for _, m := range v.fixtures() {
		if m.State == domain.Upcoming && !domain.IsLocked(m, v.now) {
			open = append(open, m)
		}
	}
	// Prefer the next three days so the card isn't a week away.
	var soon []domain.Match
	for _, m := range open {
		if startOf(m).Before(v.now.Add(72 * time.Hour)) {
			soon = append(soon, m)
		}
	}
	if len(soon) > 0 {
		open = soon
	}
	var best *domain.Match
	for i := range open {
		m := open[i]
		if best == nil || v.profile(m) > v.profile(*best) || (v.profile(m) == v.profile(*best) && startOf(m).Before(startOf(*best))) {
			best = &m
		}
	}
	return best
}

func (s *Server) handleHome(v *view, w http.ResponseWriter, r *http.Request) {
	onNow := []domain.Match{}
	for _, m := range v.fixtures() {
		if m.State == domain.Live {
			onNow = append(onNow, m)
		}
	}
	slices.SortStableFunc(onNow, byStart)
	y, mo, d := v.now.Local().Date()
	var today []domain.Match
	for _, m := range v.fixtures() {
		if m.State == domain.Live {
			continue
		}
		if ty, tmo, td := startOf(m).Local().Date(); ty == y && tmo == mo && td == d {
			today = append(today, m)
		}
	}
	slices.SortStableFunc(today, byStart)
	onNow = append(onNow, today...)

	resp := domain.HomeResponse{OnNow: onNow, Feed: []domain.FeedItem{}, JustFinished: []domain.JustFinished{}, Board: []domain.BoardRow{}}
	if yc := v.yourCall(); yc != nil {
		sp := v.split(*yc)
		resp.YourCall, resp.YourCallSplit, resp.YourCallMine = yc, &sp, v.myPrediction(yc.ID)
	}

	circle := toSet(v.circle(v.me.ID))
	var feedLogs []domain.MatchLog
	for _, l := range v.logs() {
		if l.UserID != v.me.ID && circle[l.UserID] {
			feedLogs = append(feedLogs, l)
		}
	}
	resp.Feed = v.feedItems(feedLogs, 40)

	var finished []domain.Match
	for _, m := range v.fixtures() {
		if m.State == domain.Completed && !v.endOf(m).Before(v.now.Add(-24*time.Hour)) {
			finished = append(finished, m)
		}
	}
	slices.SortStableFunc(finished, func(a, b domain.Match) int { return v.endOf(b).Compare(v.endOf(a)) })
	for _, m := range finished {
		rm := v.rated(m)
		jf := domain.JustFinished{Match: m, RatingAvg: rm.RatingAvg, RatingCount: rm.RatingCount}
		if reviews := v.reviewItems(m); len(reviews) > 0 {
			jf.TopReview = &reviews[0]
		}
		resp.JustFinished = append(resp.JustFinished, jf)
	}

	resp.Board = v.board()
	since := v.now.Add(-24 * time.Hour)
	for _, n := range v.notifications() {
		if !parseTime(n.CreatedAt).Before(since) {
			resp.Unread++
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// board: my circle's points from matches that finished in the last 7 days, best first.
func (v *view) board() []domain.BoardRow {
	rows := []domain.BoardRow{}
	since := v.now.Add(-7 * 24 * time.Hour)
	for _, id := range v.circle(v.me.ID) {
		u, ok := v.user(id)
		if !ok {
			continue
		}
		rows = append(rows, domain.BoardRow{User: u, Points: v.pointsSince(id, since), IsMe: id == v.me.ID})
	}
	slices.SortStableFunc(rows, func(a, b domain.BoardRow) int {
		if c := cmp.Compare(b.Points, a.Points); c != 0 {
			return c
		}
		return cmp.Compare(a.User.Name, b.User.Name)
	})
	return rows
}

// ---- matches ----

func (s *Server) handleMatches(v *view, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	state, comp := q.Get("state"), q.Get("competition")
	var teams map[string]bool
	if t := strings.TrimSpace(q.Get("team")); t != "" {
		teams = toSet(strings.Split(strings.ToLower(t), ","))
	}
	switch state {
	case "", "live", "upcoming", "recent":
	default:
		writeErr(w, http.StatusBadRequest, "bad-state", "state must be live, upcoming or recent")
		return
	}
	out := []domain.Match{}
	for _, m := range v.fixtures() {
		switch state {
		case "live":
			if m.State != domain.Live {
				continue
			}
		case "upcoming":
			if m.State != domain.Upcoming {
				continue
			}
		case "recent":
			if m.State != domain.Completed {
				continue
			}
		}
		if comp != "" && string(m.CompetitionID) != comp {
			continue
		}
		if teams != nil && !teams[m.TeamIDs[0]] && !teams[m.TeamIDs[1]] {
			continue
		}
		out = append(out, m)
	}
	if state == "recent" {
		slices.SortStableFunc(out, byStartDesc)
	} else {
		slices.SortStableFunc(out, byStart)
	}
	ratings := map[string]domain.FanRating{}
	for _, m := range out {
		if m.State == domain.Completed {
			ratings[m.ID] = v.fanRating(m)
		}
	}
	writeJSON(w, http.StatusOK, domain.MatchesResponse{Matches: out, FanRatings: ratings})
}

func (s *Server) handleMatch(v *view, w http.ResponseWriter, r *http.Request) {
	m, ok := v.match(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such match")
		return
	}
	resp := domain.MatchPageResponse{
		Match:           m,
		Locked:          domain.IsLocked(m, v.now),
		Split:           v.split(m),
		FriendsWatching: []domain.User{},
		Results:         []domain.PredictionResultRow{},
	}
	resp.Mine.Prediction = v.myPrediction(m.ID)
	if resp.Mine.Prediction != nil && m.State == domain.Completed {
		resp.Mine.Outcome = v.outcomeFor(v.me.ID, m.ID)
	}
	resp.Mine.Log = v.myLog(m.ID)
	watching := must(s.Store.Watching(v.ctx, m.ID))
	if p, ok := watching[v.me.ID]; ok {
		resp.Mine.Watching = &p
	}
	following := toSet(v.followsOf(v.me.ID))
	seen := map[string]bool{}
	addWatcher := func(id string) {
		if seen[id] || !following[id] {
			return
		}
		if u, ok := v.user(id); ok {
			seen[id] = true
			resp.FriendsWatching = append(resp.FriendsWatching, u)
		}
	}
	for _, id := range s.World.FriendsWatching(m.ID, v.now) {
		addWatcher(id)
	}
	for _, id := range v.followsOf(v.me.ID) { // follow order keeps the output stable
		if _, ok := watching[id]; ok {
			addWatcher(id)
		}
	}
	if m.State == domain.Completed {
		fs := v.fanSummary(m)
		resp.Fans = &fs
		v.loadPreds()
		circle := toSet(v.circle(v.me.ID))
		for _, p := range v.predsByMatch[m.ID] {
			if !circle[p.UserID] {
				continue
			}
			u, ok := v.user(p.UserID)
			o := v.outcomeFor(p.UserID, m.ID)
			if !ok || o == nil {
				continue
			}
			resp.Results = append(resp.Results, domain.PredictionResultRow{User: u, Prediction: p, Outcome: *o})
		}
		slices.SortStableFunc(resp.Results, func(a, b domain.PredictionResultRow) int {
			if c := cmp.Compare(b.Outcome.Points, a.Outcome.Points); c != 0 {
				return c
			}
			return cmp.Compare(a.User.Name, b.User.Name)
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---- vault ----

func (v *view) top5(userID string) []domain.Match {
	out := []domain.Match{}
	if !v.isReal(userID) {
		return out
	}
	for _, id := range must(v.s.Store.Top5(v.ctx, userID)) {
		if m, ok := v.match(id); ok {
			out = append(out, m)
		}
	}
	return out
}

func sortRated(ms []domain.RatedMatch) {
	slices.SortStableFunc(ms, func(a, b domain.RatedMatch) int {
		if c := cmp.Compare(b.RatingAvg, a.RatingAvg); c != 0 {
			return c
		}
		return cmp.Compare(b.RatingCount, a.RatingCount)
	})
}

func (s *Server) handleVault(v *view, w http.ResponseWriter, r *http.Request) {
	resp := domain.VaultResponse{Collections: []domain.CollectionShelf{}, Top5: v.top5(v.me.ID), FanRatings: map[string]domain.FanRating{}, Eras: []string{}}
	for _, m := range resp.Top5 {
		resp.FanRatings[m.ID] = v.fanRating(m)
	}
	for _, c := range collections {
		shelf := domain.CollectionShelf{Collection: c, Matches: []domain.RatedMatch{}}
		for _, m := range v.classics() {
			if m.Classic != nil && slices.Contains(m.Classic.Collections, c.ID) {
				shelf.Matches = append(shelf.Matches, v.rated(m))
			}
		}
		sortRated(shelf.Matches)
		resp.Collections = append(resp.Collections, shelf)
	}
	eras := map[string]bool{}
	today := v.now.Local()
	todayDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	bestDays := -1
	for _, m := range v.classics() {
		eras[eraOf(m)] = true
		if m.Classic == nil {
			continue
		}
		played, err := time.Parse("2006-01-02", m.Classic.Date)
		if err != nil {
			continue
		}
		next := time.Date(todayDay.Year(), played.Month(), played.Day(), 0, 0, 0, 0, time.UTC)
		if next.Before(todayDay) {
			next = time.Date(todayDay.Year()+1, played.Month(), played.Day(), 0, 0, 0, 0, time.UTC)
		}
		days := int(next.Sub(todayDay).Hours() / 24)
		if bestDays < 0 || days < bestDays {
			bestDays = days
			resp.OnThisDay = &domain.OnThisDay{Match: m, YearsAgo: next.Year() - played.Year(), Exact: days == 0}
		}
	}
	for e := range eras {
		resp.Eras = append(resp.Eras, e)
	}
	slices.Sort(resp.Eras)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRankings(v *view, w http.ResponseWriter, r *http.Request) {
	format, era := r.URL.Query().Get("format"), r.URL.Query().Get("era")
	if format == "all" {
		format = ""
	}
	if era == "all" {
		era = ""
	}
	out := []domain.RatedMatch{}
	consider := func(m domain.Match) {
		if format != "" && string(m.Format) != format {
			return
		}
		if era != "" && eraOf(m) != era {
			return
		}
		out = append(out, v.rated(m))
	}
	for _, m := range v.classics() {
		consider(m)
	}
	for _, m := range v.fixtures() {
		if m.State == domain.Completed {
			consider(m)
		}
	}
	sortRated(out)
	writeJSON(w, http.StatusOK, domain.RankingsResponse{Matches: out})
}

// ---- users ----

func (v *view) stats(userID string) domain.FanStats {
	st := domain.FanStats{Badges: []domain.BadgeDef{}, Watch: []domain.WatchCount{}, Stadiums: []domain.Venue{}}
	logs := v.userLogs(userID)
	st.Logged = len(logs)
	sum := 0.0
	watch := map[domain.WatchPlace]int{}
	stadiumLogs := 0
	seenVenue := map[string]bool{}
	for _, l := range logs {
		sum += l.Rating
		watch[l.WatchedAt]++
		if l.WatchedAt == domain.WatchStadium {
			stadiumLogs++
			if m, ok := v.match(l.MatchID); ok && !seenVenue[m.Venue.ID] {
				seenVenue[m.Venue.ID] = true
				st.Stadiums = append(st.Stadiums, m.Venue)
			}
		}
	}
	if len(logs) > 0 {
		avg := domain.Round2(sum / float64(len(logs)))
		st.AvgRatingGiven = &avg
	}
	st.Watch = domain.WatchCounts(watch)
	hs := domain.SummariseHistory(v.outcomes(userID), v.ties(userID))
	st.Points, st.Streak, st.Accuracy = hs.Points, hs.Streak, hs.Accuracy
	st.Badges = domain.Badges(hs, stadiumLogs)
	circle := v.circle(userID)
	st.Rank, st.RankOf = 1, len(circle)
	for _, id := range circle {
		if id != userID && domain.SummariseHistory(v.outcomes(id), nil).Points > st.Points {
			st.Rank++
		}
	}
	return st
}

func (s *Server) handleUser(v *view, w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "me" {
		id = v.me.ID
	}
	u, ok := v.user(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such user")
		return
	}
	writeJSON(w, http.StatusOK, domain.UserResponse{
		User: u, IsMe: id == v.me.ID, Following: id != v.me.ID && v.iFollow(id),
		Stats: v.stats(id), Logs: v.feedItems(v.userLogs(id), 0), Top5: v.top5(id),
	})
}

func (s *Server) handleFriends(v *view, w http.ResponseWriter, r *http.Request) {
	rows := []domain.FriendRow{}
	seen := map[string]bool{}
	for _, u := range s.World.Friends() {
		seen[u.ID] = true
		rows = append(rows, domain.FriendRow{User: u, Following: v.iFollow(u.ID)})
	}
	for _, id := range v.followsOf(v.me.ID) { // real users (and public fans) I follow
		if u, ok := v.user(id); ok && !seen[id] {
			seen[id] = true
			rows = append(rows, domain.FriendRow{User: u, Following: true})
		}
	}
	writeJSON(w, http.StatusOK, domain.FriendsResponse{Friends: rows})
}

func (s *Server) handleLog(v *view, w http.ResponseWriter, r *http.Request) {
	l, ok := v.logByID(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such log")
		return
	}
	fi, ok := v.feedItem(l)
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "log refers to an unknown match or user")
		return
	}
	writeJSON(w, http.StatusOK, fi)
}

// ---- search ----

func (v *view) haystack(m domain.Match) string {
	parts := []string{m.Stage, string(m.CompetitionID), string(m.Format), m.Venue.Name, m.Venue.City, m.Venue.Country}
	parts = append(parts, competitionNames[m.CompetitionID]...)
	for _, id := range m.TeamIDs {
		parts = append(parts, id)
		if t, ok := v.teams()[id]; ok {
			parts = append(parts, t.Name, t.Short)
		}
	}
	if m.Classic != nil {
		parts = append(parts, m.Classic.CompetitionName, m.Classic.Date, m.Classic.Headline)
	}
	if t := startOf(m); !t.IsZero() {
		parts = append(parts, strconv.Itoa(t.Year()))
	}
	return strings.ToLower(strings.Join(parts, " | "))
}

func (s *Server) handleSearch(v *view, w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	resp := domain.SearchResponse{Matches: []domain.Match{}, Users: []domain.User{}}
	if q == "" {
		writeJSON(w, http.StatusOK, resp)
		return
	}
	tokens := strings.Fields(q)
	hit := func(h string) bool {
		for _, t := range tokens {
			if !strings.Contains(h, t) {
				return false
			}
		}
		return true
	}
	var fx []domain.Match
	for _, m := range v.fixtures() {
		if hit(v.haystack(m)) {
			fx = append(fx, m)
		}
	}
	slices.SortStableFunc(fx, func(a, b domain.Match) int { // live, then upcoming soonest, then recent newest
		rank := func(m domain.Match) int {
			return map[domain.MatchState]int{domain.Live: 0, domain.Upcoming: 1, domain.Completed: 2}[m.State]
		}
		if c := cmp.Compare(rank(a), rank(b)); c != 0 {
			return c
		}
		if a.State == domain.Completed {
			return byStartDesc(a, b)
		}
		return byStart(a, b)
	})
	resp.Matches = append(resp.Matches, fx...)
	for _, m := range v.classics() {
		if hit(v.haystack(m)) {
			resp.Matches = append(resp.Matches, m)
		}
	}
	if len(resp.Matches) > 50 {
		resp.Matches = resp.Matches[:50]
	}
	v.loadUsers()
	users := append(append(slices.Clone(s.World.Friends()), s.World.PublicFans()...), v.realOrder...)
	for _, u := range users {
		if hit(strings.ToLower(u.Name + " " + u.Handle)) {
			resp.Users = append(resp.Users, u)
		}
		if len(resp.Users) >= 20 {
			break
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleLogSuggestions(v *view, w http.ResponseWriter, r *http.Request) {
	resp := domain.LogSuggestionsResponse{Suggested: []domain.Match{}, Recent: []domain.Match{}}
	var live, done []domain.Match
	since := v.now.Add(-24 * time.Hour)
	for _, m := range v.fixtures() {
		switch {
		case m.State == domain.Live:
			live = append(live, m)
		case m.State == domain.Completed:
			done = append(done, m)
		}
	}
	slices.SortStableFunc(live, byStart)
	slices.SortStableFunc(done, byStartDesc)
	resp.Suggested = append(resp.Suggested, live...)
	for _, m := range done {
		if !v.endOf(m).Before(since) {
			resp.Suggested = append(resp.Suggested, m)
		} else if len(resp.Recent) < 50 {
			resp.Recent = append(resp.Recent, m)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---- notifications ----

func (v *view) notifications() []domain.Notification {
	out := []domain.Notification{}
	add := func(id, text string, matchID, logID *string, at time.Time) {
		if at.After(v.now) {
			at = v.now
		}
		out = append(out, domain.Notification{ID: id, Text: text, MatchID: matchID, LogID: logID, CreatedAt: at.UTC().Format(time.RFC3339)})
	}
	// Predictions locking within 3 h that I haven't made.
	for _, m := range v.fixtures() {
		start := startOf(m)
		if domain.IsLocked(m, v.now) || start.After(v.now.Add(3*time.Hour)) || v.myPrediction(m.ID) != nil {
			continue
		}
		mins := int(start.Sub(v.now).Minutes())
		add("lock:"+m.ID, fmt.Sprintf("%s locks in %dh %02dm. Make your call.", v.title(m), mins/60, mins%60), domain.Ptr(m.ID), nil, start.Add(-3*time.Hour))
	}
	// Friends' new logs (last 3 days).
	following := toSet(v.followsOf(v.me.ID))
	since := v.now.Add(-72 * time.Hour)
	n := 0
	for _, l := range v.logs() {
		if n >= 20 {
			break
		}
		at := parseTime(l.CreatedAt)
		if !following[l.UserID] || at.Before(since) {
			continue
		}
		u, ok1 := v.user(l.UserID)
		m, ok2 := v.match(l.MatchID)
		if !ok1 || !ok2 {
			continue
		}
		n++
		add("log:"+l.ID, fmt.Sprintf("%s logged %s: %s balls", u.Name, v.title(m), strconv.FormatFloat(l.Rating, 'f', -1, 64)), domain.Ptr(m.ID), domain.Ptr(l.ID), at)
	}
	// My settled predictions with points (last 7 days).
	for _, o := range v.outcomes(v.me.ID) {
		m, ok := v.match(o.MatchID)
		if !ok || !o.Settled || o.Points <= 0 || v.endOf(m).Before(v.now.Add(-7*24*time.Hour)) {
			continue
		}
		add("pts:"+m.ID, fmt.Sprintf("You scored %d points on %s.", o.Points, v.title(m)), domain.Ptr(m.ID), nil, v.endOf(m))
	}
	// Replies and likes on my logs.
	mine := map[string]domain.MatchLog{}
	for _, l := range v.userLogs(v.me.ID) {
		mine[l.ID] = l
	}
	for logID, rs := range v.realReplies() {
		l, ok := mine[logID]
		if !ok {
			continue
		}
		for _, rp := range rs {
			if u, ok := v.user(rp.UserID); ok && rp.UserID != v.me.ID {
				add("reply:"+rp.ID, fmt.Sprintf("%s replied: %s", u.Name, rp.Text), domain.Ptr(l.MatchID), domain.Ptr(l.ID), parseTime(rp.CreatedAt))
			}
		}
	}
	for _, rx := range v.reactionRows("likes") {
		l, ok := mine[rx.LogID]
		if !ok || rx.UserID == v.me.ID {
			continue
		}
		if u, ok := v.user(rx.UserID); ok {
			m, _ := v.match(l.MatchID)
			add("like:"+rx.LogID+":"+rx.UserID, fmt.Sprintf("%s liked your %s stub.", u.Name, v.title(m)), domain.Ptr(l.MatchID), domain.Ptr(l.ID), parseTime(rx.CreatedAt))
		}
	}
	slices.SortStableFunc(out, func(a, b domain.Notification) int { return parseTime(b.CreatedAt).Compare(parseTime(a.CreatedAt)) })
	if len(out) > 50 {
		out = out[:50]
	}
	return out
}

func (s *Server) handleNotifications(v *view, w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, domain.NotificationsResponse{Items: v.notifications()})
}
