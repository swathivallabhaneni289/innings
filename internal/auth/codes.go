package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// Emailed-code limits (docs/API.md → Auth, "Codes").
const (
	CodeTTL       = 10 * time.Minute // how long a code works
	ResendGap     = 30 * time.Second // before another code can go to the same address
	MaxTries      = 5                // guesses allowed per code
	SendsPerHour  = 5                // codes per address per rolling hour (stored in the database)
	ResetTokenTTL = 15 * time.Minute // how long the right reset code leaves to choose a new password
)

// Purpose says what a code is for. Each address has at most one live code per purpose.
type Purpose string

const (
	PurposeVerify Purpose = "verify" // confirms the address after sign-up
	PurposeReset  Purpose = "reset"  // starts a new password
)

var codeRE = regexp.MustCompile(`^\d{6}$`)

// newCode is 6 uniformly random digits from crypto/rand.
func newCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		panic(err) // crypto/rand doesn't fail on supported platforms
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func errExpired() error {
	return &Error{Status: http.StatusGone, Code: "code-expired", Message: "That code has expired or was already used. Ask for a new one."}
}

func errEmailFailed() error {
	return &Error{Status: http.StatusBadGateway, Code: "email-failed", Message: "We couldn't send the email just now. Try again in a moment."}
}

// sendCode makes a new code for email and purpose and emails it when deliver is true. deliver is false when there
// is no account to send it to; the answer is the same either way, and the address's limits still count the send, so
// nobody can learn from the response whether an address has an account. With soft, a code that went out under the
// resend gap ago isn't an error: the answer just says how long until another can be sent, and the earlier code
// still works. A send the mailer refuses (502) still counts towards the cap; its code is dropped.
func (s *Service) sendCode(ctx context.Context, email string, purpose Purpose, deliver, soft bool) (domain.AuthCodeResponse, error) {
	if s.Mailer == nil {
		return domain.AuthCodeResponse{}, errEmailUnavailable()
	}
	now := s.Now()
	resp := domain.AuthCodeResponse{Email: email, ResendInSeconds: int(ResendGap / time.Second), ExpiresInSeconds: int(CodeTTL / time.Second)}
	code := newCode()
	c := store.EmailCode{Email: email, Purpose: string(purpose), CodeHash: hashCode(code), ExpiresAt: now.Add(CodeTTL), SentAt: now}
	wait, err := s.Store.StartCode(ctx, c, ResendGap, SendsPerHour)
	switch {
	case errors.Is(err, store.ErrTooSoon) && soft:
		resp.ResendInSeconds = int(math.Ceil(wait.Seconds()))
		return resp, nil
	case errors.Is(err, store.ErrTooSoon):
		return resp, &Error{Status: http.StatusTooManyRequests, Code: "too-soon", Message: "A code is on its way. You can ask for another in a few seconds.", RetryAfter: wait}
	case errors.Is(err, store.ErrTooMany):
		return resp, &Error{Status: http.StatusTooManyRequests, Code: "too-many", Message: fmt.Sprintf("That's %d codes this hour for this email. Try again later.", SendsPerHour), RetryAfter: wait}
	case err != nil:
		return resp, err
	}
	if !deliver {
		return resp, nil
	}
	if err := s.Mailer.Send(ctx, codeEmail(email, purpose, code)); err != nil {
		s.Log.Error("code email failed", "purpose", purpose, "err", err) // no address in the log: it's personal data
		if err := s.Store.DropCode(ctx, email, string(purpose)); err != nil {
			s.Log.Error("drop code after failed send", "err", err)
		}
		return domain.AuthCodeResponse{}, errEmailFailed()
	}
	if _, dev := s.Mailer.(devCodes); dev {
		resp.DevCode = code
	}
	return resp, nil
}

// useCode checks code against the address's live code for purpose and, when it is right, uses it up. Each call
// spends one of the code's tries before anything is compared, and a malformed guess spends one too.
func (s *Service) useCode(ctx context.Context, email string, purpose Purpose, code string) error {
	c, live, err := s.Store.TryCode(ctx, email, string(purpose), s.Now(), MaxTries)
	if err != nil {
		return err
	}
	if !live {
		return errExpired()
	}
	code = strings.TrimSpace(code)
	if !codeRE.MatchString(code) || subtle.ConstantTimeCompare([]byte(hashCode(code)), []byte(c.CodeHash)) != 1 {
		return &Error{Status: http.StatusBadRequest, Code: "bad-code", Message: wrongCodeText(MaxTries - c.Attempts)}
	}
	if used, err := s.Store.ConsumeCode(ctx, email, string(purpose), c.SentAt); err != nil {
		return err
	} else if !used {
		return errExpired() // a concurrent check used it first
	}
	return nil
}

func wrongCodeText(left int) string {
	switch left {
	case 0:
		return "That code isn't right, and that was the last try. Ask for a new code."
	case 1:
		return "That code isn't right. 1 try left."
	default:
		return fmt.Sprintf("That code isn't right. %d tries left.", left)
	}
}
