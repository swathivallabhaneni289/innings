package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/auth"
	"github.com/swathivallabhaneni289/innings/internal/sim"
)

// errCode decodes {"error": ...} from a body.
func errCode(b []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	json.Unmarshal(b, &e)
	return e.Error
}

func (c *client) confirm(email, code string) (int, []byte) {
	c.t.Helper()
	return c.do("POST", "/v1/auth/confirm", map[string]string{"email": email, "code": code})
}

func (c *client) signIn(email, password string) (int, []byte) {
	c.t.Helper()
	return c.do("POST", "/v1/auth/signin", map[string]string{"email": email, "password": password})
}

func wrong(code string) string {
	if code == "000000" {
		return "111111"
	}
	return "000000"
}

func TestSignUpAndSecondDevice(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	email := nextEmail()

	// The app may send the address as typed; it comes back trimmed and lower-cased.
	status, b, _ := c.req("POST", "/v1/auth/signup", map[string]string{"email": "  " + strings.ToUpper(email) + " ", "password": testPassword})
	var sent struct {
		Email            string `json:"email"`
		ResendInSeconds  int    `json:"resendInSeconds"`
		ExpiresInSeconds int    `json:"expiresInSeconds"`
		DevCode          string `json:"devCode"`
	}
	json.Unmarshal(b, &sent)
	if status != 200 || sent.Email != email || sent.ResendInSeconds != 30 || sent.ExpiresInSeconds != 600 || len(sent.DevCode) != 6 {
		t.Fatalf("sign up: %d %s", status, b)
	}
	// Not confirmed yet: the right password isn't enough.
	if status, b := c.signIn(email, testPassword); status != http.StatusForbidden || errCode(b) != "email-unconfirmed" {
		t.Fatalf("sign in before confirming: %d %s", status, b)
	}
	status, b = c.confirm(email, sent.DevCode)
	if status != 200 {
		t.Fatalf("confirm: %d %s", status, b)
	}
	var first struct {
		Token   string         `json:"token"`
		Me      map[string]any `json:"me"`
		Account map[string]any `json:"account"`
		IsNew   bool           `json:"isNew"`
	}
	json.Unmarshal(b, &first)
	if !first.IsNew || len(first.Token) != 64 || first.Me["name"] != "" || first.Me["handle"] != "" ||
		first.Me["avatarColour"] != "#3B3F6B" || first.Me["favouriteTeamId"] != "ind" {
		t.Fatalf("new account: %s", b)
	}
	if first.Account["email"] != email || first.Account["onboarded"] != false || first.Account["createdAt"] == "" {
		t.Fatalf("account: %v", first.Account)
	}
	if _, ok := first.Me["email"]; ok {
		t.Fatal("email must never be part of User")
	}
	c.token, c.email, c.password = first.Token, email, testPassword

	// A new account starts empty: no logs, predictions or follows.
	me := c.getJSON("/v1/users/me")
	if len(list(me["logs"])) != 0 || me["stats"].(map[string]any)["logged"] != 0.0 {
		t.Fatalf("new account has history: %v", me["logs"])
	}
	for _, f := range list(c.getJSON("/v1/friends")["friends"]) {
		if f.(map[string]any)["following"] == true {
			t.Fatalf("new account follows %v", str(f, "user", "id"))
		}
	}
	// Someone else looking me up never sees my address.
	other := c.another()
	if _, b := other.do("GET", "/v1/users/"+str(first.Me, "id"), nil); strings.Contains(string(b), email) {
		t.Fatalf("email leaked in a profile: %s", b)
	}

	// Same email and password on a second device: same fan, new session, both work.
	second := c.fresh()
	got := second.signInWith(strings.ToUpper(email), testPassword)
	if got["isNew"] != false || second.id != str(first.Me, "id") || second.token == c.token {
		t.Fatalf("second device: %v", got)
	}
	if str(c.getJSON("/v1/users/me")["user"], "id") != second.id {
		t.Fatal("first device lost its session")
	}
	// A spent confirmation code doesn't sign anyone in again.
	if status, b := c.confirm(email, sent.DevCode); status != http.StatusGone || errCode(b) != "code-expired" {
		t.Fatalf("reused code: %d %s", status, b)
	}
}

