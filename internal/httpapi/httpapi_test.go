package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/swathivallabhaneni289/innings/internal/sim"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// Fixed anchor + pinned clock so runs are deterministic whatever the sim generates.
var (
	testAnchor = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	testNow    = testAnchor.Add(90 * time.Minute)
)

// arrayKeys are JSON keys whose TS type is an array or a map: they must never be null.
var arrayKeys = map[string]bool{}

func init() {
	for _, k := range strings.Fields(`teams players onNow feed justFinished board matches fanRatings buckets reviews
		friendReviews shares friendShares friendPicks friendsWatching results collections top5 eras logs users friends
		suggested recent items badges watch stadiums playerRatings watchedWith replies innings topRunScorerIds
		topWicketTakerIds teamIds xi`) {
		arrayKeys[k] = true
	}
}

type client struct {
	t     *testing.T
	base  string
	token string
}

// eachWorld runs fn against the real sim (empty until it lands, full after) and against a small fake world.
func eachWorld(t *testing.T, fn func(t *testing.T, c *client)) {
	t.Run("sim", func(t *testing.T) { fn(t, newTestAPI(t, sim.NewWorld(testAnchor))) })
	t.Run("fake", func(t *testing.T) { fn(t, newTestAPI(t, newFakeWorld(testNow))) })
}

func newTestAPI(t *testing.T, world World) *client {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(New(st, world, func() time.Time { return testNow }, logger).Handler())
	t.Cleanup(srv.Close)
	c := &client{t: t, base: srv.URL}
	c.token = c.bootstrap("")
	return c
}

func (c *client) do(method, path string, body any) (int, []byte) {
	c.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.base+path, rd)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, b
}

// getJSON GETs path, expects 200, checks required keys and nulls, and returns the decoded object.
func (c *client) getJSON(path string, required ...string) map[string]any {
	c.t.Helper()
	status, b := c.do("GET", path, nil)
	if status != http.StatusOK {
		c.t.Fatalf("GET %s: status %d: %s", path, status, b)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		c.t.Fatalf("GET %s: %v", path, err)
	}
	for _, k := range required {
		if _, ok := obj[k]; !ok {
			c.t.Errorf("GET %s: missing key %q", path, k)
		}
	}
	walkNulls(c.t, path, "", obj)
	return obj
}

func walkNulls(t *testing.T, path, at string, v any) {
	t.Helper()
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if val == nil && arrayKeys[k] {
				t.Errorf("GET %s: %s.%s is null (must be [] or {})", path, at, k)
			}
			walkNulls(t, path, at+"."+k, val)
		}
	case []any:
		for _, val := range x {
			walkNulls(t, path, at+"[]", val)
		}
	}
}

