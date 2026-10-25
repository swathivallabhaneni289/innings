package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// mockTwilio is an httptest stand-in for the two Twilio Verify endpoints, built from Twilio's documented request
// and response shapes. These tests exercise our client against it; nothing here talks to real Twilio.
type mockTwilio struct {
	mu       sync.Mutex
	code     map[string]string // phone → pending code
	sends    int
	failSend bool
	calls    []string
}

const (
	mockSID     = "AC00000000000000000000000000000000"
	mockToken   = "secret-token"
	mockService = "VA00000000000000000000000000000000"
)

func newMockTwilio(t *testing.T) (*mockTwilio, *Twilio) {
	m := &mockTwilio{code: map[string]string{}}
	srv := httptest.NewServer(m)
	t.Cleanup(srv.Close)
	return m, &Twilio{AccountSID: mockSID, AuthToken: mockToken, ServiceSID: mockService, BaseURL: srv.URL}
}

func (m *mockTwilio) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, r.Method+" "+r.URL.Path)
	user, pass, ok := r.BasicAuth()
	if !ok || user != mockSID || pass != mockToken {
		http.Error(w, `{"code":20003,"message":"Authenticate","status":401}`, http.StatusUnauthorized)
		return
	}
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" || r.ParseForm() != nil {
		http.Error(w, `{"code":20001,"message":"bad form","status":400}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	to := r.PostForm.Get("To")
	switch r.URL.Path {
	case "/v2/Services/" + mockService + "/Verifications":
		if m.failSend || r.PostForm.Get("Channel") != "sms" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"code":60200,"message":"Invalid parameter","status":400}`))
			return
		}
		m.sends++
		if m.code[to] == "" {
			m.code[to] = "424242"
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"sid":"VE1","to":"` + to + `","channel":"sms","status":"pending"}`))
	case "/v2/Services/" + mockService + "/VerificationCheck":
		code, live := m.code[to]
		if !live {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code":20404,"message":"The requested resource was not found","status":404}`))
			return
		}
		status := "pending"
		if r.PostForm.Get("Code") == code {
			status = "approved"
			delete(m.code, to) // an approved verification is gone, so checking again gets 404
		}
		w.Write([]byte(`{"sid":"VE1","to":"` + to + `","status":"` + status + `","valid":` + map[bool]string{true: "true", false: "false"}[status == "approved"] + `}`))
	default:
		http.NotFound(w, r)
	}
}

func TestTwilioAgainstMock(t *testing.T) {
	ctx := context.Background()
	m, tw := newMockTwilio(t)
	const phone = "+919876543210"
	if err := tw.Send(ctx, phone, ""); err != nil {
		t.Fatalf("send: %v", err)
	}
	if ok, err := tw.Check(ctx, phone, "000000"); err != nil || ok {
		t.Fatalf("wrong code: %v %v", ok, err)
	}
	if ok, err := tw.Check(ctx, phone, "424242"); err != nil || !ok {
		t.Fatalf("right code: %v %v", ok, err)
	}
	if _, err := tw.Check(ctx, phone, "424242"); !errors.Is(err, ErrCodeGone) {
		t.Fatalf("approved verification checked again: want ErrCodeGone, got %v", err)
	}
	if _, err := tw.Check(ctx, "+14155550100", "123456"); !errors.Is(err, ErrCodeGone) {
		t.Fatalf("never started: want ErrCodeGone, got %v", err)
	}
	m.failSend = true
	if err := tw.Send(ctx, phone, ""); err == nil {
		t.Fatal("a 400 from Twilio must be an error")
	}
	bad := *tw
	bad.AuthToken = "wrong"
	if err := bad.Send(ctx, phone, ""); err == nil {
		t.Fatal("a 401 from Twilio must be an error")
	}
	if _, err := bad.Check(ctx, phone, "123456"); err == nil || errors.Is(err, ErrCodeGone) {
		t.Fatalf("a 401 on check is a failure, not an expired code: %v", err)
	}
	want := "POST /v2/Services/" + mockService + "/Verifications"
	if len(m.calls) == 0 || m.calls[0] != want {
		t.Fatalf("calls: %v", m.calls)
	}
}

func TestTwilioUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	tw := &Twilio{AccountSID: "a", AuthToken: "b", ServiceSID: "c", BaseURL: url}
	if err := tw.Send(context.Background(), "+919876543210", ""); err == nil {
		t.Fatal("unreachable Twilio must be an error")
	}
	if _, err := tw.Check(context.Background(), "+919876543210", "123456"); err == nil || errors.Is(err, ErrCodeGone) {
		t.Fatalf("unreachable on check: %v", err)
	}
}