func TestSignUpRules(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	signUp := func(email, password string) (int, []byte) {
		return c.do("POST", "/v1/auth/signup", map[string]string{"email": email, "password": password})
	}
	for _, bad := range []string{"", "swathi", "swathi@gmail", "a b@example.com"} {
		if status, b := signUp(bad, testPassword); status != 400 || errCode(b) != "bad-email" {
			t.Errorf("email %q: %d %s", bad, status, b)
		}
	}
	email := nextEmail()
	for pw, want := range map[string]string{"short": "at least 8", "password123": "first passwords", email: "email address"} {
		status, b := signUp(email, pw)
		if status != 400 || errCode(b) != "weak-password" || !strings.Contains(string(b), want) {
			t.Errorf("password %q: %d %s", pw, status, b)
		}
	}

	// An unconfirmed account is taken over by the next sign-up: the newest password is the one that works.
	first := c.askCode("/v1/auth/signup", map[string]string{"email": email, "password": "first-password-1"})
	c.clock.advance(auth.ResendGap)
	second := c.askCode("/v1/auth/signup", map[string]string{"email": email, "password": "second-password-2"})
	if status, _ := c.confirm(email, first); status != http.StatusBadRequest {
		t.Fatalf("a replaced code must not work: %d", status)
	}
	if status, b := c.confirm(email, second); status != 200 {
		t.Fatalf("confirm: %d %s", status, b)
	}
	if status, _ := c.signIn(email, "first-password-1"); status != http.StatusUnauthorized {
		t.Fatalf("the replaced password still works: %d", status)
	}
	c.fresh().signInWith(email, "second-password-2")

	// A confirmed address can't be signed up again.
	if status, b := signUp(email, "third-password-3"); status != http.StatusConflict || errCode(b) != "email-taken" {
		t.Fatalf("taken: %d %s", status, b)
	}
	c.fresh().signInWith(email, "second-password-2") // and its password is unchanged

	// Signing up again within the resend gap still answers 200, with the wait and no new code.
	pending := nextEmail()
	c.askCode("/v1/auth/signup", map[string]string{"email": pending, "password": testPassword})
	c.clock.advance(10 * time.Second)
	status, b := signUp(pending, testPassword)
	if status != 200 || !strings.Contains(string(b), `"resendInSeconds":20`) || strings.Contains(string(b), "devCode") {
		t.Fatalf("sign up again within the gap: %d %s", status, b)
	}
}

func TestSignInErrors(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	same := func(status int, b []byte) bool {
		return status == http.StatusUnauthorized && errCode(b) == "bad-credentials" && strings.Contains(string(b), "don't match")
	}
	if status, b := c.signIn(c.email, "not-my-password"); !same(status, b) {
		t.Fatalf("wrong password: %d %s", status, b)
	}
	if status, b := c.signIn(nextEmail(), testPassword); !same(status, b) {
		t.Fatalf("unknown address must look the same as a wrong password: %d %s", status, b)
	}
	if status, b := c.signIn("not-an-email", testPassword); status != 400 || errCode(b) != "bad-email" {
		t.Fatalf("bad email: %d %s", status, b)
	}
	// A wrong password on a signed-in device doesn't sign it out.
	c.getJSON("/v1/home")

	// An unconfirmed account with a wrong password says nothing about being unconfirmed.
	pending := nextEmail()
	c.askCode("/v1/auth/signup", map[string]string{"email": pending, "password": testPassword})
	if status, b := c.signIn(pending, "not-the-password"); !same(status, b) {
		t.Fatalf("unconfirmed, wrong password: %d %s", status, b)
	}
}

func TestWrongPasswordCap(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	// The right password wipes the slate, so nine wrong ones and a right one leave a fresh allowance.
	for range auth.FailsPerAccount - 1 {
		c.signIn(c.email, "wrong-password")
	}
	c.fresh().signInWith(c.email, testPassword)
	for i := range auth.FailsPerAccount {
		if status, b := c.signIn(c.email, "wrong-password"); status != http.StatusUnauthorized {
			t.Fatalf("wrong password %d: %d %s", i+1, status, b)
		}
	}
	// Ten wrong in a row: even the right one waits now.
	status, b, h := c.req("POST", "/v1/auth/signin", map[string]string{"email": c.email, "password": testPassword})
	if status != http.StatusTooManyRequests || errCode(b) != "too-many" || h.Get("Retry-After") != "900" || !strings.Contains(string(b), "15 minutes") {
		t.Fatalf("after %d wrong: %d %s Retry-After=%q", auth.FailsPerAccount, status, b, h.Get("Retry-After"))
	}
	c.another() // other accounts are unaffected
	c.clock.advance(auth.FailWindow)
	c.fresh().signInWith(c.email, testPassword)
}