func (c *client) bootstrap(token string) string {
	c.t.Helper()
	req, _ := http.NewRequest("POST", c.base+"/v1/bootstrap", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	var obj map[string]any
	json.NewDecoder(res.Body).Decode(&obj)
	if res.StatusCode != 200 {
		c.t.Fatalf("bootstrap: %d %v", res.StatusCode, obj)
	}
	for _, k := range []string{"token", "me", "teams", "players", "serverTime"} {
		if _, ok := obj[k]; !ok {
			c.t.Fatalf("bootstrap: missing %s", k)
		}
	}
	walkNulls(c.t, "/v1/bootstrap", "", obj)
	tok, _ := obj["token"].(string)
	if tok == "" {
		c.t.Fatal("bootstrap: empty token")
	}
	return tok
}

func str(v any, path ...string) string {
	for _, p := range path {
		m, _ := v.(map[string]any)
		v = m[p]
	}
	s, _ := v.(string)
	return s
}

func list(v any) []any { l, _ := v.([]any); return l }

func TestBootstrapResumes(t *testing.T) { eachWorld(t, testBootstrapResumes) }

func testBootstrapResumes(t *testing.T, c *client) {
	if again := c.bootstrap(c.token); again != c.token {
		t.Fatalf("resume returned a new token")
	}
	if other := c.bootstrap("not-a-real-token"); other == c.token {
		t.Fatalf("unknown token should create a new guest")
	}
}

func TestUnauthorized(t *testing.T) { eachWorld(t, testUnauthorized) }

func testUnauthorized(t *testing.T, c *client) {
	c.token = ""
	for _, p := range []string{"/v1/home", "/v1/matches", "/v1/users/me", "/v1/notifications"} {
		status, b := c.do("GET", p, nil)
		if status != http.StatusUnauthorized || !strings.Contains(string(b), `"error"`) {
			t.Errorf("%s without token: %d %s", p, status, b)
		}
	}
	c.token = "bogus"
	if status, _ := c.do("GET", "/v1/home", nil); status != http.StatusUnauthorized {
		t.Errorf("bogus token: %d", status)
	}
}

func TestCORSPreflight(t *testing.T) { eachWorld(t, testCORSPreflight) }

func testCORSPreflight(t *testing.T, c *client) {
	status, _ := c.do("OPTIONS", "/v1/home", nil)
	if status != http.StatusNoContent {
		t.Fatalf("preflight: %d", status)
	}
}

func TestEveryGET(t *testing.T) { eachWorld(t, testEveryGET) }

func testEveryGET(t *testing.T, c *client) {
	home := c.getJSON("/v1/home", "onNow", "yourCall", "yourCallSplit", "yourCallMine", "feed", "justFinished", "board", "unread")
	for _, q := range []string{"", "?state=live", "?state=upcoming", "?state=recent", "?competition=ipl", "?team=ind,ind-w"} {
		c.getJSON("/v1/matches"+q, "matches", "fanRatings")
	}
	if status, _ := c.do("GET", "/v1/matches?state=bogus", nil); status != http.StatusBadRequest {
		t.Errorf("bad state: %d", status)
	}
	matchKeys := []string{"match", "locked", "split", "mine", "friendsWatching", "fans", "results"}
	ids := map[string]bool{}
	for _, state := range []string{"live", "upcoming", "recent"} {
		if ms := list(c.getJSON("/v1/matches?state=" + state)["matches"]); len(ms) > 0 {
			ids[str(ms[0], "id")] = true
		}
	}
	vault := c.getJSON("/v1/vault", "collections", "onThisDay", "top5", "eras")
	if cols := list(vault["collections"]); len(cols) != 5 {
		t.Errorf("vault: want 5 collections, got %d", len(cols))
	}
	if ranks := list(c.getJSON("/v1/vault/rankings", "matches")["matches"]); len(ranks) > 0 {
		ids[str(ranks[0], "match", "id")] = true // a classic or completed fixture
	}
	c.getJSON("/v1/vault/rankings?format=T20&era=2020s", "matches")
	if len(ids) == 0 {
		t.Log("world has no matches yet; skipping match page checks")
	}
	for id := range ids {
		page := c.getJSON("/v1/matches/"+id, matchKeys...)
		mine, _ := page["mine"].(map[string]any)
		for _, k := range []string{"prediction", "outcome", "log", "watching"} {
			if _, ok := mine[k]; !ok {
				t.Errorf("match %s: mine.%s missing", id, k)
			}
		}
	}
	if status, _ := c.do("GET", "/v1/matches/no-such-match", nil); status != http.StatusNotFound {
		t.Errorf("unknown match: %d", status)
	}

	me := c.getJSON("/v1/users/me", "user", "isMe", "following", "stats", "logs", "top5")
	if me["isMe"] != true {
		t.Error("users/me: isMe false")
	}
	stats, _ := me["stats"].(map[string]any)
	for _, k := range []string{"logged", "avgRatingGiven", "accuracy", "points", "streak", "rank", "rankOf", "badges", "watch", "stadiums"} {
		if _, ok := stats[k]; !ok {
			t.Errorf("stats.%s missing", k)
		}
	}
	friends := list(c.getJSON("/v1/friends", "friends")["friends"])
	if len(friends) > 0 {
		c.getJSON("/v1/users/"+str(friends[0], "user", "id"), "user", "isMe", "following", "stats", "logs", "top5")
	}
	// A log to fetch: the feed, else my own stubs.
	logID := ""
	if feed := list(home["feed"]); len(feed) > 0 {
		logID = str(feed[0], "log", "id")
	} else if logs := list(me["logs"]); len(logs) > 0 {
		logID = str(logs[0], "log", "id")
	}
	if logID != "" {
		c.getJSON("/v1/logs/"+logID, "log", "user", "match", "likes", "liked", "thereToo", "iWasThere", "replies")
	} else {
		t.Log("no logs in the world yet; skipping /v1/logs/{id}")
	}
	c.getJSON("/v1/search?q=india", "matches", "users")
	c.getJSON("/v1/search?q=", "matches", "users")
	c.getJSON("/v1/log/suggestions", "suggested", "recent")
	c.getJSON("/v1/notifications", "items")
}

// findMatch returns the id of a fixture in state (live|recent|upcoming), or "".
func (c *client) findMatch(state string) string {
	ms := list(c.getJSON("/v1/matches?state=" + state)["matches"])
	if len(ms) == 0 {
		return ""
	}
	return str(ms[0], "id")
}

func TestPredictionLocked(t *testing.T) { eachWorld(t, testPredictionLocked) }

func testPredictionLocked(t *testing.T, c *client) {
	id := c.findMatch("live")
	if id == "" {
		id = c.findMatch("recent")
	}
	if id == "" {
		t.Skip("world has no live or completed fixture")
	}
	status, b := c.do("PUT", "/v1/matches/"+id+"/prediction", map[string]string{"winnerId": "x", "topRunScorerId": "y", "topWicketTakerId": "z"})
	if status != http.StatusConflict || !strings.Contains(string(b), `"locked"`) {
		t.Fatalf("want 409 locked, got %d %s", status, b)
	}
}

func TestPredictionUpcoming(t *testing.T) { eachWorld(t, testPredictionUpcoming) }

func testPredictionUpcoming(t *testing.T, c *client) {
	id := c.findMatch("upcoming")
	if id == "" {
		t.Skip("world has no upcoming fixture")
	}
	page := c.getJSON("/v1/matches/" + id)
	m := page["match"].(map[string]any)
	teams := list(m["teamIds"])
	xi, _ := m["xi"].(map[string]any)
	if len(xi) == 0 {
		t.Skip("upcoming fixture has no XI")
	}
	bat := str(list(xi[teams[0].(string)])[0])
	bowl := str(list(xi[teams[1].(string)])[10])
	if status, b := c.do("PUT", "/v1/matches/"+id+"/prediction", map[string]string{"winnerId": "nope", "topRunScorerId": bat, "topWicketTakerId": bowl}); status != 400 {
		t.Fatalf("bad winner: %d %s", status, b)
	}
	status, b := c.do("PUT", "/v1/matches/"+id+"/prediction", map[string]string{"winnerId": teams[0].(string), "topRunScorerId": bat, "topWicketTakerId": bowl})
	if status != 200 || !strings.Contains(string(b), `"prediction"`) {
		t.Fatalf("predict: %d %s", status, b)
	}
	page = c.getJSON("/v1/matches/" + id)
	if page["mine"].(map[string]any)["prediction"] == nil {
		t.Fatal("prediction not on match page")
	}
}

func TestLogCreateUpdateKeepsStub(t *testing.T) { eachWorld(t, testLogCreateUpdateKeepsStub) }

func testLogCreateUpdateKeepsStub(t *testing.T, c *client) {
	id := c.findMatch("recent")
	if id == "" {
		if ranks := list(c.getJSON("/v1/vault/rankings")["matches"]); len(ranks) > 0 {
			id = str(ranks[0], "match", "id")
		}
	}
	if id == "" {
		t.Skip("world has no completed match to log")
	}
	body := map[string]any{"matchId": id, "rating": 4.5, "instantClassic": true, "playerRatings": map[string]int{}, "fanPotmId": nil,
		"watchedAt": "hostel", "watchedWith": []string{}, "review": "What a game"}
	status, b := c.do("POST", "/v1/logs", body)
	if status != http.StatusCreated {
		t.Fatalf("create: %d %s", status, b)
	}
	var first struct {
		Log struct {
			ID     string  `json:"id"`
			StubNo int     `json:"stubNo"`
			Rating float64 `json:"rating"`
		} `json:"log"`
	}
	json.Unmarshal(b, &first)
	body["rating"] = 2.0
	status, b = c.do("POST", "/v1/logs", body)
	if status != http.StatusCreated {
		t.Fatalf("update: %d %s", status, b)
	}
	second := first
	json.Unmarshal(b, &second)
	if second.Log.StubNo != first.Log.StubNo || second.Log.ID != first.Log.ID || second.Log.Rating != 2 {
		t.Fatalf("update changed stub/id: %+v → %+v", first, second)
	}
	for _, bad := range []map[string]any{
		{"matchId": id, "rating": 4.2, "watchedAt": "home"},
		{"matchId": id, "rating": 0, "watchedAt": "home"},
		{"matchId": id, "rating": 3, "watchedAt": "moon"},
		{"matchId": id, "rating": 3, "watchedAt": "home", "playerRatings": map[string]int{"p": 11}},
		{"matchId": "nope", "rating": 3, "watchedAt": "home"},
	} {
		if status, b := c.do("POST", "/v1/logs", bad); status != http.StatusBadRequest {
			t.Errorf("bad log %v: %d %s", bad, status, b)
		}
	}
	// Like toggles on my own new log.
	logID := first.Log.ID
	_, b1 := c.do("POST", "/v1/logs/"+logID+"/like", nil)
	_, b2 := c.do("POST", "/v1/logs/"+logID+"/like", nil)
	if !strings.Contains(string(b1), `"liked":true`) || !strings.Contains(string(b1), `"likes":1`) || !strings.Contains(string(b2), `"likes":0`) {
		t.Fatalf("like toggle: %s / %s", b1, b2)
	}
	if status, b := c.do("POST", "/v1/logs/"+logID+"/replies", map[string]string{"text": "Agreed"}); status != http.StatusCreated {
		t.Fatalf("reply: %d %s", status, b)
	}
	fi := c.getJSON("/v1/logs/"+logID, "replies")
	if len(list(fi["replies"])) != 1 {
		t.Fatalf("reply not on log: %v", fi["replies"])
	}
	if status, _ := c.do("DELETE", "/v1/logs/"+logID, nil); status != http.StatusNoContent {
		t.Fatalf("delete: %d", status)
	}
	if status, _ := c.do("GET", "/v1/logs/"+logID, nil); status != http.StatusNotFound {
		t.Fatalf("deleted log still there: %d", status)
	}
}

func TestLikeToggleOnSimLog(t *testing.T) { eachWorld(t, testLikeToggleOnSimLog) }

func testLikeToggleOnSimLog(t *testing.T, c *client) {
	feed := list(c.getJSON("/v1/home")["feed"])
	if len(feed) == 0 {
		t.Skip("no friend logs in the world")
	}
	id := str(feed[0], "log", "id")
	base := feed[0].(map[string]any)["likes"].(float64)
	var r1, r2 struct {
		Liked bool    `json:"liked"`
		Likes float64 `json:"likes"`
	}
	_, b := c.do("POST", "/v1/logs/"+id+"/like", nil)
	json.Unmarshal(b, &r1)
	_, b = c.do("POST", "/v1/logs/"+id+"/like", nil)
	json.Unmarshal(b, &r2)
	if !r1.Liked || r1.Likes != base+1 || r2.Liked || r2.Likes != base {
		t.Fatalf("toggle: base %v, %+v, %+v", base, r1, r2)
	}
}

func TestFollowToggle(t *testing.T) { eachWorld(t, testFollowToggle) }

func testFollowToggle(t *testing.T, c *client) {
	other := &client{t: t, base: c.base}
	other.token = other.bootstrap("")
	otherID := str(other.getJSON("/v1/users/me")["user"], "id")

	_, b := c.do("POST", "/v1/users/"+otherID+"/follow", nil)
	if !strings.Contains(string(b), `"following":true`) {
		t.Fatalf("follow: %s", b)
	}
	if u := c.getJSON("/v1/users/" + otherID); u["following"] != true || u["isMe"] != false {
		t.Fatalf("user page: following=%v isMe=%v", u["following"], u["isMe"])
	}
	board := list(c.getJSON("/v1/home")["board"])
	found := false
	for _, row := range board {
		if str(row, "user", "id") == otherID {
			found = true
		}
	}
	if !found {
		t.Error("followed user missing from the board")
	}
	_, b = c.do("POST", "/v1/users/"+otherID+"/follow", nil)
	if !strings.Contains(string(b), `"following":false`) {
		t.Fatalf("unfollow: %s", b)
	}
	if status, _ := c.do("POST", "/v1/users/me/follow", nil); status != http.StatusBadRequest {
		t.Errorf("self-follow: %d", status)
	}
	if status, _ := c.do("POST", "/v1/users/ghost/follow", nil); status != http.StatusNotFound {
		t.Errorf("unknown user: %d", status)
	}
}

func TestTop5AndMe(t *testing.T) { eachWorld(t, testTop5AndMe) }

func testTop5AndMe(t *testing.T, c *client) {
	status, b := c.do("PUT", "/v1/me/top5", map[string]any{"matchIds": []string{"a", "b", "c", "d", "e", "f"}})
	if status != http.StatusBadRequest {
		t.Fatalf("6 matches: %d %s", status, b)
	}
	if status, _ := c.do("PUT", "/v1/me/top5", map[string]any{"matchIds": []string{"no-such-match"}}); status != http.StatusBadRequest {
		t.Fatalf("unknown match: %d", status)
	}
	status, b = c.do("PUT", "/v1/me/top5", map[string]any{"matchIds": []string{}})
	if status != 200 || !strings.Contains(string(b), `"top5":[]`) {
		t.Fatalf("empty top5: %d %s", status, b)
	}
	if ranks := list(c.getJSON("/v1/vault/rankings")["matches"]); len(ranks) > 0 {
		id := str(ranks[0], "match", "id")
		if status, b := c.do("PUT", "/v1/me/top5", map[string]any{"matchIds": []string{id}}); status != 200 {
			t.Fatalf("top5: %d %s", status, b)
		}
		if top := list(c.getJSON("/v1/vault")["top5"]); len(top) != 1 {
			t.Fatalf("vault top5: %v", top)
		}
	}
	status, b = c.do("PUT", "/v1/me", map[string]any{"name": "Swathi", "bio": "Dhoni finishes, I cry"})
	if status != 200 || !strings.Contains(string(b), `"name":"Swathi"`) {
		t.Fatalf("PUT /v1/me: %d %s", status, b)
	}
	if status, _ := c.do("PUT", "/v1/me", map[string]any{"name": " "}); status != 400 {
		t.Fatalf("blank name: %d", status)
	}
}

func TestWatching(t *testing.T) { eachWorld(t, testWatching) }

func testWatching(t *testing.T, c *client) {
	id := c.findMatch("upcoming")
	if id == "" {
		id = c.findMatch("live")
	}
	if id == "" {
		t.Skip("no live/upcoming fixture")
	}
	_, b := c.do("PUT", "/v1/matches/"+id+"/watching", map[string]any{"place": "cafe"})
	if !strings.Contains(string(b), `"watching":"cafe"`) {
		t.Fatalf("set: %s", b)
	}
	if w := c.getJSON("/v1/matches/" + id)["mine"].(map[string]any)["watching"]; w != "cafe" {
		t.Fatalf("page watching = %v", w)
	}
	_, b = c.do("PUT", "/v1/matches/"+id+"/watching", map[string]any{"place": nil})
	if !strings.Contains(string(b), `"watching":null`) {
		t.Fatalf("clear: %s", b)
	}
}
