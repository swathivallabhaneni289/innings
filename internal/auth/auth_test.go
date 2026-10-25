package auth

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"regexp"
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

func TestNormPhone(t *testing.T) {
	for raw, want := range map[string]string{
		"+919876543210":     "+919876543210",
		"+91 98765 43210":   "+919876543210",
		"+1 (415) 555-0100": "+14155550100",
		"+4412345678":       "+4412345678",
	} {
		if got, ok := NormPhone(raw); !ok || got != want {
			t.Errorf("NormPhone(%q) = %q %v", raw, got, ok)
		}
	}
	for _, raw := range []string{"", "919876543210", "+0919876543210", "+91987", "+1234567890123456", "+91 98765 4321x", "++919876543210"} {
		if _, ok := NormPhone(raw); ok {
			t.Errorf("NormPhone(%q) accepted", raw)
		}
	}
}

func TestVerifierFromEnv(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	env := func(kv map[string]string) func(string) string { return func(k string) string { return kv[k] } }
	twilio := map[string]string{"TWILIO_ACCOUNT_SID": "AC1", "TWILIO_AUTH_TOKEN": "tok", "TWILIO_VERIFY_SERVICE_SID": "VA1"}

	v, mode, err := VerifierFromEnv(env(twilio), logger)
	if _, ok := v.(*Twilio); err != nil || mode != "twilio" || !ok {
		t.Fatalf("twilio: %T %q %v", v, mode, err)
	}
	withDev := map[string]string{"INNINGS_DEV_OTP": "1"}
	for k, val := range twilio {
		withDev[k] = val
	}
	if v, mode, _ := VerifierFromEnv(env(withDev), logger); mode != "twilio" {
		t.Fatalf("Twilio must win over dev mode: %T %q", v, mode)
	}
	if v, mode, err := VerifierFromEnv(env(map[string]string{"INNINGS_DEV_OTP": "1"}), logger); err != nil || mode != "dev" {
		t.Fatalf("dev: %T %q %v", v, mode, err)
	} else if _, ok := v.(devCodes); !ok {
		t.Fatal("dev verifier must reveal codes")
	}
	if v, mode, err := VerifierFromEnv(env(nil), logger); err != nil || mode != "off" || v != nil {
		t.Fatalf("off: %T %q %v", v, mode, err)
	}
	partial := map[string]string{"TWILIO_ACCOUNT_SID": "AC1", "INNINGS_DEV_OTP": "1"}
	if _, _, err := VerifierFromEnv(env(partial), logger); err == nil {
		t.Fatal("partial Twilio settings must be an error, not dev mode")
	}
	var tw Verifier = &Twilio{}
	if _, ok := tw.(devCodes); ok {
		t.Fatal("Twilio must never reveal codes")
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