func TestWrongCodeCountsDown(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	email := nextEmail()
	code := c.askCode("/v1/auth/signup", map[string]string{"email": email, "password": testPassword})
	for _, want := range []string{"4 tries left", "3 tries left", "2 tries left", "1 try left", "last try"} {
		status, b := c.confirm(email, wrong(code))
		if status != 400 || errCode(b) != "bad-code" || !strings.Contains(string(b), want) {
			t.Fatalf("want 400 bad-code %q, got %d %s", want, status, b)
		}
	}
	// Out of tries: even the right code is refused now.
	if status, b := c.confirm(email, code); status != http.StatusGone || errCode(b) != "code-expired" {
		t.Fatalf("after 5 wrong tries: %d %s", status, b)
	}
	// A malformed guess also spends a try.
	c.clock.advance(auth.ResendGap)
	code = c.askCode("/v1/auth/confirm/resend", map[string]string{"email": email})
	if status, b := c.confirm(email, "12ab"); status != 400 || !strings.Contains(string(b), "4 tries left") {
		t.Fatalf("malformed code: %d %s", status, b)
	}
	if status, b := c.confirm(email, code); status != 200 {
		t.Fatalf("right code after one wrong: %d %s", status, b)
	}
}

func TestCodeExpires(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	email := nextEmail()
	code := c.askCode("/v1/auth/signup", map[string]string{"email": email, "password": testPassword})
	c.clock.advance(auth.CodeTTL - time.Second)
	if status, b := c.confirm(email, wrong(code)); status != 400 {
		t.Fatalf("just before expiry: %d %s", status, b)
	}
	c.clock.advance(time.Second)
	if status, b := c.confirm(email, code); status != http.StatusGone || errCode(b) != "code-expired" {
		t.Fatalf("expired code: %d %s", status, b)
	}
	if status, b := c.confirm(nextEmail(), "123456"); status != http.StatusGone {
		t.Fatalf("never sent: %d %s", status, b)
	}
}

func TestResendConfirmation(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	email := nextEmail()
	c.askCode("/v1/auth/signup", map[string]string{"email": email, "password": testPassword})
	resend := func(email string) (int, []byte, http.Header) {
		return c.req("POST", "/v1/auth/confirm/resend", map[string]string{"email": email})
	}
	status, b, h := resend(email)
	if status != http.StatusTooManyRequests || errCode(b) != "too-soon" || h.Get("Retry-After") != "30" {
		t.Fatalf("immediate resend: %d %s Retry-After=%q", status, b, h.Get("Retry-After"))
	}
	c.clock.advance(10 * time.Second)
	if _, _, h := resend(email); h.Get("Retry-After") != "20" {
		t.Fatalf("Retry-After after 10 s = %q", h.Get("Retry-After"))
	}
	c.clock.advance(20 * time.Second)
	code := c.askCode("/v1/auth/confirm/resend", map[string]string{"email": email}) // a new code replaces the old one
	if status, b := c.confirm(email, code); status != 200 {
		t.Fatalf("new code: %d %s", status, b)
	}
	// Confirmed now, and an address with no account at all: the answer has the same shape, and nothing is sent.
	c.clock.advance(auth.ResendGap)
	for _, e := range []string{email, nextEmail()} {
		status, b, _ := resend(e)
		if status != 200 || strings.Contains(string(b), "devCode") || !strings.Contains(string(b), `"resendInSeconds":30`) {
			t.Fatalf("resend for %s: %d %s", e, status, b)
		}
	}
}

func TestHourlyCapPerAddress(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	email := nextEmail()
	c.askCode("/v1/auth/signup", map[string]string{"email": email, "password": testPassword})
	for range auth.SendsPerHour - 1 {
		c.clock.advance(auth.ResendGap)
		c.askCode("/v1/auth/confirm/resend", map[string]string{"email": email})
	}
	c.clock.advance(auth.ResendGap)
	status, b, h := c.req("POST", "/v1/auth/password/forgot", map[string]string{"email": email})
	// The first of the five went out 150 s ago, so it ages out in 3450 s. Every purpose counts.
	if status != http.StatusTooManyRequests || errCode(b) != "too-many" || h.Get("Retry-After") != "3450" {
		t.Fatalf("sixth code this hour: %d %s Retry-After=%q", status, b, h.Get("Retry-After"))
	}
	c.askCode("/v1/auth/signup", map[string]string{"email": nextEmail(), "password": testPassword}) // another address is unaffected
	c.clock.advance(58 * time.Minute)
	c.askCode("/v1/auth/confirm/resend", map[string]string{"email": email})
}

