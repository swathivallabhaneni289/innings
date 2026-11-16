package auth

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// Sign-up, confirming the address, signing in, and resetting or changing a password (docs/API.md → Auth).

func errBadEmail() error {
	return &Error{Status: http.StatusBadRequest, Code: "bad-email", Message: "Enter your full email address, like name@example.com."}
}

func errWeakPassword(problem string) error {
	return &Error{Status: http.StatusBadRequest, Code: "weak-password", Message: problem}
}

func errEmailUnavailable() error {
	return &Error{Status: http.StatusServiceUnavailable, Code: "email-unavailable", Message: "Signing up by email isn't available on this server right now."}
}

func errBadCredentials() error {
	return &Error{Status: http.StatusUnauthorized, Code: "bad-credentials", Message: "That email and password don't match."}
}

func errBadPassword() error {
	return &Error{Status: http.StatusForbidden, Code: "bad-password", Message: "That isn't your current password."}
}

func errTooManyTries(wait time.Duration) error {
	mins := max(1, int(math.Ceil(wait.Minutes())))
	unit := "minutes"
	if mins == 1 {
		unit = "minute"
	}
	return &Error{Status: http.StatusTooManyRequests, Code: "too-many",
		Message: fmt.Sprintf("Too many wrong passwords. Try again in %d %s, or reset your password.", mins, unit), RetryAfter: wait}
}

func errResetExpired() error {
	return &Error{Status: http.StatusGone, Code: "reset-expired", Message: "That reset has run out. Ask for a new code from Forgot password."}
}

// allowSend spends one of ip's code emails for the hour, before any work is done for the request.
func (s *Service) allowSend(ip string, now time.Time) error {
	s.init()
	if ok, wait := s.sendIP.allow(ipKey(ip), now); !ok {
		return &Error{Status: http.StatusTooManyRequests, Code: "too-many", Message: "Too many codes asked for from this network. Try again later.", RetryAfter: wait}
	}
	return nil
}

// SignUp makes an account for email with password and emails a code to confirm the address. An address that a
// confirmed account already has is 409 email-taken. An unconfirmed account with the address is taken over (the
// password is replaced and a new code sent), since only whoever reads the inbox can confirm it.
func (s *Service) SignUp(ctx context.Context, rawEmail, password, ip string) (domain.AuthCodeResponse, error) {
	var none domain.AuthCodeResponse
	if s.Mailer == nil {
		return none, errEmailUnavailable()
	}
	email, ok := NormEmail(rawEmail)
	if !ok {
		return none, errBadEmail()
	}
	if p := PasswordProblem(password, email); p != "" {
		return none, errWeakPassword(p)
	}
	now := s.Now()
	if err := s.allowSend(ip, now); err != nil { // before hashing, so a flood can't burn the CPU
		return none, err
	}
	hash, err := HashPassword(password, s.Iterations)
	if err != nil {
		return none, err
	}
	fresh := domain.User{ID: store.NewID("u-"), AvatarColour: newAccountColour, FavouriteTeamID: newAccountTeam}
	if _, err := s.Store.SignUp(ctx, fresh, email, hash, now); errors.Is(err, store.ErrEmailTaken) {
		return none, &Error{Status: http.StatusConflict, Code: "email-taken", Message: "There's already an account with this email. Sign in, or reset the password if you've forgotten it."}
	} else if err != nil {
		return none, err
	}
	return s.sendCode(ctx, email, PurposeVerify, true, true)
}

// ResendConfirmation emails a new confirmation code to an address whose account isn't confirmed yet. The answer is
// the same whether or not there is such an account.
func (s *Service) ResendConfirmation(ctx context.Context, rawEmail, ip string) (domain.AuthCodeResponse, error) {
	email, ok := NormEmail(rawEmail)
	if !ok {
		return domain.AuthCodeResponse{}, errBadEmail()
	}
	if s.Mailer == nil {
		return domain.AuthCodeResponse{}, errEmailUnavailable()
	}
	if err := s.allowSend(ip, s.Now()); err != nil {
		return domain.AuthCodeResponse{}, err
	}
	c, found, err := s.Store.CredentialsByEmail(ctx, email)
	if err != nil {
		return domain.AuthCodeResponse{}, err
	}
	return s.sendCode(ctx, email, PurposeVerify, found && !c.Confirmed, false)
}

