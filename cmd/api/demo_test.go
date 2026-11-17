package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/auth"
	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/httpapi"
	"github.com/swathivallabhaneni289/innings/internal/sim"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// nullable are the fields the JSON contract lets be null, as in scripts/smoke.py. A null anywhere else, above all
// a list, breaks the app.
var nullable = map[string]bool{"winnerId": true, "officialPotmId": true, "fanPotmId": true, "logId": true,
	"matchId": true, "yourCall": true, "yourCallSplit": true, "yourCallMine": true, "topReview": true,
	"onThisDay": true, "avgRatingGiven": true, "accuracy": true, "prediction": true, "outcome": true, "log": true,
	"watching": true, "fans": true, "fansPotmId": true, "reason": true, "message": true}

// nullAt returns the key of a null in v that the contract doesn't allow, or "" if there's none.
func nullAt(v any, key string) string {
	switch v := v.(type) {
	case nil:
		if nullable[key] {
			return ""
		}
		return key
	case map[string]any:
		for k, x := range v {
			if p := nullAt(x, k); p != "" {
				return p
			}
		}
	case []any:
		for _, x := range v {
			if p := nullAt(x, key); p != "" {
				return p
			}
		}
	}
	return ""
}

// demoAPI seeds the demo account on a fresh database over the simulated world, as make demo does (with realMatches,
// the history it gets next to real matches), signs in with the password it printed, and returns the world and a way
// to GET as the demo fan.
func demoAPI(t *testing.T, realMatches bool) (*sim.World, func(path string) map[string]any) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "demo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	anchor := time.Date(2026, 10, 22, 18, 0, 0, 0, time.UTC)
	world := sim.NewWorld(anchor)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := httpapi.New(st, world, func() time.Time { return anchor }, logger)
	api.Auth.Mailer = auth.DevMailer{Log: logger}
	api.Auth.Iterations = 1000

	_, password, err := seedDemo(context.Background(), api.Auth, st, world, anchor, realMatches)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := seedDemo(context.Background(), api.Auth, st, world, anchor, realMatches); err == nil {
		t.Error("a second demo account on the same database: want an error")
	}

	h := api.Handler()
	do := func(method, path, token string, body any) (int, map[string]any) {
		t.Helper()
		var rd io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		}
		r := httptest.NewRequest(method, path, rd)
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		var out map[string]any
		json.Unmarshal(rec.Body.Bytes(), &out)
		if k := nullAt(out, ""); k != "" {
			t.Errorf("%s %s: %q is null", method, path, k)
		}
		return rec.Code, out
	}
	status, in := do("POST", "/v1/auth/signin", "", map[string]string{"email": demoEmail, "password": password})
	if status != http.StatusOK {
		t.Fatalf("sign in with the printed password: %d %v", status, in)
	}
	if acct, _ := in["account"].(map[string]any); acct["onboarded"] != true {
		t.Errorf("account %v: want onboarded, so the app goes straight to Home", acct)
	}
	token, _ := in["token"].(string)
	return world, func(path string) map[string]any {
		t.Helper()
		status, out := do("GET", path, token, nil)
		if status != http.StatusOK {
			t.Fatalf("GET %s: %d %v", path, status, out)
		}
		return out
	}
}

// checkDemoFan checks what the demo fan has either way: five stubs whose matches all resolve, the Top 5, and five
// friends. It returns the stubs' matches and the fan's stats.
func checkDemoFan(t *testing.T, world *sim.World, get func(string) map[string]any) ([]map[string]any, map[string]any) {
	t.Helper()
	me := get("/v1/users/me")
	user, _ := me["user"].(map[string]any)
	logs, _ := me["logs"].([]any)
	top5, _ := me["top5"].([]any)
	if user["handle"] != "demofan" || len(logs) != 5 {
		t.Errorf("profile: @%v with %d stubs, want @demofan with 5", user["handle"], len(logs))
	}
	matches := []map[string]any{}
	for _, l := range logs {
		m, _ := l.(map[string]any)["match"].(map[string]any)
		if m == nil || m["id"] == nil {
			t.Errorf("a stub without its match: %v", l)
			continue
		}
		matches = append(matches, m)
	}
	got := []string{}
	for _, m := range top5 {
		got = append(got, m.(map[string]any)["id"].(string))
	}
	if len(got) != 5 || !slices.Equal(got, world.DemoTop5()) {
		t.Errorf("Top 5 %v, want %v", got, world.DemoTop5())
	}
	following := 0
	for _, f := range get("/v1/friends")["friends"].([]any) {
		if f.(map[string]any)["following"] == true {
			following++
		}
	}
	if following != 5 {
		t.Errorf("following %d friends, want 5", following)
	}
	stats, _ := me["stats"].(map[string]any)
	return matches, stats
}