func TestIPCap(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	for range auth.SendsPerIPPerHour {
		c.askCode("/v1/auth/signup", map[string]string{"email": nextEmail(), "password": testPassword})
	}
	for _, xff := range []string{"", "198.51.100.7"} {
		req, _ := http.NewRequest("POST", c.base+"/v1/auth/password/forgot", strings.NewReader(`{"email":"`+nextEmail()+`"}`))
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff) // must not buy a fresh allowance
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var b [256]byte
		n, _ := res.Body.Read(b[:])
		res.Body.Close()
		if res.StatusCode != http.StatusTooManyRequests || errCode(b[:n]) != "too-many" || res.Header.Get("Retry-After") == "" {
			t.Fatalf("21st code from one IP (XFF %q): %d %s", xff, res.StatusCode, b[:n])
		}
	}
	c.clock.advance(time.Hour)
	c.askCode("/v1/auth/signup", map[string]string{"email": nextEmail(), "password": testPassword})
}

func TestNoMailer(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	c.api.Auth.Mailer = nil
	for path, body := range map[string]map[string]string{
		"/v1/auth/signup":          {"email": nextEmail(), "password": testPassword},
		"/v1/auth/confirm/resend":  {"email": nextEmail()},
		"/v1/auth/password/forgot": {"email": nextEmail()},
	} {
		if status, b := c.do("POST", path, body); status != http.StatusServiceUnavailable || errCode(b) != "email-unavailable" {
			t.Errorf("%s with no mailer: %d %s", path, status, b)
		}
	}
}

// resetCode asks for a reset code for email and trades it for a reset token.
func (c *client) resetToken(email string) string {
	c.t.Helper()
	code := c.askCode("/v1/auth/password/forgot", map[string]string{"email": email})
	status, b := c.do("POST", "/v1/auth/password/code", map[string]string{"email": email, "code": code})
	var out struct {
		ResetToken       string `json:"resetToken"`
		ExpiresInSeconds int    `json:"expiresInSeconds"`
	}
	json.Unmarshal(b, &out)
	if status != 200 || len(out.ResetToken) != 64 || out.ExpiresInSeconds != 900 {
		c.t.Fatalf("reset code: %d %s", status, b)
	}
	return out.ResetToken
}

func TestForgotPassword(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	other := c.device()
	c.clock.advance(auth.ResendGap)
	token := c.resetToken(c.email)
	reset := func(token, password string) (int, []byte) {
		return c.do("POST", "/v1/auth/password/reset", map[string]string{"resetToken": token, "password": password})
	}

	// A weak password is refused without using the token up.
	if status, b := reset(token, "short"); status != 400 || errCode(b) != "weak-password" {
		t.Fatalf("weak: %d %s", status, b)
	}
	status, b := reset(token, "brand-new-password")
	if status != 200 {
		t.Fatalf("reset: %d %s", status, b)
	}
	fresh := c.fresh()
	if got := fresh.keepSession(b); got["isNew"] != false || fresh.id != c.id {
		t.Fatalf("signed in after reset: %v", got)
	}
	// Every other device is signed out; the old password is dead, the new one works.
	for _, d := range []*client{c, other} {
		if status, _ := d.do("GET", "/v1/home", nil); status != http.StatusUnauthorized {
			t.Fatalf("a device stayed signed in after a reset: %d", status)
		}
	}
	if status, _ := c.signIn(c.email, testPassword); status != http.StatusUnauthorized {
		t.Fatalf("old password: %d", status)
	}
	c.fresh().signInWith(c.email, "brand-new-password")
	if status, b := reset(token, "another-new-password"); status != http.StatusGone || errCode(b) != "reset-expired" {
		t.Fatalf("token used twice: %d %s", status, b)
	}

	// An address with no account gets the same answer and no code; a guessed code for it goes nowhere.
	nobody := nextEmail()
	status, b, _ = c.req("POST", "/v1/auth/password/forgot", map[string]string{"email": nobody})
	if status != 200 || strings.Contains(string(b), "devCode") || !strings.Contains(string(b), `"expiresInSeconds":600`) {
		t.Fatalf("forgot for nobody: %d %s", status, b)
	}
	if status, b := c.do("POST", "/v1/auth/password/code", map[string]string{"email": nobody, "code": "123456"}); status != 400 || errCode(b) != "bad-code" {
		t.Fatalf("code for nobody: %d %s", status, b)
	}

	// A reset token runs out after 15 minutes.
	c.clock.advance(auth.ResendGap)
	late := c.resetToken(c.email)
	c.clock.advance(auth.ResetTokenTTL)
	if status, b := reset(late, "too-late-password"); status != http.StatusGone || errCode(b) != "reset-expired" {
		t.Fatalf("expired token: %d %s", status, b)
	}
	if status, b := reset("not-a-token", "whatever-password"); status != http.StatusGone {
		t.Fatalf("unknown token: %d %s", status, b)
	}
}

