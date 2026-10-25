package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// Sign-in code limits (docs/API.md → Auth, "Codes").
const (
	CodeTTL            = 10 * time.Minute // how long a code works
	ResendGap          = 30 * time.Second // before another code can go to the same number
	MaxTries           = 5                // guesses allowed per code
	SendsPerHour       = 5                // codes per number per rolling hour (stored in the database)
	StartsPerIPPerHour = 20               // starts per IP address per rolling hour (in memory)
)

// ErrCodeGone means the provider has no live verification for the number: expired, already approved, out of tries,
// or made under a provider setting this server no longer has. The API answers 410 code-expired.
var ErrCodeGone = errors.New("no live code")

// Verifier gets sign-in codes to phones. There are two:
//   - DevVerifier (INNINGS_DEV_OTP=1): this server makes each code, stores only its hash, checks it here, and hands
//     it back to the app as devCode. Nothing is texted.
//   - Twilio: Twilio Verify makes, texts and checks the code; this server tracks only expiry, tries and the send log.
type Verifier interface {
	// ProviderCodes is true when the provider makes and checks its own codes (Twilio Verify).
	ProviderCodes() bool
	// Send delivers a new code to phone. code is the code this server made, or "" when ProviderCodes.
	Send(ctx context.Context, phone, code string) error
	// Check asks the provider whether code is the number's current code (only when ProviderCodes). It returns
	// ErrCodeGone when the provider has no live verification for the number.
	Check(ctx context.Context, phone, code string) (bool, error)
}

// devCodes is implemented only by DevVerifier. Its method is unexported, so nothing outside this package can opt
// in: no other verifier's code ever reaches a response, and no server-made code is accepted under another verifier.
type devCodes interface{ revealsCodes() }

// DevVerifier is for local development with no SMS provider: it writes each code to the log instead of texting it.
type DevVerifier struct{ Log *slog.Logger }

func (DevVerifier) revealsCodes()       {}
func (DevVerifier) ProviderCodes() bool { return false }

func (d DevVerifier) Send(_ context.Context, phone, code string) error {
	log := d.Log
	if log == nil {
		log = slog.Default()
	}
	log.Info("dev sign-in code", "phone", phone, "code", code)
	return nil
}

// Check is never called for a verifier whose codes the server checks itself.
func (DevVerifier) Check(context.Context, string, string) (bool, error) { return false, ErrCodeGone }

// VerifierFromEnv picks how codes are sent and names the choice for the startup log:
//   - "twilio": Twilio Verify, when TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN and TWILIO_VERIFY_SERVICE_SID are all set;
//   - "dev": DevVerifier, when INNINGS_DEV_OTP=1 and Twilio isn't configured;
//   - "off": nil, so /v1/auth/start answers 503 sms-unavailable.
//
// Setting only some of the Twilio variables is an error rather than a quiet fall back to dev mode, which would hand
// sign-in codes to whoever asks for them.
func VerifierFromEnv(getenv func(string) string, logger *slog.Logger) (Verifier, string, error) {
	sid, token, service := getenv("TWILIO_ACCOUNT_SID"), getenv("TWILIO_AUTH_TOKEN"), getenv("TWILIO_VERIFY_SERVICE_SID")
	switch set := (sid != "") || (token != "") || (service != ""); {
	case sid != "" && token != "" && service != "":
		return &Twilio{AccountSID: sid, AuthToken: token, ServiceSID: service}, "twilio", nil
	case set:
		return nil, "", errors.New("set all of TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN and TWILIO_VERIFY_SERVICE_SID, or none")
	case getenv("INNINGS_DEV_OTP") == "1":
		return DevVerifier{Log: logger}, "dev", nil
	default:
		return nil, "off", nil
	}
}

var (
	phoneRE = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	codeRE  = regexp.MustCompile(`^\d{6}$`)
)

// NormPhone returns raw as E.164, dropping the spaces, dashes and brackets people type, or ok=false when it isn't
// a full international number.
func NormPhone(raw string) (string, bool) {
	p := strings.Map(func(r rune) rune {
		if strings.ContainsRune(" -()", r) {
			return -1
		}
		return r
	}, raw)
	return p, phoneRE.MatchString(p)
}

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

func errBadPhone() error {
	return &Error{Status: http.StatusBadRequest, Code: "bad-phone", Message: "Enter the full number with its country code, like +919876543210."}
}

func errExpired() error {
	return &Error{Status: http.StatusGone, Code: "code-expired", Message: "That code has expired or was already used. Ask for a new one."}
}

func errSMSFailed() error {
	return &Error{Status: http.StatusBadGateway, Code: "sms-failed", Message: "We couldn't send or check the text just now. Try again in a moment."}
}