// ConfirmEmail checks the confirmation code for email and signs this device in: the account's first sign-in.
func (s *Service) ConfirmEmail(ctx context.Context, rawEmail, code string) (SignIn, error) {
	email, ok := NormEmail(rawEmail)
	if !ok {
		return SignIn{}, errBadEmail()
	}
	if err := s.useCode(ctx, email, PurposeVerify, code); err != nil {
		return SignIn{}, err
	}
	c, found, err := s.Store.CredentialsByEmail(ctx, email)
	if err != nil {
		return SignIn{}, err
	}
	if !found {
		return SignIn{}, errExpired() // the account was cleared or deleted since the code went out
	}
	now := s.Now()
	if err := s.Store.ConfirmEmail(ctx, c.UserID, now); err != nil {
		return SignIn{}, err
	}
	return s.openSession(ctx, c.UserID, now, !c.Confirmed)
}

// checkPassword compares password with an account's hash under the failure caps for key (the account) and ip. An
// account without a hash, or none at all (account == nil), is compared with a dummy hash, so the time taken doesn't
// tell whether the address has an account. On a match it clears the account's failures and stores a stronger hash
// if the work factor has gone up since.
func (s *Service) checkPassword(ctx context.Context, account *store.Credentials, key, password, ip string) (bool, error) {
	s.init()
	now := s.Now()
	if wait := max(s.failAcct.wait(key, now), s.failIP.wait(ipKey(ip), now)); wait > 0 {
		return false, errTooManyTries(wait)
	}
	stored := s.dummyHash
	if account != nil && account.PasswordHash != "" {
		stored = account.PasswordHash
	}
	ok, rehash := CheckPassword(stored, password, s.Iterations)
	if !ok || account == nil || account.PasswordHash == "" {
		s.failAcct.record(key, now)
		s.failIP.record(ipKey(ip), now)
		return false, nil
	}
	s.failAcct.forget(key)
	if rehash {
		if h, err := HashPassword(password, s.Iterations); err == nil {
			if err := s.Store.RehashPassword(ctx, account.UserID, h); err != nil {
				s.Log.Error("rehash password", "err", err)
			}
		}
	}
	return true, nil
}

// SignInWithPassword signs this device in with an email and password. A wrong password and an unknown address get
// the same answer, 401 bad-credentials. Only with the right password does an unconfirmed account learn it must
// confirm its address first (403 email-unconfirmed).
func (s *Service) SignInWithPassword(ctx context.Context, rawEmail, password, ip string) (SignIn, error) {
	email, ok := NormEmail(rawEmail)
	if !ok {
		return SignIn{}, errBadEmail()
	}
	c, found, err := s.Store.CredentialsByEmail(ctx, email)
	if err != nil {
		return SignIn{}, err
	}
	var account *store.Credentials
	if found {
		account = &c
	}
	right, err := s.checkPassword(ctx, account, "email:"+email, password, ip)
	if err != nil {
		return SignIn{}, err
	}
	if !right {
		return SignIn{}, errBadCredentials()
	}
	if !c.Confirmed {
		return SignIn{}, &Error{Status: http.StatusForbidden, Code: "email-unconfirmed", Message: "Confirm your email to finish signing up. We'll send you a new code."}
	}
	return s.openSession(ctx, c.UserID, s.Now(), false)
}

// ForgotPassword emails a reset code to email if it has an account. The answer is the same either way.
func (s *Service) ForgotPassword(ctx context.Context, rawEmail, ip string) (domain.AuthCodeResponse, error) {
	email, ok := NormEmail(rawEmail)
	if !ok {
		return domain.AuthCodeResponse{}, errBadEmail()
	}
	if s.Mailer == nil {
		return domain.AuthCodeResponse{}, errEmailUnavailable()
	}
	if err := s.allowSend(ip, s.Now()); err != nil {
		return domain.AuthCodeResponse{}, err
	}
	c, found, err := s.Store.CredentialsByEmail(ctx, email)
	if err != nil {
		return domain.AuthCodeResponse{}, err
	}
	return s.sendCode(ctx, email, PurposeReset, found && c.PasswordHash != "", true)
}

