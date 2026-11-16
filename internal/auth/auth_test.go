package auth

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestBearerToken(t *testing.T) {
	for h, want := range map[string]string{"": "", "Bearer abc": "abc", "bearer  xyz ": "xyz", "Basic abc": "", "Bearer": ""} {
		r := httptest.NewRequest("GET", "/", nil)
		if h != "" {
			r.Header.Set("Authorization", h)
		}
		if got := BearerToken(r); got != want {
			t.Errorf("%q: got %q want %q", h, got, want)
		}
	}
}

func TestTokensAndCodes(t *testing.T) {
	tok := newToken()
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(tok) || tok == newToken() {
		t.Fatalf("token %q", tok)
	}
	if h := HashToken(tok); len(h) != 64 || h == tok {
		t.Fatalf("hash %q", h)
	}
	seen := map[string]bool{}
	for range 200 {
		c := newCode()
		if !codeRE.MatchString(c) {
			t.Fatalf("code %q", c)
		}
		seen[c] = true
	}
	if len(seen) < 190 {
		t.Fatalf("codes repeat too often: %d distinct of 200", len(seen))
	}
}

func TestNormEmail(t *testing.T) {
	for raw, want := range map[string]string{
		"swathi@gmail.com":                      "swathi@gmail.com",
		"  Swathi.V@Gmail.COM ":                 "swathi.v@gmail.com",
		"fan+innings@example.co.uk":             "fan+innings@example.co.uk",
		"a_b-c@mail.example-site.org":           "a_b-c@mail.example-site.org",
		"o'brien@example.ie":                    "o'brien@example.ie",
		"x@" + strings.Repeat("a", 60) + ".com": "x@" + strings.Repeat("a", 60) + ".com",
	} {
		if got, ok := NormEmail(raw); !ok || got != want {
			t.Errorf("NormEmail(%q) = %q %v", raw, got, ok)
		}
	}
	for _, raw := range []string{
		"", "swathi", "swathi@", "@gmail.com", "swathi@gmail", "swathi@@gmail.com", "a@b@c.com", "swathi @gmail.com",
		"swathi@gmail..com", "swathi@.gmail.com", "swathi@gmail.com.", "swathi@-gmail.com", "Swathi <s@gmail.com>",
		"s@gmail.com\r\nBcc: x@y.com", "s;t@gmail.com", strings.Repeat("a", 65) + "@gmail.com",
		"a@" + strings.Repeat("b", 250) + ".com",
	} {
		if got, ok := NormEmail(raw); ok {
			t.Errorf("NormEmail(%q) accepted as %q", raw, got)
		}
	}
}

