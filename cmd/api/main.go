// Command api serves the Innings JSON API.
//
//	PORT            listen port (default 8088; 8080 is often taken by other local APIs)
//	DB_PATH         SQLite file (default ./innings.db)
//	INNINGS_ANCHOR  optional RFC 3339 time to pin the simulated clock (default: process start)
//
// Sign-in codes (docs/API.md → Auth, "SMS providers"):
//
//	TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_VERIFY_SERVICE_SID
//	                all three: codes go through Twilio Verify (setting only some is a startup error)
//	INNINGS_DEV_OTP=1
//	                with no Twilio settings: the server makes codes, logs them and returns them as devCode
//	                (what `make api` does). With neither, /v1/auth/start answers 503 sms-unavailable.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/auth"
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

	verifier, smsMode, err := auth.VerifierFromEnv(os.Getenv, logger)
	if err != nil {
		return err
	}

	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	api := httpapi.New(st, sim.NewWorld(anchor), now, logger)
	api.Auth.Verifier = verifier
	switch smsMode {
	case "twilio":
		logger.Info("sign-in codes: Twilio Verify")
	case "dev":
		logger.Warn("sign-in codes: DEV mode (INNINGS_DEV_OTP=1): codes are logged and returned as devCode, never texted")
	default:
		logger.Warn("sign-in codes: OFF (no Twilio settings and INNINGS_DEV_OTP isn't 1): /v1/auth/start answers 503")
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		logger.Info("innings api listening", "addr", "http://localhost:"+port, "db", dbPath, "anchor", anchor.UTC().Format(time.RFC3339))
		errc <- srv.ListenAndServe()
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
