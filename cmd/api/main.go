// Command api serves the Innings JSON API.
//
//	PORT            listen port (default 8088; 8080 is often taken by other local APIs)
//	DB_PATH         SQLite file (default ./innings.db)
//	INNINGS_ANCHOR  optional RFC 3339 time to pin the simulated clock (default: process start)
//
// Matches (internal/feed):
//
//	INNINGS_MATCHES real (default): fixtures, live scores, squads and results from ESPN's public scores feed, on
//	                the real clock (INNINGS_ANCHOR is ignored). sim: the simulated cricket world instead.
//
// Sign-up emails (docs/API.md → Auth, "Sending the emails"):
//
//	SMTP_HOST, SMTP_USERNAME, SMTP_PASSWORD
//	                all three: codes are emailed through that server (setting only some is a startup error).
//	                SMTP_PORT defaults to 587; MAIL_FROM defaults to "Innings <SMTP_USERNAME>".
//	INNINGS_DEV_CODES=1
//	                with no SMTP settings: nothing is emailed; the server logs each code and returns it as devCode
//	                (what `make api` does). With neither, every call that would email a code answers 503.
//	INNINGS_ENV_FILE
//	                a file of KEY=VALUE lines to read first (`make api` passes .env); the environment wins over it.
//
// The demo (`make demo`):
//
//	INNINGS_DEMO=1  opens demo@example.com, a fan with stubs, friends, a Top 5 and points from settled predictions
//	                (on real matches, once the feed has the week's results), and prints its password, made up fresh
//	                each run. Needs dev codes, and a database without that account yet (make demo starts from an
//	                empty demo.db).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/auth"
	"github.com/swathivallabhaneni289/innings/internal/feed"
	"github.com/swathivallabhaneni289/innings/internal/httpapi"
	"github.com/swathivallabhaneni289/innings/internal/sim"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	port := envOr("PORT", "8088")
	dbPath := envOr("DB_PATH", "./innings.db")

	started := time.Now()
	anchor := started
	if a := os.Getenv("INNINGS_ANCHOR"); a != "" {
		t, err := time.Parse(time.RFC3339, a)
		if err != nil {
			return errors.New("INNINGS_ANCHOR must be RFC 3339, e.g. 2026-10-22T18:00:00Z")
		}
		anchor = t
	}
	// The sim clock starts at the anchor and runs in real time from there.
	now := func() time.Time { return anchor.Add(time.Since(started)) }

	if path := os.Getenv("INNINGS_ENV_FILE"); path != "" {
		keys, err := loadEnvFile(path)
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			logger.Info("settings read from "+path, "keys", strings.Join(keys, ", ")) // names only, never values
		}
	}
	mailer, mailMode, err := auth.MailerFromEnv(os.Getenv, logger)
	if err != nil {
		return err
	}
	demo := os.Getenv("INNINGS_DEMO") == "1"
	if _, dev := mailer.(auth.DevMailer); demo && !dev {
		return errors.New("INNINGS_DEMO=1 needs INNINGS_DEV_CODES=1 with no SMTP settings (make demo sets them)")
	}

	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	simWorld := sim.NewWorld(anchor)
	var world httpapi.World = simWorld
	matches, realMatches := "simulated (INNINGS_MATCHES=sim)", false
	if src := feed.FromEnv(os.Getenv, st, simWorld.Teams(), logger); src != nil {
		if err := src.Load(ctx); err != nil {
			return err
		}
		go src.Run(ctx)
		world, matches, now, realMatches = feed.NewWorld(simWorld, src), "real (ESPN scores feed)", time.Now, true
	}

	api := httpapi.New(st, world, now, logger)
	api.Auth.Mailer = mailer
	switch m := mailer.(type) {
	case *auth.SMTPMailer:
		logger.Info("sign-up emails: SMTP", "server", m.Host+":"+m.Port, "from", m.From.String())
	case auth.DevMailer:
		logger.Warn("sign-up emails: DEV mode (INNINGS_DEV_CODES=1): codes are logged and returned as devCode, never emailed")
	default:
		logger.Warn("sign-up emails: OFF (no SMTP settings and INNINGS_DEV_CODES isn't 1): sign-up and password resets answer 503", "mode", mailMode)
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	// Listen first, so a port that's taken stops the server before the demo account is made.
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		if demo {
			return fmt.Errorf("%w: stop make api (or whatever has port %s) and run make demo again", err, port)
		}
		return err
	}
	if demo {
		userID, password, err := seedDemo(ctx, api.Auth, st, simWorld, now(), realMatches)
		if err != nil {
			ln.Close()
			return err
		}
		printDemo(os.Stdout, port, password, realMatches)
		if realMatches {
			go addDemoPicks(ctx, st, world, userID, now, os.Stdout, logger)
		}
	}
	errc := make(chan error, 1)
	go func() {
		logger.Info("innings api listening", "addr", "http://localhost:"+port, "db", dbPath, "matches", matches, "anchor", anchor.UTC().Format(time.RFC3339))
		errc <- srv.Serve(ln)
	}()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