func TestPasswordHashing(t *testing.T) {
	h, err := HashPassword("correct horse battery", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^pbkdf2-sha256\$1000\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`).MatchString(h) {
		t.Fatalf("stored form %q", h)
	}
	if h2, _ := HashPassword("correct horse battery", 1000); h2 == h {
		t.Fatal("two hashes of one password must differ (salt)")
	}
	if ok, rehash := CheckPassword(h, "correct horse battery", 1000); !ok || rehash {
		t.Fatalf("right password: %v %v", ok, rehash)
	}
	if ok, rehash := CheckPassword(h, "correct horse battery", 2000); !ok || !rehash {
		t.Fatalf("a stronger work factor asks for a rehash: %v %v", ok, rehash)
	}
	for _, wrong := range []string{"", "correct horse batter", "Correct horse battery", "correct horse battery "} {
		if ok, _ := CheckPassword(h, wrong, 1000); ok {
			t.Errorf("wrong password %q accepted", wrong)
		}
	}
	for _, bad := range []string{"", "plain", "bcrypt$1000$abc$def", "pbkdf2-sha256$x$abc$def", "pbkdf2-sha256$0$abc$def",
		"pbkdf2-sha256$1000$$def", "pbkdf2-sha256$1000$abc", "pbkdf2-sha256$1000$!!$def"} {
		if ok, _ := CheckPassword(bad, "plain", 1000); ok {
			t.Errorf("unreadable hash %q matched", bad)
		}
	}
	if DefaultIterations < 600_000 {
		t.Fatalf("work factor %d is below OWASP's 600,000 for PBKDF2-HMAC-SHA256", DefaultIterations)
	}
}

func TestPasswordProblem(t *testing.T) {
	const email = "swathi.v@gmail.com"
	for _, ok := range []string{"tests-are-fun", "Kohli scored 82!", "correct horse battery staple", "उत्तम-पासवर्ड1", strings.Repeat("x", 120) + "y"} {
		if p := PasswordProblem(ok, email); p != "" {
			t.Errorf("%q refused: %s", ok, p)
		}
	}
	for pw, want := range map[string]string{
		"short1":                           "at least 8",
		"सात सात":                          "at least 8", // 7 characters, though 19 bytes
		strings.Repeat("ab", 65):           "at most 128",
		"password":                         "first passwords",
		"Password123":                      "first passwords",
		"CRICKET123":                       "first passwords",
		"aaaaaaaaaa":                       "first passwords",
		"abcdefghij":                       "first passwords",
		"qwertyuiop":                       "first passwords",
		"98765432":                         "first passwords",
		email:                              "email address",
		strings.ToUpper(email):             "email address",
		"swathi.v":                         "email address",
		"\xff\xfe\xfd\xfc\xfb\xfa\xf9\xf8": "can't read",
	} {
		if got := PasswordProblem(pw, email); !strings.Contains(got, want) {
			t.Errorf("PasswordProblem(%q) = %q, want it to mention %q", pw, got, want)
		}
	}
}

func TestMailerFromEnv(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	env := func(kv map[string]string) func(string) string { return func(k string) string { return kv[k] } }
	smtpEnv := map[string]string{"SMTP_HOST": "smtp.gmail.com", "SMTP_USERNAME": "fan@gmail.com", "SMTP_PASSWORD": "app-pass"}

	m, mode, err := MailerFromEnv(env(smtpEnv), logger)
	sm, ok := m.(*SMTPMailer)
	if err != nil || mode != "smtp" || !ok || sm.Port != "587" || sm.From.String() != `"Innings" <fan@gmail.com>` {
		t.Fatalf("smtp: %+v %q %v", m, mode, err)
	}
	withDev := map[string]string{"INNINGS_DEV_CODES": "1", "SMTP_PORT": "465", "MAIL_FROM": "Cricket Diary <diary@example.com>"}
	for k, v := range smtpEnv {
		withDev[k] = v
	}
	if m, mode, _ := MailerFromEnv(env(withDev), logger); mode != "smtp" {
		t.Fatalf("SMTP must win over dev mode: %T %q", m, mode)
	} else if sm := m.(*SMTPMailer); sm.Port != "465" || sm.From.Address != "diary@example.com" || sm.From.Name != "Cricket Diary" {
		t.Fatalf("port and from: %+v", sm)
	}
	if m, mode, err := MailerFromEnv(env(map[string]string{"INNINGS_DEV_CODES": "1"}), logger); err != nil || mode != "dev" {
		t.Fatalf("dev: %T %q %v", m, mode, err)
	} else if _, ok := m.(devCodes); !ok {
		t.Fatal("the dev mailer must reveal codes")
	}
	if m, mode, err := MailerFromEnv(env(nil), logger); err != nil || mode != "off" || m != nil {
		t.Fatalf("off: %T %q %v", m, mode, err)
	}
	if _, _, err := MailerFromEnv(env(map[string]string{"SMTP_HOST": "smtp.gmail.com", "INNINGS_DEV_CODES": "1"}), logger); err == nil {
		t.Fatal("partial SMTP settings must be an error, not dev mode")
	}
	bad := map[string]string{"MAIL_FROM": "not an address"}
	for k, v := range smtpEnv {
		bad[k] = v
	}
	if _, _, err := MailerFromEnv(env(bad), logger); err == nil {
		t.Fatal("a bad MAIL_FROM must be an error")
	}
	var real Mailer = &SMTPMailer{}
	if _, ok := real.(devCodes); ok {
		t.Fatal("the SMTP mailer must never reveal codes")
	}
}

func TestCodeEmail(t *testing.T) {
	v := codeEmail("fan@example.com", PurposeVerify, "042917")
	if v.To != "fan@example.com" || !strings.HasPrefix(v.Subject, "042917 ") || !strings.Contains(v.Text, "code is 042917") ||
		!strings.Contains(v.Text, "confirm this email") || !strings.Contains(v.Text, "10 minutes") {
		t.Fatalf("verify email: %+v", v)
	}
	r := codeEmail("fan@example.com", PurposeReset, "042917")
	if !strings.Contains(r.Text, "new password") || !strings.Contains(r.Text, "password stays the same") {
		t.Fatalf("reset email: %+v", r)
	}
}

func TestWindowLimiter(t *testing.T) {
	l := newWindowLimiter(3, time.Hour)
	t0 := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	for i := range 3 {
		if ok, _ := l.allow("a", t0.Add(time.Duration(i)*time.Minute)); !ok {
			t.Fatalf("hit %d refused", i)
		}
	}
	if ok, wait := l.allow("a", t0.Add(10*time.Minute)); ok || wait != 50*time.Minute {
		t.Fatalf("4th: %v %v", ok, wait)
	}
	if ok, _ := l.allow("b", t0.Add(10*time.Minute)); !ok {
		t.Fatal("keys must be separate")
	}
	if ok, _ := l.allow("a", t0.Add(time.Hour+time.Second)); !ok {
		t.Fatal("oldest hit should have aged out")
	}
	// Stale keys are swept once enough pile up.
	for i := range sweepAt + 1 {
		l.allow(time.Duration(i).String(), t0)
	}
	l.allow("late", t0.Add(2*time.Hour))
	if len(l.hits) > 2 {
		t.Fatalf("stale keys kept: %d", len(l.hits))
	}
}

func TestFailureCounting(t *testing.T) {
	l := newWindowLimiter(2, 15*time.Minute)
	t0 := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	if l.wait("k", t0) != 0 {
		t.Fatal("fresh key must be free")
	}
	l.record("k", t0)
	if l.wait("k", t0.Add(time.Minute)) != 0 {
		t.Fatal("one failure of two must still be free")
	}
	l.record("k", t0.Add(5*time.Minute))
	if w := l.wait("k", t0.Add(6*time.Minute)); w != 9*time.Minute {
		t.Fatalf("wait after two failures = %v", w)
	}
	// Recording past the cap keeps counting: the wait is until enough of them age out.
	l.record("k", t0.Add(6*time.Minute))
	if w := l.wait("k", t0.Add(6*time.Minute)); w != 14*time.Minute {
		t.Fatalf("wait after three failures = %v", w)
	}
	l.forget("k")
	if l.wait("k", t0.Add(6*time.Minute)) != 0 {
		t.Fatal("forget must clear the key")
	}
}

func TestClientIPAndKey(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "203.0.113.7:51234"
	r.Header.Set("X-Forwarded-For", "198.51.100.1")
	if ip := ClientIP(r); ip != "203.0.113.7" {
		t.Fatalf("ClientIP = %q (X-Forwarded-For must be ignored)", ip)
	}
	r.RemoteAddr = "[2001:db8:1:2:aaaa::1]:443"
	if k := ipKey(ClientIP(r)); k != "2001:db8:1:2::/64" {
		t.Fatalf("ipv6 key = %q", k)
	}
	if ipKey("2001:db8:1:2:ffff::9") != ipKey("2001:db8:1:2::1") {
		t.Fatal("one /64 must share a key")
	}
	if k := ipKey("::ffff:203.0.113.7"); k != "203.0.113.7" {
		t.Fatalf("mapped v4 key = %q", k)
	}
}
