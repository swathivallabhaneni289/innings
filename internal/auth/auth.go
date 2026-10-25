// Package auth signs fans in with a phone number and a texted code (docs/API.md → Auth), keeps one session per
// device, and guards every other route with Bearer-token middleware.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

type Service struct {
	Store *store.Store
	// Verifier gets codes to phones (Twilio Verify, or DevVerifier in development). nil means neither is configured,
	// and /v1/auth/start answers 503 sms-unavailable.
	Verifier Verifier
	// Now is the clock for code expiry, resend gaps and rate limits. It is wall-clock time rather than the
	// simulation's clock: these times are stored across restarts, and INNINGS_ANCHOR can move the sim clock back.
	// Tests swap it to expire codes.
	Now func() time.Time
	Log *slog.Logger

	ipLimit *windowLimiter // starts per IP per hour, in memory
}

// NewService returns a Service with no Verifier (set one from VerifierFromEnv) on the wall clock.
func NewService(st *store.Store, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{Store: st, Now: time.Now, Log: logger, ipLimit: newWindowLimiter(StartsPerIPPerHour, time.Hour)}
}

// Error is a sign-in failure the API reports as-is: HTTP status, error code (docs/API.md → Auth) and message.
type Error struct {
	Status     int
	Code       string
	Message    string
	RetryAfter time.Duration // > 0 sets the Retry-After header
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// SignIn is a successful verify: a new session token for this device and the fan it belongs to.
type SignIn struct {
	Token   string
	User    domain.User
	Account domain.Account
	IsNew   bool
}

// BearerToken extracts the token from "Authorization: Bearer <token>" ("" if absent).
func BearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// newToken is a session token: 32 bytes from crypto/rand as hex. Only HashToken of it is stored.
func newToken() string {
	b := make([]byte, 32)
	rand.Read(b) // never fails since Go 1.24
	return hex.EncodeToString(b)
}

// HashToken is the SHA-256 hex of a session token, the key it is stored under. Tokens carry 256 random bits, so an
// unsalted fast hash is enough: there is nothing to guess.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// signIn opens a session for the account with this phone, creating the account the first time the number verifies.
// New accounts start empty: blank name and handle, the default avatar colour and team, and no history.
func (s *Service) signIn(ctx context.Context, phone string, now time.Time) (SignIn, error) {
	u, found, err := s.Store.UserByPhone(ctx, phone)
	if err != nil {
		return SignIn{}, err
	}
	isNew := false
	if !found {
		u = domain.User{ID: store.NewID("u-"), AvatarColour: "#3B3F6B", FavouriteTeamID: "ind"}
		switch err := s.Store.CreateAccount(ctx, u, phone, now); {
		case errors.Is(err, store.ErrPhoneTaken): // a concurrent first sign-in for this number won; join it
			if u, found, err = s.Store.UserByPhone(ctx, phone); err != nil || !found {
				return SignIn{}, fmt.Errorf("account for phone vanished: %v", err)
			}
		case err != nil:
			return SignIn{}, err
		default:
			isNew = true
		}
	}
	token := newToken()
	if err := s.Store.CreateSession(ctx, HashToken(token), u.ID, now); err != nil {
		return SignIn{}, err
	}
	acct, err := s.Store.Account(ctx, u.ID)
	if err != nil {
		return SignIn{}, err
	}
	return SignIn{Token: token, User: u, Account: acct, IsNew: isNew}, nil
}

// Logout ends one device's session (the one whose token hashes to tokenHash).
func (s *Service) Logout(ctx context.Context, tokenHash string) error {
	return s.Store.DeleteSession(ctx, tokenHash)
}

type ctxKey struct{}
type sessionKey struct{}

// UserFrom returns the authenticated user put in the context by Middleware.
func UserFrom(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(ctxKey{}).(domain.User)
	return u, ok
}

// WithUser returns ctx carrying u (for tests and Middleware).
func WithUser(ctx context.Context, u domain.User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

// SessionFrom returns the hash of the session token this request was authenticated with ("" if none).
func SessionFrom(ctx context.Context) string {
	h, _ := ctx.Value(sessionKey{}).(string)
	return h
}

// Middleware resolves the Bearer token through the sessions table to a user, or answers 401 JSON. Tokens from the
// guest era aren't sessions, so they get 401 too and the app sends those fans to sign in.
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := BearerToken(r)
		if token == "" {
			unauthorized(w, "missing bearer token")
			return
		}
		hash := HashToken(token)
		u, ok, err := s.Store.UserBySession(r.Context(), hash)
		if err != nil {
			s.Log.Error("session lookup", "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal", "message": "could not check token"})
			return
		}
		if !ok {
			unauthorized(w, "unknown or signed-out token")
			return
		}
		ctx := context.WithValue(WithUser(r.Context(), u), sessionKey{}, hash)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": msg})
}