// CheckResetCode trades the right reset code for a reset token, good for ResetTokenTTL.
func (s *Service) CheckResetCode(ctx context.Context, rawEmail, code string) (domain.ResetCodeResponse, error) {
	email, ok := NormEmail(rawEmail)
	if !ok {
		return domain.ResetCodeResponse{}, errBadEmail()
	}
	if err := s.useCode(ctx, email, PurposeReset, code); err != nil {
		return domain.ResetCodeResponse{}, err
	}
	c, found, err := s.Store.CredentialsByEmail(ctx, email)
	if err != nil {
		return domain.ResetCodeResponse{}, err
	}
	if !found {
		return domain.ResetCodeResponse{}, errExpired()
	}
	now := s.Now()
	token := newToken()
	if err := s.Store.CreateResetToken(ctx, HashToken(token), c.UserID, now, now.Add(ResetTokenTTL)); err != nil {
		return domain.ResetCodeResponse{}, err
	}
	return domain.ResetCodeResponse{ResetToken: token, ExpiresInSeconds: int(ResetTokenTTL / time.Second)}, nil
}

// ResetPassword sets a new password with a reset token and signs this device in. Every other device is signed out,
// and the address counts as confirmed, since the code proved the inbox. A weak password doesn't use the token up.
func (s *Service) ResetPassword(ctx context.Context, resetToken, password string) (SignIn, error) {
	now := s.Now()
	c, ok, err := s.Store.ResetTokenUser(ctx, HashToken(resetToken), now)
	if err != nil {
		return SignIn{}, err
	}
	if !ok {
		return SignIn{}, errResetExpired()
	}
	if p := PasswordProblem(password, c.Email); p != "" {
		return SignIn{}, errWeakPassword(p)
	}
	hash, err := HashPassword(password, s.Iterations)
	if err != nil {
		return SignIn{}, err
	}
	userID, ok, err := s.Store.ResetPassword(ctx, HashToken(resetToken), hash, now)
	if err != nil {
		return SignIn{}, err
	}
	if !ok {
		return SignIn{}, errResetExpired() // a concurrent reset used it first
	}
	s.init()
	s.failAcct.forget("email:" + c.Email)
	return s.openSession(ctx, userID, now, false)
}

// CheckCurrentPassword confirms that the signed-in fan knows their password, before a change that needs it
// (deleting the account). A wrong one is 403 bad-password, never 401, which would sign the device out.
func (s *Service) CheckCurrentPassword(ctx context.Context, userID, password, ip string) error {
	c, found, err := s.Store.CredentialsByID(ctx, userID)
	if err != nil {
		return err
	}
	var account *store.Credentials
	if found {
		account = &c
	}
	right, err := s.checkPassword(ctx, account, "user:"+userID, password, ip)
	if err != nil {
		return err
	}
	if !right {
		return errBadPassword()
	}
	return nil
}

// ChangePassword replaces the signed-in fan's password after checking the current one, and signs out every other
// device. The session in keep (a token hash) stays signed in.
func (s *Service) ChangePassword(ctx context.Context, userID, keep, current, next, ip string) error {
	if err := s.CheckCurrentPassword(ctx, userID, current, ip); err != nil {
		return err
	}
	c, _, err := s.Store.CredentialsByID(ctx, userID)
	if err != nil {
		return err
	}
	if p := PasswordProblem(next, c.Email); p != "" {
		return errWeakPassword(p)
	}
	if next == current {
		return errWeakPassword("That's your current password. Choose a new one.")
	}
	hash, err := HashPassword(next, s.Iterations)
	if err != nil {
		return err
	}
	return s.Store.SetPassword(ctx, userID, hash, keep)
}