// TestForgotPasswordConfirms: someone who never entered the sign-up code can reset instead, which proves the inbox.
func TestForgotPasswordConfirms(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	email := nextEmail()
	c.askCode("/v1/auth/signup", map[string]string{"email": email, "password": testPassword})
	c.clock.advance(auth.ResendGap)
	token := c.resetToken(email)
	status, b := c.do("POST", "/v1/auth/password/reset", map[string]string{"resetToken": token, "password": "after-reset-password"})
	if status != 200 {
		t.Fatalf("reset: %d %s", status, b)
	}
	c.fresh().signInWith(email, "after-reset-password")
}

func TestChangePassword(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	other := c.device()
	change := func(current, next string) (int, []byte) {
		return c.do("PUT", "/v1/me/password", map[string]string{"currentPassword": current, "newPassword": next})
	}
	// A wrong current password is 403, not 401: it must never sign this device out.
	if status, b := change("not-my-password", "next-password-1"); status != http.StatusForbidden || errCode(b) != "bad-password" {
		t.Fatalf("wrong current: %d %s", status, b)
	}
	c.getJSON("/v1/home")
	if status, b := change(testPassword, "short"); status != 400 || errCode(b) != "weak-password" {
		t.Fatalf("weak: %d %s", status, b)
	}
	if status, b := change(testPassword, testPassword); status != 400 || errCode(b) != "weak-password" {
		t.Fatalf("unchanged: %d %s", status, b)
	}
	if status, b := change(testPassword, "next-password-1"); status != http.StatusNoContent {
		t.Fatalf("change: %d %s", status, b)
	}
	c.getJSON("/v1/home") // this device stays in
	if status, _ := other.do("GET", "/v1/home", nil); status != http.StatusUnauthorized {
		t.Fatalf("the other device stayed signed in: %d", status)
	}
	if status, _ := c.signIn(c.email, testPassword); status != http.StatusUnauthorized {
		t.Fatalf("old password: %d", status)
	}
	c.fresh().signInWith(c.email, "next-password-1")
	if status, _ := c.fresh().do("PUT", "/v1/me/password", map[string]string{"currentPassword": "x", "newPassword": "y"}); status != http.StatusUnauthorized {
		t.Fatalf("no session: %d", status)
	}
}

func TestLogoutEndsOnlyThisSession(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	other := c.device()
	if status, _ := c.do("POST", "/v1/auth/logout", nil); status != http.StatusNoContent {
		t.Fatalf("logout: %d", status)
	}
	if status, _ := c.do("GET", "/v1/home", nil); status != http.StatusUnauthorized {
		t.Fatalf("signed-out token still works: %d", status)
	}
	if status, _ := c.do("POST", "/v1/auth/logout", nil); status != http.StatusUnauthorized {
		t.Fatalf("logout twice: %d", status)
	}
	other.getJSON("/v1/home") // the same fan's other device is still signed in
}

func TestHandles(t *testing.T) { eachWorld(t, testHandles) }

func testHandles(t *testing.T, c *client) {
	check := func(h, reason string) {
		t.Helper()
		r := c.getJSON("/v1/handles/"+h, "handle", "available", "reason", "message")
		if reason == "" && (r["available"] != true || r["reason"] != nil || r["message"] != nil) {
			t.Errorf("%s: want available, got %v", h, r)
		}
		if reason != "" && (r["available"] != false || r["reason"] != reason || r["message"] == nil) {
			t.Errorf("%s: want %s, got %v", h, reason, r)
		}
	}
	if r := c.getJSON("/v1/handles/Swathi.R"); r["handle"] != "swathi.r" || r["available"] != true {
		t.Fatalf("lower-cased check: %v", r)
	}
	for _, h := range []string{"ab", "1abc", "abc.", "ab-c", "abcdefghijklmnopqrstu"} {
		check(h, "invalid")
	}
	for _, h := range []string{"innings", "Admin", "support", "help", "settings", "official", "team"} {
		check(h, "reserved")
	}
	check("me", "invalid") // too short as well as reserved
	friends := list(c.getJSON("/v1/friends")["friends"])
	if len(friends) == 0 {
		t.Fatal("world has no simulated friends")
	}
	simHandle := str(friends[0], "user", "handle")
	check(strings.ToUpper(simHandle), "taken")

	if status, b := c.do("PUT", "/v1/me", map[string]any{"handle": "Swathi.R"}); status != 200 || !strings.Contains(string(b), `"handle":"swathi.r"`) {
		t.Fatalf("set handle: %d %s", status, b)
	}
	check("SWATHI.R", "") // my own handle counts as available
	if status, b := c.do("PUT", "/v1/me", map[string]any{"handle": "swathi.r"}); status != 200 {
		t.Fatalf("re-save own handle: %d %s", status, b)
	}

	other := c.another()
	if r := other.getJSON("/v1/handles/swathi.r"); r["reason"] != "taken" {
		t.Fatalf("someone else's handle: %v", r)
	}
	for h, want := range map[string]string{"Swathi.R": "handle-taken", simHandle: "handle-taken", "admin": "bad-handle", "x": "bad-handle", "": "bad-handle"} {
		code := map[string]int{"handle-taken": 409, "bad-handle": 400}[want]
		if status, b := other.do("PUT", "/v1/me", map[string]any{"handle": h}); status != code || errCode(b) != want {
			t.Errorf("PUT handle %q: %d %s", h, status, b)
		}
	}
	if u := other.getJSON("/v1/users/me")["user"]; str(u, "handle") != "" {
		t.Fatalf("a refused handle was saved: %v", u)
	}
}

