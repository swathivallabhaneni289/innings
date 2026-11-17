package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/auth"
	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/httpapi"
	"github.com/swathivallabhaneni289/innings/internal/sim"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// demoEmail is the demo account's address (make demo, INNINGS_DEMO=1). example.com is reserved for examples, so no
// mail can reach anyone there, and dev mode never sends any.
const demoEmail = "demo@example.com"

// demoWords make up the demo password: three of them and two digits, e.g. "googly-crease-bails-47".
var demoWords = []string{"yorker", "googly", "doosra", "bouncer", "sweep", "flick", "gully", "maiden", "crease",
	"stumps", "bails", "spell", "glance", "cutter", "seamer", "boundary"}

// demoPassword makes up a new password for the demo account, from crypto/rand.
func demoPassword() (string, error) {
	pick := func(n int) (int, error) {
		k, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
		return int(k.Int64()), err
	}
	parts := []string{}
	for range 3 {
		k, err := pick(len(demoWords))
		if err != nil {
			return "", err
		}
		parts = append(parts, demoWords[k])
	}
	n, err := pick(90)
	if err != nil {
		return "", err
	}
	return strings.Join(parts, "-") + "-" + strconv.Itoa(10+n), nil
}

// seedDemo opens the demo account through the normal sign-up and confirmation (dev mode hands the code back), gives
// it a fan card, a history and a Top 5, and returns its id and the password it made up. On simulated matches the
// history is the one guest accounts used to start with (sim.World.DemoHistory: stubs, settled predictions with their
// points, and five friends). On real matches it is DemoHistoryReal, and the predictions come later from demoPicks.
// Predictions on finished matches can't go through the API (409 locked), which is why the demo is made here.
func seedDemo(ctx context.Context, svc *auth.Service, st *store.Store, w *sim.World, now time.Time, realMatches bool) (userID, password string, err error) {
	if password, err = demoPassword(); err != nil {
		return "", "", err
	}
	sent, err := svc.SignUp(ctx, demoEmail, password, "127.0.0.1")
	if err != nil {
		return "", "", fmt.Errorf("demo account: %w", err)
	}
	in, err := svc.ConfirmEmail(ctx, demoEmail, sent.DevCode)
	if err != nil {
		return "", "", fmt.Errorf("demo account: %w", err)
	}
	u := in.User
	u.Name, u.Handle, u.Bio = "Demo Fan", "demofan", "Here for the last-over finishes."
	u.TeamIDs = []string{"ind", "csk"}
	u.FavouriteTeamID = u.TeamIDs[0]
	if err := st.UpdateUser(ctx, u); err != nil {
		return "", "", err
	}
	if err := st.SetOnboarded(ctx, u.ID, now); err != nil {
		return "", "", err
	}
	logs, preds, follows := w.DemoHistory(u.ID, now)
	if realMatches {
		logs, follows = w.DemoHistoryReal(u.ID, now)
		preds = nil
	}
	if err := st.SeedHistory(ctx, u.ID, logs, preds, follows, now); err != nil {
		return "", "", err
	}
	return u.ID, password, st.SetTop5(ctx, u.ID, w.DemoTop5())
}

// demoPicks gives the demo fan predictions on up to four real matches that finished in the past week with a winner and
// both XIs known, newest first: the winner called right on all but the third, and on the first the top scorer and
// wicket-taker too. They're dated three hours before the first ball, as if made in time; they weren't, which is fine
// for a demo account and nothing else.
func demoPicks(userID string, fixtures []domain.Match, now time.Time) []domain.Prediction {
	type done struct {
		m     domain.Match
		start time.Time
	}
	var ms []done
	for _, m := range fixtures {
		start, err := time.Parse(time.RFC3339, m.StartTime)
		if err != nil || m.State != domain.Completed || m.Result == nil || m.Result.WinnerID == nil ||
			!slices.Contains(m.TeamIDs[:], *m.Result.WinnerID) || len(m.XI[m.TeamIDs[0]]) == 0 ||
			len(m.XI[m.TeamIDs[1]]) == 0 || now.Sub(start) > 7*24*time.Hour {
			continue
		}
		ms = append(ms, done{m, start})
	}
	slices.SortFunc(ms, func(a, b done) int { return b.start.Compare(a.start) })
	picks := []domain.Prediction{}
	for k, d := range ms[:min(4, len(ms))] {
		m := d.m
		pick, other := *m.Result.WinnerID, m.TeamIDs[0]
		if other == pick {
			other = m.TeamIDs[1]
		}
		if k == 2 {
			pick, other = other, pick
		}
		p := domain.Prediction{UserID: userID, MatchID: m.ID, WinnerID: pick, TopRunScorerID: m.XI[pick][0],
			TopWicketTakerID: m.XI[other][len(m.XI[other])-1], CreatedAt: d.start.Add(-3 * time.Hour).UTC().Format(time.RFC3339)}
		if r := m.Result; k == 0 && len(r.TopRunScorerIDs) > 0 && len(r.TopWicketTakerIDs) > 0 {
			p.TopRunScorerID, p.TopWicketTakerID = r.TopRunScorerIDs[0], r.TopWicketTakerIDs[0]
		}
		picks = append(picks, p)
	}
	return picks
}

// addDemoPicks waits for the feed to bring in the week's results with their XIs, which a fresh database gets over its
// first few polls, then saves demoPicks for the demo fan and says so on out. It gives up after 20 minutes, keeping
// whatever picks there are by then.
func addDemoPicks(ctx context.Context, st *store.Store, world httpapi.World, userID string, now func() time.Time, out io.Writer, logger *slog.Logger) {
	deadline := time.Now().Add(20 * time.Minute)
	for {
		picks := demoPicks(userID, world.Fixtures(now()), now())
		late := time.Now().After(deadline)
		if len(picks) >= 3 || (late && len(picks) > 0) {
			for _, p := range picks {
				if err := st.SavePrediction(ctx, p); err != nil {
					logger.Error("demo: saving a prediction", "match", p.MatchID, "err", err)
					return
				}
			}
			fmt.Fprintf(out, "\nDemo: the demo fan's predictions on %d of this week's real matches are in. Pull to refresh.\n\n", len(picks))
			return
		}
		if late {
			logger.Warn("demo: no real results from the past week with their XIs, so the demo fan has no predictions")
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(20 * time.Second):
		}
	}
}

// printDemo tells whoever ran make demo how to sign in.
func printDemo(out io.Writer, port, password string, realMatches bool) {
	matches := "simulated matches"
	wait := ""
	if realMatches {
		matches = "real matches"
		wait = "\nToday's real matches take a minute to load, and the demo fan's predictions on this week's\nresults follow within a few minutes; a line here says when.\n"
	}
	fmt.Fprintf(out, `
Innings demo, on %s: http://localhost:%s

Sign in on the app as
    email     %s
    password  %s
%s
Run make app in another terminal if Metro isn't running. Ctrl+C stops the demo, and the next
make demo starts again from scratch, with a new password.

`, matches, port, demoEmail, password, wait)
}
