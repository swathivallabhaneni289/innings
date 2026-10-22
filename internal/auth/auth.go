// Package auth handles guest accounts: bootstrap (create or resume) and Bearer-token middleware.
package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/sim"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// DemoSource seeds new guests; *sim.World implements it.
type DemoSource interface {
	DemoHistory(userID string, now time.Time) ([]domain.MatchLog, []domain.Prediction, []string)
}

var _ DemoSource = (*sim.World)(nil)

type Service struct {
	Store *store.Store
	World DemoSource
	Now   func() time.Time
}

// BearerToken extracts the token from "Authorization: Bearer <token>" ("" if absent).
func BearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// Bootstrap resumes the guest owning token, or (no/unknown token) creates a new guest seeded with demo history.
func (s *Service) Bootstrap(ctx context.Context, token string) (domain.User, string, error) {
	if token != "" {
		if u, ok, err := s.Store.UserByToken(ctx, token); err != nil {
			return u, "", err
		} else if ok {
			return u, token, nil
		}
	}
	now := s.Now()
	suffix := store.RandomHex(3)
	u := domain.User{
		ID:              "u-" + store.RandomHex(8),
		Name:            "Guest Fan",
		Handle:          "fan-" + suffix,
		AvatarColour:    "#B3261E",
		FavouriteTeamID: "ind",
		Bio:             "",
	}
	token = store.RandomHex(24)
	if err := s.Store.CreateUser(ctx, u, token, now); err != nil {
		return u, "", err
	}
	logs, preds, follows := s.World.DemoHistory(u.ID, now)
	if err := s.Store.SeedHistory(ctx, u.ID, logs, preds, follows, now); err != nil {
		return u, "", err
	}
	return u, token, nil
}

type ctxKey struct{}

// UserFrom returns the authenticated user put in the context by Middleware.
func UserFrom(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(ctxKey{}).(domain.User)
	return u, ok
}

// WithUser returns ctx carrying u (for tests and Middleware).
func WithUser(ctx context.Context, u domain.User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

// Middleware resolves the Bearer token to a user, or answers 401 JSON.
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := BearerToken(r)
		if token == "" {
			unauthorized(w, "missing bearer token")
			return
		}
		u, ok, err := s.Store.UserByToken(r.Context(), token)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal", "message": "could not check token"})
			return
		}
		if !ok {
			unauthorized(w, "unknown token")
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
	})
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": msg})
}