// onboard saves a name and handle and finishes sign-up.
func (c *client) onboard(name, handle string) {
	c.t.Helper()
	if status, b := c.do("PUT", "/v1/me", map[string]any{"name": name, "handle": handle}); status != 200 {
		c.t.Fatalf("profile: %d %s", status, b)
	}
	if status, b := c.do("POST", "/v1/me/onboarded", nil); status != 200 {
		c.t.Fatalf("onboarded: %d %s", status, b)
	}
}

func TestOnboarding(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	if status, b := c.do("POST", "/v1/me/onboarded", nil); status != 400 || errCode(b) != "profile-incomplete" {
		t.Fatalf("blank profile: %d %s", status, b)
	}
	c.do("PUT", "/v1/me", map[string]any{"name": "Swathi"})
	if status, b := c.do("POST", "/v1/me/onboarded", nil); status != 400 || errCode(b) != "profile-incomplete" {
		t.Fatalf("no handle yet: %d %s", status, b)
	}
	c.do("PUT", "/v1/me", map[string]any{"handle": "swathi"})
	status, b := c.do("POST", "/v1/me/onboarded", nil)
	var out struct {
		Account map[string]any `json:"account"`
	}
	json.Unmarshal(b, &out)
	if status != 200 || out.Account["onboarded"] != true || out.Account["email"] != c.email {
		t.Fatalf("onboarded: %d %s", status, b)
	}
	if _, boot := c.bootstrapStatus(c.token); boot["account"].(map[string]any)["onboarded"] != true {
		t.Fatalf("bootstrap account: %v", boot["account"])
	}
}

func TestNotOnboardedIsHidden(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	shown, hidden := c.another(), c.another()
	shown.onboard("Zebedee Shown", "zshown")
	hidden.do("PUT", "/v1/me", map[string]any{"name": "Zebedee Hidden", "handle": "zhidden"}) // never finishes sign-up

	users := list(c.getJSON("/v1/search?q=zebedee")["users"])
	if len(users) != 1 || str(users[0], "id") != shown.id {
		t.Fatalf("search: %v", users)
	}
	c.do("POST", "/v1/users/"+shown.id+"/follow", nil)
	c.do("POST", "/v1/users/"+hidden.id+"/follow", nil)
	ids := map[string]bool{}
	for _, f := range list(c.getJSON("/v1/friends")["friends"]) {
		ids[str(f, "user", "id")] = true
	}
	if !ids[shown.id] || ids[hidden.id] {
		t.Fatalf("friends list: shown=%v hidden=%v", ids[shown.id], ids[hidden.id])
	}
}

func emailHash(email string) string {
	sum := sha256.Sum256([]byte(email))
	return hex.EncodeToString(sum[:])
}

func TestContactsMatch(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	c.onboard("Me", "me_myself")
	friend, halfway := c.another(), c.another()
	friend.onboard("Aarav", "aarav")
	match := func(hashes []string) (int, []byte) {
		return c.do("POST", "/v1/contacts/match", map[string]any{"hashes": hashes})
	}
	hashes := []string{emailHash(friend.email), emailHash(halfway.email), emailHash(c.email), emailHash("stranger@example.com"), emailHash(friend.email)}
	status, b := match(hashes)
	var out struct {
		Matches []struct {
			User      map[string]any `json:"user"`
			Following bool           `json:"following"`
		} `json:"matches"`
	}
	json.Unmarshal(b, &out)
	if status != 200 || len(out.Matches) != 1 || out.Matches[0].User["id"] != friend.id || out.Matches[0].Following {
		t.Fatalf("match: %d %s", status, b)
	}
	if strings.Contains(string(b), friend.email) {
		t.Fatalf("email in contact match: %s", b)
	}
	c.do("POST", "/v1/users/"+friend.id+"/follow", nil)
	if _, b := match(hashes[:1]); !strings.Contains(string(b), `"following":true`) {
		t.Fatalf("following after follow: %s", b)
	}
	if status, b := match([]string{}); status != 200 || !strings.Contains(string(b), `"matches":[]`) {
		t.Fatalf("no hashes: %d %s", status, b)
	}
	tooMany := make([]string, 5001)
	for i := range tooMany {
		tooMany[i] = emailHash(fmt.Sprint(i))
	}
	for _, bad := range [][]string{{strings.ToUpper(emailHash(friend.email))}, {"abc"}, {friend.email}, tooMany} {
		if status, b := match(bad); status != 400 || errCode(b) != "bad-hashes" {
			t.Errorf("bad hashes (%d): %d %s", len(bad), status, b)
		}
	}
	if status, b := match(tooMany[:5000]); status != 200 {
		t.Fatalf("5000 hashes: %d %s", status, b)
	}
}