// On simulated matches the demo fan opens with points from settled predictions, which no API call could have made,
// and is on the Bragging Board next to friends' stubs.
func TestSeedDemo(t *testing.T) {
	world, get := demoAPI(t, false)
	_, stats := checkDemoFan(t, world, get)
	if points, _ := stats["points"].(float64); points <= 0 {
		t.Errorf("points %v: want the settled predictions to have scored", stats["points"])
	}
	home := get("/v1/home")
	onBoard := false
	for _, row := range home["board"].([]any) {
		onBoard = onBoard || row.(map[string]any)["isMe"] == true
	}
	if feed, _ := home["feed"].([]any); !onBoard || len(feed) == 0 {
		t.Errorf("home: demo fan on the Bragging Board %v, %d friends' stubs; want on it, and some stubs", onBoard, len(feed))
	}
}

// On real matches the simulated fixtures don't exist, so the stubs are two Vault classics and three archive finals,
// and there are no predictions until demoPicks adds them.
func TestSeedDemoRealMatches(t *testing.T) {
	world, get := demoAPI(t, true)
	matches, stats := checkDemoFan(t, world, get)
	kinds := map[string]int{}
	for _, m := range matches {
		switch {
		case m["kind"] == "classic":
			kinds["classic"]++
		case m["archived"] == true && strings.HasPrefix(m["id"].(string), "ar-"):
			kinds["archive"]++
		default:
			kinds["other"]++
		}
	}
	if kinds["classic"] != 2 || kinds["archive"] != 3 || kinds["other"] != 0 {
		t.Errorf("stubs %v, want 2 classics and 3 archive matches", kinds)
	}
	if points, _ := stats["points"].(float64); points != 0 {
		t.Errorf("points %v, want none before demoPicks", stats["points"])
	}
}

func TestDemoPicks(t *testing.T) {
	now := time.Date(2026, 11, 17, 9, 0, 0, 0, time.UTC)
	match := func(id string, daysAgo int, winner string, state domain.MatchState, xi bool) domain.Match {
		m := domain.Match{ID: id, TeamIDs: [2]string{"aa", "bb"}, State: state,
			StartTime: now.Add(-time.Duration(daysAgo) * 24 * time.Hour).Format(time.RFC3339)}
		if state == domain.Completed {
			m.Result = &domain.MatchResult{TopRunScorerIDs: []string{"aa-3"}, TopWicketTakerIDs: []string{"bb-7"}}
			if winner != "" {
				m.Result.WinnerID = &winner
			}
		}
		if xi {
			m.XI = map[string][]string{"aa": {"aa-1", "aa-2", "aa-3", "aa-11"}, "bb": {"bb-1", "bb-7", "bb-11"}}
		}
		return m
	}
	fixtures := []domain.Match{
		match("older", 5, "aa", domain.Completed, true), // a fifth: one too many
		match("m3", 3, "aa", domain.Completed, true),
		match("m1", 1, "aa", domain.Completed, true),
		match("last-week", 8, "aa", domain.Completed, true),
		match("no-winner", 1, "", domain.Completed, true),
		match("no-xi", 1, "aa", domain.Completed, false),
		match("odd-winner", 1, "cc", domain.Completed, true),
		match("live", 0, "", domain.Live, true),
		match("m2", 2, "bb", domain.Completed, true),
		match("m4", 4, "bb", domain.Completed, true),
	}
	picks := demoPicks("u-demo", fixtures, now)
	type pick struct{ match, winner, bat, bowl string }
	got := []pick{}
	for _, p := range picks {
		got = append(got, pick{p.MatchID, p.WinnerID, p.TopRunScorerID, p.TopWicketTakerID})
		start, _ := time.Parse(time.RFC3339, fixtures[slices.IndexFunc(fixtures, func(m domain.Match) bool { return m.ID == p.MatchID })].StartTime)
		if created, err := time.Parse(time.RFC3339, p.CreatedAt); err != nil || !created.Before(start) || p.UserID != "u-demo" {
			t.Errorf("%s: made at %q by %q, want before the first ball (%s), by u-demo", p.MatchID, p.CreatedAt, p.UserID, start)
		}
	}
	want := []pick{
		{"m1", "aa", "aa-3", "bb-7"},  // the newest: all three right
		{"m2", "bb", "bb-1", "aa-11"}, // the winner right
		{"m3", "bb", "bb-1", "aa-11"}, // the one wrong call: aa won
		{"m4", "bb", "bb-1", "aa-11"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("picks\n got %v\nwant %v", got, want)
	}
	if picks := demoPicks("u-demo", nil, now); len(picks) != 0 {
		t.Errorf("no matches: got %v", picks)
	}
}
