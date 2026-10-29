package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

func TestLiveCalls(t *testing.T) {
	me := newTestAPI(t, newFakeWorld(testNow))
	friend := me.another()

	got := calls(me, "live")
	if cv := got["live:total"]; cv.Call.State != domain.CallOpen || cv.Mine != nil || cv.Counts["170"] != 0 || len(cv.Call.Options) != 4 {
		t.Fatalf("open call: %+v", cv)
	}

	// Pick, then change the pick while it's open; the counts follow.
	if status, b := me.do("PUT", "/v1/matches/live/calls/live:total", map[string]string{"optionId": "170"}); status != http.StatusOK {
		t.Fatalf("pick: %d %s", status, b)
	}
	friend.do("PUT", "/v1/matches/live/calls/live:total", map[string]string{"optionId": "170"})
	me.do("PUT", "/v1/matches/live/calls/live:total", map[string]string{"optionId": "190plus"})
	got = calls(me, "live")
	if cv := got["live:total"]; cv.Mine == nil || *cv.Mine != "190plus" || cv.Counts["170"] != 1 || cv.Counts["190plus"] != 1 {
		t.Fatalf("after picks: %+v", cv)
	}

	// Locked and settled calls take no picks; bad options, unknown calls and another match's calls are refused.
	for _, tc := range []struct {
		path, option string
		want         int
	}{
		{"/v1/matches/live/calls/live:ms:1:ind-p0:50", "yes", http.StatusConflict},
		{"/v1/matches/done/calls/done:chase", "no", http.StatusConflict},
		{"/v1/matches/live/calls/live:total", "200plus", http.StatusBadRequest},
		{"/v1/matches/live/calls/nope", "yes", http.StatusNotFound},
		{"/v1/matches/live/calls/done:chase", "no", http.StatusNotFound},
	} {
		if status, _ := me.do("PUT", tc.path, map[string]string{"optionId": tc.option}); status != tc.want {
			t.Fatalf("%s %s: %d, want %d", tc.path, tc.option, status, tc.want)
		}
	}

	// A right call on a settled one is worth 5 points on the profile and the Bragging Board; a wrong one nothing.
	before := points(t, me)
	check(me.api.Store.SaveCallPick(t.Context(), domain.CallPick{UserID: me.id, CallID: "done:chase", MatchID: "done", OptionID: "no", CreatedAt: "2026-10-22T10:00:00Z"}))
	check(friend.api.Store.SaveCallPick(t.Context(), domain.CallPick{UserID: friend.id, CallID: "done:chase", MatchID: "done", OptionID: "yes", CreatedAt: "2026-10-22T10:00:00Z"}))
	if got := calls(me, "done")["done:chase"]; got.Points != domain.PointsLiveCall || got.Counts["yes"] != 1 {
		t.Fatalf("settled call: %+v", got)
	}
	if after := points(t, me); after != before+domain.PointsLiveCall {
		t.Fatalf("profile points %d → %d", before, after)
	}
	if got := calls(friend, "done")["done:chase"]; got.Points != 0 {
		t.Fatalf("wrong call scored: %+v", got)
	}
	home := me.getJSON("/v1/home", "board")
	found := false
	for _, row := range home["board"].([]any) {
		r := row.(map[string]any)
		if r["isMe"] == true && r["points"].(float64) >= float64(domain.PointsLiveCall) {
			found = true
		}
	}
	if !found {
		t.Fatalf("board: %v", home["board"])
	}
}

func points(t *testing.T, c *client) int {
	t.Helper()
	me := c.getJSON("/v1/users/me", "stats")
	return int(me["stats"].(map[string]any)["points"].(float64))
}

func TestFanCalls(t *testing.T) {
	me := newTestAPI(t, newFakeWorld(testNow))
	friend := me.another()
	post := func(c *client, path string, body map[string]any) (int, domain.CallView) {
		t.Helper()
		status, b := c.do("POST", path, body)
		var cv domain.CallView
		json.Unmarshal(b, &cv)
		return status, cv
	}
	call := map[string]any{"kind": "runs", "playerId": "ind-p0", "target": 50, "side": "yes"}
	status, cv := post(me, "/v1/matches/live/calls", call)
	if status != http.StatusCreated || cv.Call.CreatedBy != me.id || cv.Mine == nil || *cv.Mine != "yes" || cv.Counts["yes"] != 1 {
		t.Fatalf("post: %d %+v", status, cv)
	}
	// Asking the same thing joins the call: one card, both sides counted, one chat message.
	call["side"] = "no"
	if status, cv := post(friend, "/v1/matches/live/calls", call); status != http.StatusCreated || cv.Call.ID != "live:fan:runs:ind-p0:50" || cv.Counts["no"] != 1 || cv.Counts["yes"] != 1 {
		t.Fatalf("join: %d %+v", status, cv)
	}
	_, b := me.do("GET", "/v1/matches/live/chat", nil)
	var room domain.ChatResponse
	json.Unmarshal(b, &room)
	n := 0
	for _, m := range room.Messages {
		if m.CallID == "live:fan:runs:ind-p0:50" {
			n++
			if m.UserID != me.id || m.Text != "Made a call: Will ind-p0 make 50?" {
				t.Fatalf("call message: %+v", m)
			}
		}
	}
	if n != 1 {
		t.Fatalf("call messages = %d", n)
	}
	// It shows on the match page with the calls the app made.
	if got := calls(me, "live")["live:fan:runs:ind-p0:50"]; got.Call.CreatedBy != me.id {
		t.Fatalf("on the page: %+v", got)
	}

	for _, tc := range []struct {
		path string
		body map[string]any
		want int
	}{
		{"/v1/matches/live/calls", map[string]any{"kind": "runs", "playerId": "ind-p0", "target": 50, "side": "maybe"}, http.StatusBadRequest},
		{"/v1/matches/live/calls", map[string]any{"kind": "runs", "playerId": "nobody", "target": 50, "side": "yes"}, http.StatusBadRequest},
		{"/v1/matches/done/calls", call, http.StatusConflict},
		{"/v1/matches/nope/calls", call, http.StatusNotFound},
	} {
		if status, _ := me.do("POST", tc.path, tc.body); status != tc.want {
			t.Fatalf("%s %v: %d, want %d", tc.path, tc.body, status, tc.want)
		}
	}
}

// calls reads a match page's live calls by id.
func calls(c *client, matchID string) map[string]domain.CallView {
	t := c.t
	t.Helper()
	_, b := c.do("GET", "/v1/matches/"+matchID, nil)
	var page struct {
		Calls []domain.CallView `json:"calls"`
	}
	if err := json.Unmarshal(b, &page); err != nil || page.Calls == nil {
		t.Fatalf("calls on %s: %v %s", matchID, err, b)
	}
	out := map[string]domain.CallView{}
	for _, cv := range page.Calls {
		if cv.Counts == nil || cv.Call.Options == nil {
			t.Fatalf("null counts or options: %+v", cv)
		}
		out[cv.Call.ID] = cv
	}
	return out
}