func TestDeleteAccount(t *testing.T) {
	gone := newTestAPI(t, newFakeWorld(testNow))
	gone.onboard("Leaving Soon", "leaving")
	stays := gone.another()
	stays.onboard("Staying Put", "staying")
	goneDevice2 := gone.device()

	newLog := func(c *client, matchID string) string {
		status, b := c.do("POST", "/v1/logs", map[string]any{"matchId": matchID, "rating": 4, "watchedAt": "home", "review": "Good one"})
		if status != 201 {
			t.Fatalf("log: %d %s", status, b)
		}
		var fi struct {
			Log struct {
				ID string `json:"id"`
			} `json:"log"`
		}
		json.Unmarshal(b, &fi)
		return fi.Log.ID
	}
	goneLog, staysLog := newLog(gone, "done"), newLog(stays, "old")
	gone.do("POST", "/v1/users/"+stays.id+"/follow", nil)
	stays.do("POST", "/v1/users/"+gone.id+"/follow", nil)
	for _, pair := range []struct {
		c   *client
		log string
	}{{gone, staysLog}, {stays, goneLog}} {
		pair.c.do("POST", "/v1/logs/"+pair.log+"/like", nil)
		pair.c.do("POST", "/v1/logs/"+pair.log+"/there-too", nil)
		if status, b := pair.c.do("POST", "/v1/logs/"+pair.log+"/replies", map[string]string{"text": "Was there!"}); status != 201 {
			t.Fatalf("reply: %d %s", status, b)
		}
	}
	gone.do("PUT", "/v1/matches/soon/prediction", map[string]string{"winnerId": "mi", "topRunScorerId": "mi-p0", "topWicketTakerId": "csk-p1"})
	gone.do("PUT", "/v1/matches/live/watching", map[string]any{"place": "cafe"})
	gone.do("PUT", "/v1/me/top5", map[string]any{"matchIds": []string{"done"}})
	// Before: stays sees gone everywhere.
	if fi := stays.getJSON("/v1/logs/" + staysLog); fi["likes"] != 1.0 || fi["thereToo"] != 1.0 || len(list(fi["replies"])) != 1 {
		t.Fatalf("before delete: %v", fi)
	}
	if !strings.Contains(fmt.Sprint(stays.getJSON("/v1/notifications")["items"]), "Leaving Soon") {
		t.Fatal("before delete: no notification from the leaving fan")
	}

	// The password is asked for again, and a wrong one keeps the account (and doesn't sign the device out).
	if status, b := gone.do("DELETE", "/v1/me", map[string]string{"password": "not-my-password"}); status != http.StatusForbidden || errCode(b) != "bad-password" {
		t.Fatalf("delete with a wrong password: %d %s", status, b)
	}
	if status, b := gone.do("DELETE", "/v1/me", nil); status != http.StatusBadRequest {
		t.Fatalf("delete with no body: %d %s", status, b)
	}
	gone.getJSON("/v1/home")
	if status, b := gone.do("DELETE", "/v1/me", map[string]string{"password": testPassword}); status != http.StatusNoContent {
		t.Fatalf("delete: %d %s", status, b)
	}
	for _, d := range []*client{gone, goneDevice2} {
		if status, _ := d.do("GET", "/v1/home", nil); status != http.StatusUnauthorized {
			t.Fatalf("deleted account's session still works: %d", status)
		}
	}

	// Everything stays's screens load, with no trace of the deleted fan.
	home := stays.getJSON("/v1/home", "feed", "board")
	for _, row := range list(home["board"]) {
		if str(row, "user", "id") == gone.id {
			t.Error("deleted fan still on the board")
		}
	}
	for _, fi := range list(home["feed"]) {
		if str(fi, "user", "id") == gone.id {
			t.Error("deleted fan still in the feed")
		}
	}
	if notes := stays.getJSON("/v1/notifications", "items"); strings.Contains(fmt.Sprint(notes["items"]), "Leaving Soon") {
		t.Errorf("notifications still mention the deleted fan: %v", notes["items"])
	}
	fi := stays.getJSON("/v1/logs/"+staysLog, "log", "replies")
	if fi["likes"] != 0.0 || fi["thereToo"] != 0.0 || len(list(fi["replies"])) != 0 {
		t.Fatalf("stays's stub after delete: %v", fi)
	}
	if status, _ := stays.do("GET", "/v1/logs/"+goneLog, nil); status != http.StatusNotFound {
		t.Fatalf("deleted fan's log: %d", status)
	}
	if status, _ := stays.do("GET", "/v1/users/"+gone.id, nil); status != http.StatusNotFound {
		t.Fatalf("deleted fan's profile: %d", status)
	}
	for _, f := range list(stays.getJSON("/v1/friends")["friends"]) {
		if str(f, "user", "id") == gone.id {
			t.Error("deleted fan still in friends")
		}
	}
	if users := list(stays.getJSON("/v1/search?q=leaving")["users"]); len(users) != 0 {
		t.Errorf("deleted fan still in search: %v", users)
	}
	stays.getJSON("/v1/users/me", "stats", "logs")
	stays.getJSON("/v1/matches/done", "fans")
	if sp := stays.getJSON("/v1/matches/soon")["split"].(map[string]any); sp["total"] != 0.0 {
		t.Errorf("deleted fan's prediction still counted: %v", sp)
	}
	if w := stays.getJSON("/v1/matches/live")["friendsWatching"]; strings.Contains(fmt.Sprint(w), gone.id) {
		t.Errorf("deleted fan still watching: %v", w)
	}

	// The address is free again: signing up makes a brand-new account.
	again := gone.fresh()
	if got := again.signUp(gone.email); got["isNew"] != true || again.id == gone.id {
		t.Fatalf("re-sign-up: %v", got)
	}
}

