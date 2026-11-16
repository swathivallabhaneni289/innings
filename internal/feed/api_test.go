package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/auth"
	"github.com/swathivallabhaneni289/innings/internal/httpapi"
	"github.com/swathivallabhaneni289/innings/internal/sim"
)

// The whole API over today's real matches (testdata): every read answers, predictions work against a squad, and nothing the
// simulation invented shows up next to a real match.
func TestAPIServesRealMatches(t *testing.T) {
	clock := now
	st := openStore(t)
	src := testSource(t, &espnFake{calls: map[string]int{}}, st, &clock)
	src.poll(context.Background())

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	simWorld := sim.NewWorld(now.Add(-72 * time.Hour))
	classic := map[string]bool{}
	for _, m := range simWorld.Classics() {
		classic[m.ID] = true
	}
	api := httpapi.New(st, NewWorld(simWorld, src), func() time.Time { return clock }, logger)
	api.Auth.Now = func() time.Time { return clock }
	api.Auth.Mailer = auth.DevMailer{Log: logger}
	api.Auth.Iterations = 1000
	srv := httptest.NewServer(api.Handler())
	t.Cleanup(srv.Close)

	token := ""
	call := func(method, path string, body any) (int, map[string]any) {
		t.Helper()
		var rd io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, srv.URL+path, rd)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		raw, _ := io.ReadAll(res.Body)
		if strings.Contains(string(raw), ":null") && !strings.Contains(path, "/v1/auth") {
			// Lists are never null (AGENTS.md); single objects may be.
			for _, k := range []string{`"innings":null`, `"xi":null`, `"feed":null`, `"onNow":null`, `"matches":null`, `"players":null`, `"reviews":null`} {
				if strings.Contains(string(raw), k) {
					t.Fatalf("%s %s: %s in %s", method, path, k, raw)
				}
			}
		}
		var out map[string]any
		json.Unmarshal(raw, &out)
		return res.StatusCode, out
	}
	_, sent := call("POST", "/v1/auth/signup", map[string]string{"email": "fan@example.com", "password": "tests-are-fun-42"})
	_, confirmed := call("POST", "/v1/auth/confirm", map[string]string{"email": "fan@example.com", "code": sent["devCode"].(string)})
	token = confirmed["token"].(string)

	status, boot := call("POST", "/v1/bootstrap", nil)
	if status != 200 || boot["teams"].(map[string]any)["oman"] == nil || boot["players"].(map[string]any)["esp-502714"] == nil {
		t.Fatalf("bootstrap %d: teams/players missing the feed's", status)
	}

	status, home := call("GET", "/v1/home", nil)
	if status != 200 {
		t.Fatalf("home %d %v", status, home)
	}
	ids := []string{}
	for _, m := range home["onNow"].([]any) {
		ids = append(ids, m.(map[string]any)["id"].(string))
	}
	for _, id := range ids {
		if !strings.HasPrefix(id, "es-") {
			t.Fatalf("a made-up fixture is on now: %v", ids)
		}
	}
	if len(ids) == 0 {
		t.Fatalf("nothing on now: %v", home)
	}
	for _, f := range home["feed"].([]any) {
		if mid := f.(map[string]any)["match"].(map[string]any)["id"].(string); !classic[mid] && !strings.HasPrefix(mid, "es-") {
			t.Fatalf("feed shows a made-up fixture %s", mid)
		}
	}

	for _, path := range []string{"/v1/matches", "/v1/matches?state=live", "/v1/matches?state=upcoming", "/v1/matches?state=recent", "/v1/matches/es-1539435", "/v1/matches/es-1550718", "/v1/matches/es-1543945", "/v1/matches/es-1525661", "/v1/vault", "/v1/notifications", "/v1/friends", "/v1/log/suggestions", "/v1/search?q=nep"} {
		if status, body := call("GET", path, nil); status != 200 {
			t.Fatalf("GET %s: %d %v", path, status, body)
		}
	}
	_, page := call("GET", "/v1/matches/es-1543945", nil)
	if fans := page["fans"].(map[string]any); fans["ratingCount"].(float64) != 0 {
		t.Fatalf("a real match has invented ratings: %v", fans)
	}

	// Predict tomorrow's New Zealand v India from the announced squads: India, Abhishek Sharma, Arshdeep Singh.
	pick := map[string]string{"winnerId": "ind", "topRunScorerId": "esp-1070183", "topWicketTakerId": "esp-1125976"}
	if status, body := call("PUT", "/v1/matches/es-1539435/prediction", pick); status != 200 {
		t.Fatalf("predict: %d %v", status, body)
	}
	// Someone outside the squads can't be picked; the live ODI is locked.
	if status, _ := call("PUT", "/v1/matches/es-1539435/prediction", map[string]string{"winnerId": "ind", "topRunScorerId": "esp-329336", "topWicketTakerId": "esp-1125976"}); status != 400 {
		t.Fatalf("outsider pick: %d", status)
	}
	if status, _ := call("PUT", "/v1/matches/es-1550718/prediction", map[string]string{"winnerId": "eng", "topRunScorerId": "esp-877051", "topWicketTakerId": "esp-784373"}); status != 409 {
		t.Fatalf("live match not locked: %d", status)
	}
	// Log the finished WBBL game.
	if status, body := call("POST", "/v1/logs", map[string]any{"matchId": "es-1543945", "rating": 4.5, "review": "What a chase.", "watchedAt": "home", "watchedWith": []string{}, "playerRatings": map[string]int{}}); status != 200 && status != 201 {
		t.Fatalf("log: %d %v", status, body)
	}
}
