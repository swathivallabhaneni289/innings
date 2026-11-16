// Package auth signs fans up and in with an email address and a password (docs/API.md → Auth). A 6-digit code
// emailed to the address confirms it, and another resets a forgotten password. It keeps one session per device and
// guards every other route with Bearer-token middleware.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// Limits on how fast one network or one account can go (docs/API.md → Auth). They live in memory; the per-address
// cap on emailed codes, which protects people's inboxes, is in the database (see SendsPerHour).
const (
	SendsPerIPPerHour    = 20               // code emails one IP address can ask for per rolling hour
	FailsPerAccount      = 10               // wrong passwords per account per FailWindow
	FailWindow           = 15 * time.Minute //
	FailsPerIPPerHour    = 50               // wrong passwords from one IP address per rolling hour
	newAccountColour     = "#3B3F6B"
	newAccountTeam       = "ind"
	sessionTokenHexChars = 64
)

type Service struct {
	Store *store.Store
	// Mailer sends the emailed codes (SMTP, or DevMailer in development). nil means neither is configured, and every
	// call that would email a code answers 503 email-unavailable.
	Mailer Mailer
	// Now is the clock for code expiry, resend gaps and rate limits. It is wall-clock time rather than the
	// simulation's clock: these times are stored across restarts, and INNINGS_ANCHOR can move the sim clock back.
	// Tests swap it to expire codes.
	Now func() time.Time
	Log *slog.Logger
	// Iterations is the PBKDF2 work factor for new password hashes (DefaultIterations). Tests lower it.
	Iterations int
	// The in-memory caps, read once when the first request needs them. Tests may lower or raise them before that.
	SendsPerIP, FailsPerAccount, FailsPerIP int

	once      sync.Once
	sendIP    *windowLimiter // code emails per IP per hour
	failAcct  *windowLimiter // wrong passwords per account per FailWindow
	failIP    *windowLimiter // wrong passwords per IP per hour
	dummyHash string         // checked against when an address has no account, so timing doesn't tell
}

// NewService returns a Service with no Mailer (set one from MailerFromEnv) on the wall clock.
func NewService(st *store.Store, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		Store: st, Now: time.Now, Log: logger, Iterations: DefaultIterations,
		SendsPerIP: SendsPerIPPerHour, FailsPerAccount: FailsPerAccount, FailsPerIP: FailsPerIPPerHour,
	}
}

func (s *Service) init() {
	s.once.Do(func() {
		s.sendIP = newWindowLimiter(s.SendsPerIP, time.Hour)
		s.failAcct = newWindowLimiter(s.FailsPerAccount, FailWindow)
		s.failIP = newWindowLimiter(s.FailsPerIP, time.Hour)
		h, err := HashPassword(newToken(), s.Iterations)
		if err != nil {
			panic(err)
		}
		s.dummyHash = h
	})
}

// Error is a sign-in failure the API reports as-is: HTTP status, error code (docs/API.md → Auth) and message.
type Error struct {
	Status     int
	Code       string
	Message    string
	RetryAfter time.Duration // > 0 sets the Retry-After header
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// SignIn is a device that just signed in: its new session token and the fan it belongs to. IsNew is true the
// first time an account signs in (when its email is confirmed).
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

// newToken is a session or reset token: 32 bytes from crypto/rand as hex. Only HashToken of it is stored.
func newToken() string {
	b := make([]byte, sessionTokenHexChars/2)
	rand.Read(b) // never fails since Go 1.24
	return hex.EncodeToString(b)
}

// HashToken is the SHA-256 hex of a token, the key it is stored under. Tokens carry 256 random bits, so an unsalted
// fast hash is enough: there is nothing to guess.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// openSession signs a device in to userID.
func (s *Service) openSession(ctx context.Context, userID string, now time.Time, isNew bool) (SignIn, error) {
	token := newToken()
	if err := s.Store.CreateSession(ctx, HashToken(token), userID, now); err != nil {
		return SignIn{}, err
	}
	u, _, err := s.Store.UserByID(ctx, userID)
	if err != nil {
		return SignIn{}, err
	}
	acct, err := s.Store.Account(ctx, userID)
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
// guest era aren't sessions, and sessions of phone-era accounts (no password) no longer resolve, so both get 401
// and the app sends those fans to sign up.
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