func TestTeams(t *testing.T) {
	c := newTestAPI(t, sim.NewWorld(testAnchor)) // the whole team catalog
	teams := func(u any) []string {
		out := []string{}
		for _, id := range list(u.(map[string]any)["teamIds"]) {
			out = append(out, id.(string))
		}
		return out
	}
	me := c.getJSON("/v1/users/me")["user"]
	if got := teams(me); len(got) != 1 || got[0] != "ind" || str(me, "favouriteTeamId") != "ind" {
		t.Fatalf("a new fan backs India: %v", me)
	}
	put := func(body map[string]any) (int, []byte) { return c.do("PUT", "/v1/me", body) }

	// Several teams, main team first.
	status, b := put(map[string]any{"teamIds": []string{"rcb", "ind", "mi"}})
	if status != 200 || !strings.Contains(string(b), `"favouriteTeamId":"rcb","teamIds":["rcb","ind","mi"]`) {
		t.Fatalf("set teams: %d %s", status, b)
	}
	// Another fan sees them too.
	if u := c.another().getJSON("/v1/users/" + c.id)["user"]; strings.Join(teams(u), ",") != "rcb,ind,mi" {
		t.Fatalf("seen by others: %v", u)
	}
	// favouriteTeamId alone moves that team to the front and keeps the rest.
	if status, b := put(map[string]any{"favouriteTeamId": "mi"}); status != 200 || !strings.Contains(string(b), `"teamIds":["mi","rcb","ind"]`) {
		t.Fatalf("main team only: %d %s", status, b)
	}
	for name, ids := range map[string][]string{
		"none":     {},
		"unknown":  {"ind", "nope"},
		"repeated": {"ind", "rcb", "ind"},
		"six":      {"ind", "aus", "eng", "sa", "nz", "pak"},
	} {
		if status, b := put(map[string]any{"teamIds": ids}); status != 400 || errCode(b) != "bad-team" {
			t.Errorf("%s: %d %s", name, status, b)
		}
	}
	if u := c.getJSON("/v1/users/me")["user"]; strings.Join(teams(u), ",") != "mi,rcb,ind" {
		t.Fatalf("a refused change was saved: %v", u)
	}
	// Simulated fans, made with one favourite team, still send a list.
	for _, f := range list(c.getJSON("/v1/friends")["friends"]) {
		if u := f.(map[string]any)["user"]; len(teams(u)) != 1 || teams(u)[0] != str(u, "favouriteTeamId") {
			t.Fatalf("sim fan teams: %v", u)
		}
	}
}