// Start sends a sign-in code to rawPhone, coming from IP address ip. The answer is the same whether or not the
// number has an account. A send the provider refuses (502) still counts towards the number's hourly cap, so a
// failing number can't be hammered; its code row is dropped so verify says expired rather than asking the provider.
func (s *Service) Start(ctx context.Context, rawPhone, ip string) (domain.AuthStartResponse, error) {
	var resp domain.AuthStartResponse
	if s.Verifier == nil {
		return resp, &Error{Status: http.StatusServiceUnavailable, Code: "sms-unavailable", Message: "Sign-in by text isn't available on this server right now."}
	}
	phone, ok := NormPhone(rawPhone)
	if !ok {
		return resp, errBadPhone()
	}
	now := s.Now()
	if ok, wait := s.ipLimit.allow(ipKey(ip), now); !ok {
		return resp, &Error{Status: http.StatusTooManyRequests, Code: "too-many", Message: "Too many sign-ins from this network. Try again later.", RetryAfter: wait}
	}
	code, hash := "", ""
	if !s.Verifier.ProviderCodes() {
		code = newCode()
		hash = hashCode(code)
	}
	o := store.OTP{Phone: phone, CodeHash: hash, ExpiresAt: now.Add(CodeTTL), SentAt: now}
	wait, err := s.Store.StartOTP(ctx, o, ResendGap, SendsPerHour)
	switch {
	case errors.Is(err, store.ErrTooSoon):
		return resp, &Error{Status: http.StatusTooManyRequests, Code: "too-soon", Message: "A code is on its way. You can ask for another in a few seconds.", RetryAfter: wait}
	case errors.Is(err, store.ErrTooMany):
		return resp, &Error{Status: http.StatusTooManyRequests, Code: "too-many", Message: "That's 5 codes this hour for this number. Try again later.", RetryAfter: wait}
	case err != nil:
		return resp, err
	}
	if err := s.Verifier.Send(ctx, phone, code); err != nil {
		s.Log.Error("sign-in code send failed", "err", err) // no phone in the log: it's personal data
		if err := s.Store.DropOTP(ctx, phone); err != nil {
			s.Log.Error("drop code after failed send", "err", err)
		}
		return resp, errSMSFailed()
	}
	resp = domain.AuthStartResponse{Phone: phone, ResendInSeconds: int(ResendGap / time.Second), ExpiresInSeconds: int(CodeTTL / time.Second)}
	if _, dev := s.Verifier.(devCodes); dev {
		resp.DevCode = code
	}
	return resp, nil
}

// Verify checks code for rawPhone and signs in, creating the account the first time the number verifies. Each call
// spends one of the code's tries before anything is compared; the right code is then consumed so it can't be
// used again.
func (s *Service) Verify(ctx context.Context, rawPhone, code string) (SignIn, error) {
	phone, ok := NormPhone(rawPhone)
	if !ok {
		return SignIn{}, errBadPhone()
	}
	now := s.Now()
	o, live, err := s.Store.TryOTP(ctx, phone, now, MaxTries)
	if err != nil {
		return SignIn{}, err
	}
	if !live {
		return SignIn{}, errExpired()
	}
	right, err := s.checkCode(ctx, o, strings.TrimSpace(code))
	if errors.Is(err, ErrCodeGone) {
		if err := s.Store.DropOTP(ctx, phone); err != nil {
			return SignIn{}, err
		}
		return SignIn{}, errExpired()
	}
	if err != nil {
		s.Log.Error("sign-in code check failed", "err", err)
		return SignIn{}, errSMSFailed()
	}
	if !right {
		return SignIn{}, &Error{Status: http.StatusBadRequest, Code: "bad-code", Message: wrongCodeText(MaxTries - o.Attempts)}
	}
	if used, err := s.Store.ConsumeOTP(ctx, phone, o.SentAt); err != nil {
		return SignIn{}, err
	} else if !used {
		return SignIn{}, errExpired() // a concurrent verify used it first
	}
	return s.signIn(ctx, phone, now)
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

// checkCode compares code with the live one: by hash in constant time for a code this server made (dev mode only),
// or by asking the provider for one it made. A code made under a different verifier setting (say, a dev code
// after the server moved to Twilio) is treated as gone.
func (s *Service) checkCode(ctx context.Context, o store.OTP, code string) (bool, error) {
	if !codeRE.MatchString(code) {
		return false, nil // a malformed guess still spends its try, and never reaches the provider
	}
	_, dev := s.Verifier.(devCodes)
	switch {
	case o.CodeHash != "" && dev:
		return subtle.ConstantTimeCompare([]byte(hashCode(code)), []byte(o.CodeHash)) == 1, nil
	case o.CodeHash == "" && s.Verifier != nil && s.Verifier.ProviderCodes():
		return s.Verifier.Check(ctx, o.Phone, code)
	default:
		return false, ErrCodeGone
	}
}
