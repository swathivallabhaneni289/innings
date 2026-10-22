// Package httpapi serves the Innings JSON API (docs/API.md) by merging the simulated world (internal/sim) with
// real guests' rows from the store.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/auth"
	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/sim"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// World is what the API needs from the simulation; *sim.World implements it (tests use a fake).
type World interface {
	Teams() map[string]domain.Team
	Players() map[string]domain.Player
	Fixtures(now time.Time) []domain.Match
	Classics() []domain.Match
	Friends() []domain.User
	PublicFans() []domain.User
	User(id string) (domain.User, bool)
	Logs(now time.Time) []domain.MatchLog
	Predictions(now time.Time) []domain.Prediction
	BaseLikes(logID string) int
	BaseThereToo(logID string) int
	Replies(logID string, now time.Time) []domain.Reply
	Crowd(matchID string) sim.Crowd
	FriendsWatching(matchID string, now time.Time) []string
	DemoHistory(userID string, now time.Time) ([]domain.MatchLog, []domain.Prediction, []string)
}

var _ World = (*sim.World)(nil)

type Server struct {
	Store *store.Store
	World World
	Auth  *auth.Service
	Now   func() time.Time
	Log   *slog.Logger
}

func New(st *store.Store, world World, now func() time.Time, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{Store: st, World: world, Now: now, Log: logger, Auth: &auth.Service{Store: st, World: world, Now: now}}
}

// Handler returns the full API with CORS, request logging and panic recovery.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, "ok")
	})
	mux.HandleFunc("POST /v1/bootstrap", s.handleBootstrap)

	authed := func(pattern string, fn func(v *view, w http.ResponseWriter, r *http.Request)) {
		mux.Handle(pattern, s.Auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			me, _ := auth.UserFrom(r.Context())
			fn(s.newView(r.Context(), me), w, r)
		})))
	}
	// Reads.
	authed("GET /v1/home", s.handleHome)
	authed("GET /v1/matches", s.handleMatches)
	authed("GET /v1/matches/{id}", s.handleMatch)
	authed("GET /v1/vault", s.handleVault)
	authed("GET /v1/vault/rankings", s.handleRankings)
	authed("GET /v1/users/{id}", s.handleUser)
	authed("GET /v1/friends", s.handleFriends)
	authed("GET /v1/logs/{id}", s.handleLog)
	authed("GET /v1/search", s.handleSearch)
	authed("GET /v1/log/suggestions", s.handleLogSuggestions)
	authed("GET /v1/notifications", s.handleNotifications)
	// Writes.
	authed("POST /v1/logs", s.handleCreateLog)
	authed("DELETE /v1/logs/{id}", s.handleDeleteLog)
	authed("PUT /v1/matches/{id}/prediction", s.handlePrediction)
	authed("PUT /v1/matches/{id}/watching", s.handleWatching)
	authed("POST /v1/logs/{id}/like", s.handleLike)
	authed("POST /v1/logs/{id}/there-too", s.handleThereToo)
	authed("POST /v1/logs/{id}/replies", s.handleReply)
	authed("POST /v1/users/{id}/follow", s.handleFollow)
	authed("PUT /v1/me", s.handleUpdateMe)
	authed("PUT /v1/me/top5", s.handleTop5)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, http.StatusNotFound, "not-found", "no such route")
	})
	return s.cors(s.logRequests(s.recoverer(mux)))
}

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	me, token, err := s.Auth.Bootstrap(r.Context(), auth.BearerToken(r))
	if err != nil {
		s.Log.Error("bootstrap", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal", "could not create guest")
		return
	}
	v := s.newView(r.Context(), me)
	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"me":         me,
		"teams":      v.teams(),
		"players":    nonNilMap(s.World.Players()),
		"serverTime": v.now.UTC().Format(time.RFC3339),
	})
}

// ---- middleware ----

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		h.Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		s.Log.Info("request", "method", r.Method, "path", r.URL.Path, "status", sw.status, "dur", time.Since(start).Round(time.Microsecond))
	})
}

// storeErr is panicked by view helpers when the database fails; recoverer turns it (and any panic) into a 500.
type storeErr struct{ error }

func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				if p == http.ErrAbortHandler {
					panic(p)
				}
				var se storeErr
				if err, ok := p.(error); ok && errors.As(err, &se) {
					s.Log.Error("store", "path", r.URL.Path, "err", se.error)
				} else {
					s.Log.Error("panic", "path", r.URL.Path, "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
				}
				writeErr(w, http.StatusInternalServerError, "internal", "something went wrong")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(storeErr{err})
	}
	return v
}

func check(err error) {
	if err != nil {
		panic(storeErr{err})
	}
}

// ---- JSON helpers ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"error": code, "message": msg})
}

// decode reads a JSON body (max 1 MB) into dst; on failure it writes a 400 and returns false.
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(dst); err != nil {
		writeErr(w, http.StatusBadRequest, "bad-json", "could not read JSON body: "+err.Error())
		return false
	}
	return true
}
