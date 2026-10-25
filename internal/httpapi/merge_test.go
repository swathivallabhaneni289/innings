package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// TestMergeRules checks the sim + real merge rules against the fake world, where every number is known.
func TestMergeRules(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	c.seedFakeHistory()

	page := c.getJSON("/v1/matches/done")
	fans := page["fans"].(map[string]any)
	if fans["ratingCount"] != 13.0 || fans["ratingAvg"] != 3.85 {
		t.Errorf("fan rating: %v/%v", fans["ratingAvg"], fans["ratingCount"])
	}
	if fans["fansPotmId"] != "mi-p1" || fans["officialPotmId"] != "csk-p0" || fans["fansDisagree"] != true {
		t.Errorf("potm: fans=%v official=%v disagree=%v", fans["fansPotmId"], fans["officialPotmId"], fans["fansDisagree"])
	}
	if rv := list(fans["reviews"]); len(rv) != 2 || str(rv[0], "log", "id") != "sl-1" || rv[0].(map[string]any)["likes"] != 12.0 {
		t.Errorf("reviews: %v", rv)
	}
	if fr := list(fans["friendReviews"]); len(fr) != 1 || str(fr[0], "log", "id") != "sl-1" {
		t.Errorf("friend reviews: %v", fr)
	}
	results := list(page["results"])
	if len(results) != 1 || str(results[0], "user", "id") != "f-aditi" {
		t.Fatalf("results: %v", results)
	}
	out := results[0].(map[string]any)["outcome"].(map[string]any)
	if out["upset"] != true || out["points"] != 50.0 {
		t.Errorf("aditi outcome: %v", out)
	}
	split := page["split"].(map[string]any)
	if split["total"] != 12.0 || split["shares"].(map[string]any)["csk"] != 0.25 || split["friendsTotal"] != 1.0 {
		t.Errorf("split: %v", split)
	}

	live := c.getJSON("/v1/matches/live")
	if fw := list(live["friendsWatching"]); len(fw) != 1 || str(fw[0], "id") != "f-aditi" {
		t.Errorf("friendsWatching: %v", fw)
	}
	if live["locked"] != true || live["fans"] != nil {
		t.Errorf("live page: locked=%v fans=%v", live["locked"], live["fans"])
	}

	home := c.getJSON("/v1/home")
	if str(home["yourCall"], "id") != "big" {
		t.Errorf("yourCall = %v", str(home["yourCall"], "id"))
	}
	if picks := list(home["yourCallSplit"].(map[string]any)["friendPicks"]); len(picks) != 1 {
		t.Errorf("yourCall friend picks: %v", picks)
	}
	if onNow := list(home["onNow"]); len(onNow) == 0 || str(onNow[0], "id") != "live" {
		t.Errorf("onNow: %v", onNow)
	}
	if feed := list(home["feed"]); len(feed) != 2 || str(feed[0], "log", "id") != "sl-1" || len(list(feed[0].(map[string]any)["replies"])) != 1 {
		t.Errorf("feed: %v", feed)
	}
	jf := list(home["justFinished"])
	if len(jf) != 1 || str(jf[0], "match", "id") != "done" || str(jf[0], "topReview", "log", "id") != "sl-1" {
		t.Errorf("justFinished: %v", jf)
	}
	board := list(home["board"])
	if len(board) != 2 || str(board[0], "user", "id") != "f-aditi" || board[0].(map[string]any)["points"] != 50.0 || board[1].(map[string]any)["points"] != 40.0 {
		t.Errorf("board: %v", board)
	}
	if home["unread"].(float64) < 1 {
		t.Errorf("unread = %v", home["unread"])
	}

	me := c.getJSON("/v1/users/me")
	st := me["stats"].(map[string]any)
	if st["logged"] != 2.0 || st["points"] != 40.0 || st["accuracy"] != 1.0 || st["rank"] != 2.0 || st["rankOf"] != 2.0 || len(list(st["stadiums"])) != 1 {
		t.Errorf("stats: %v", st)
	}
	badges := []string{}
	for _, b := range list(st["badges"]) {
		badges = append(badges, str(b, "id"))
	}
	if strings.Join(badges, ",") != "first-call,sharpshooter" {
		t.Errorf("badges: %v", badges)
	}
	if logs := list(me["logs"]); len(logs) != 2 || logs[0].(map[string]any)["log"].(map[string]any)["stubNo"] == logs[1].(map[string]any)["log"].(map[string]any)["stubNo"] {
		t.Errorf("my logs: %v", logs)
	}

	notes := list(c.getJSON("/v1/notifications")["items"])
	kinds := map[string]bool{}
	for _, n := range notes {
		kinds[strings.SplitN(str(n, "id"), ":", 2)[0]] = true
	}
	for _, k := range []string{"lock", "log", "pts"} {
		if !kinds[k] {
			t.Errorf("notifications missing %s: %v", k, notes)
		}
	}

	vault := c.getJSON("/v1/vault")
	otd := vault["onThisDay"].(map[string]any)
	if str(otd, "match", "id") != "wc-2011-final" || otd["exact"] != false {
		t.Errorf("onThisDay: %v", otd)
	}
	if m := list(vault["collections"].([]any)[0].(map[string]any)["matches"]); len(m) != 1 {
		t.Errorf("world cup finals shelf: %v", m)
	}
	if eras := list(vault["eras"]); len(eras) != 1 || eras[0] != "2010s" {
		t.Errorf("eras: %v", eras)
	}
	if r := list(c.getJSON("/v1/vault/rankings?format=ODI&era=2010s")["matches"]); len(r) != 1 {
		t.Errorf("rankings ODI 2010s: %v", r)
	}
	if r := list(c.getJSON("/v1/vault/rankings?era=2020s")["matches"]); len(r) != 2 {
		t.Errorf("rankings 2020s: %v", r)
	}

	if ms := list(c.getJSON("/v1/matches?team=ind")["matches"]); len(ms) != 3 {
		t.Errorf("team=ind: %d", len(ms))
	}
	recent := c.getJSON("/v1/matches?state=recent")
	if ms := list(recent["matches"]); len(ms) != 2 || str(ms[0], "id") != "done" {
		t.Errorf("recent: %v", ms)
	}
	if fr := recent["fanRatings"].(map[string]any); len(fr) != 2 {
		t.Errorf("fanRatings: %v", fr)
	}
	sug := c.getJSON("/v1/log/suggestions")
	if len(list(sug["suggested"])) != 2 || len(list(sug["recent"])) != 1 {
		t.Errorf("suggestions: %v", sug)
	}
	s := c.getJSON("/v1/search?q=aditi")
	if len(list(s["users"])) != 1 {
		t.Errorf("search users: %v", s["users"])
	}
	if ms := list(c.getJSON("/v1/search?q=world+cup+2011")["matches"]); len(ms) != 1 {
		t.Errorf("search classic: %v", ms)
	}

	// Real users' predictions join the split; the prediction must name players from the XIs.
	status, b := c.do("PUT", "/v1/matches/soon/prediction", map[string]string{"winnerId": "mi", "topRunScorerId": "ind-p0", "topWicketTakerId": "csk-p1"})
	if status != 400 {
		t.Errorf("player outside XI: %d %s", status, b)
	}
	c.do("PUT", "/v1/matches/soon/prediction", map[string]string{"winnerId": "mi", "topRunScorerId": "mi-p0", "topWicketTakerId": "csk-p1"})
	other := c.another()
	if sp := other.getJSON("/v1/matches/soon")["split"].(map[string]any); sp["total"] != 1.0 || sp["friendsTotal"] != 0.0 {
		t.Errorf("split after a real prediction: %v", sp)
	}
}

// seedFakeHistory gives c the history guest accounts used to start with (fakeWorld.DemoHistory), which the
// numbers above depend on. The follow and the logs go through the API. The prediction can't: it is on a match that
// has finished, and the API rightly refuses those (409 locked), so it is written to the store directly.
func (c *client) seedFakeHistory() {
	c.t.Helper()
	logs, preds, follows := newFakeWorld(testNow).DemoHistory(c.id, testNow)
	for _, id := range follows {
		if status, b := c.do("POST", "/v1/users/"+id+"/follow", nil); status != http.StatusOK {
			c.t.Fatalf("follow %s: %d %s", id, status, b)
		}
	}
	for _, l := range logs {
		body := map[string]any{"matchId": l.MatchID, "rating": l.Rating, "watchedAt": l.WatchedAt, "review": l.Review}
		if status, b := c.do("POST", "/v1/logs", body); status != http.StatusCreated {
			c.t.Fatalf("log %s: %d %s", l.MatchID, status, b)
		}
	}
	for _, p := range preds {
		p.UserID = c.id
		if err := c.api.Store.SavePrediction(context.Background(), p); err != nil {
			c.t.Fatal(err)
		}
	}
}
