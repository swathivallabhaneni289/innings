package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestMatchChat(t *testing.T) {
	me := newTestAPI(t, newFakeWorld(testNow))
	friend := me.another()

	chat := func(c *client, path string) (int, struct {
		Messages []struct{ ID, MatchID, UserID, Text, CreatedAt string }
		Users    map[string]struct{ Name string }
		Open     bool
		OpensAt  string
	}) {
		t.Helper()
		status, b := c.do("GET", path, nil)
		var out struct {
			Messages []struct{ ID, MatchID, UserID, Text, CreatedAt string }
			Users    map[string]struct{ Name string }
			Open     bool
			OpensAt  string
		}
		json.Unmarshal(b, &out)
		if status == http.StatusOK && (out.Messages == nil || out.Users == nil) {
			t.Fatalf("%s: null list or map: %s", path, b)
		}
		return status, out
	}

	// An empty room on a live match is open.
	if status, got := chat(me, "/v1/matches/live/chat"); status != 200 || !got.Open || len(got.Messages) != 0 {
		t.Fatalf("empty room: %d %+v", status, got)
	}

	if status, b := me.do("POST", "/v1/matches/live/chat", map[string]string{"text": "  What a start by India!  "}); status != http.StatusCreated || !strings.Contains(string(b), `"text":"What a start by India!"`) {
		t.Fatalf("post: %d %s", status, b)
	}
	if status, b := friend.do("POST", "/v1/matches/live/chat", map[string]string{"text": "Australia will come back"}); status != http.StatusCreated {
		t.Fatalf("friend post: %d %s", status, b)
	}

	// Both see both messages, oldest first, with their authors.
	_, got := chat(friend, "/v1/matches/live/chat")
	if len(got.Messages) != 2 || got.Messages[0].Text != "What a start by India!" || got.Messages[1].UserID != friend.id || len(got.Users) != 2 {
		t.Fatalf("room: %+v", got)
	}
	// ?after= returns from that message on (the app drops what it already has by id).
	if _, since := chat(me, "/v1/matches/live/chat?after="+got.Messages[1].CreatedAt); len(since.Messages) < 1 || since.Messages[len(since.Messages)-1].Text != "Australia will come back" {
		t.Fatalf("after: %+v", since)
	}
	// Rooms are per match.
	if _, other := chat(me, "/v1/matches/soon/chat"); len(other.Messages) != 0 || !other.Open {
		t.Fatalf("other room: %+v", other)
	}

	// One message a second per fan.
	status, _, h := me.req("POST", "/v1/matches/live/chat", map[string]string{"text": "again"})
	if status != http.StatusTooManyRequests || h.Get("Retry-After") != "1" {
		t.Fatalf("rate limit: %d %v", status, h)
	}
	for _, bad := range []string{"", "   ", strings.Repeat("x", 501)} {
		if status, _ := friend.do("POST", "/v1/matches/soon/chat", map[string]string{"text": bad}); status != http.StatusBadRequest {
			t.Fatalf("text %q: %d", bad, status)
		}
	}

	// A match long finished and a Vault classic have closed rooms; an unknown match has none.
	if status, got := chat(me, "/v1/matches/old/chat"); status != 200 || got.Open {
		t.Fatalf("old: %d %+v", status, got)
	}
	if status, _ := friend.do("POST", "/v1/matches/old/chat", map[string]string{"text": "late"}); status != http.StatusConflict {
		t.Fatalf("old post: %d", status)
	}
	if status, _ := friend.do("POST", "/v1/matches/wc-2011-final/chat", map[string]string{"text": "Dhoni!"}); status != http.StatusConflict {
		t.Fatalf("classic post: %d", status)
	}
	if status, _ := me.do("GET", "/v1/matches/nope/chat", nil); status != http.StatusNotFound {
		t.Fatalf("unknown: %d", status)
	}
	if status, _ := me.do("GET", "/v1/matches/live/chat?after=yesterday", nil); status != http.StatusBadRequest {
		t.Fatalf("bad after: %d", status)
	}

	// Deleting an account takes its messages with it.
	if status, _ := friend.do("DELETE", "/v1/me", map[string]string{"password": friend.password}); status != http.StatusNoContent && status != http.StatusOK {
		t.Fatalf("delete: %d", status)
	}
	if _, got := chat(me, "/v1/matches/live/chat"); len(got.Messages) != 1 {
		t.Fatalf("after delete: %+v", got)
	}
}
