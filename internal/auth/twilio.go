package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Twilio is a Verifier backed by Twilio Verify, called over plain net/http (no SDK). Twilio makes, texts and checks
// the code; it also expires verifications after 10 minutes and stops after 5 wrong checks, matching our limits.
//
//	Start: POST {base}/v2/Services/{ServiceSID}/Verifications      To=<E.164>&Channel=sms  → {"status":"pending"}
//	Check: POST {base}/v2/Services/{ServiceSID}/VerificationCheck  To=<E.164>&Code=<code>  → {"status":"approved"|"pending"}
//
// Both use HTTP Basic auth (Account SID : Auth Token). A 404 from Check means the verification is gone (expired,
// already approved, or out of attempts).
//
// This has only been tested against an httptest mock of those two endpoints (twilio_test.go), built from Twilio's
// documented request and response shapes. It has not been run against real Twilio.
type Twilio struct {
	AccountSID, AuthToken, ServiceSID string
	BaseURL                           string       // "" = https://verify.twilio.com; tests point it at a mock
	Client                            *http.Client // nil = a client with a 10 s timeout
}

var defaultTwilioClient = &http.Client{Timeout: 10 * time.Second}

func (*Twilio) ProviderCodes() bool { return true }

// Send starts (or, within Twilio's 10 minutes, re-sends) a verification by SMS. The code argument is ignored:
// Twilio makes its own.
func (t *Twilio) Send(ctx context.Context, phone, _ string) error {
	var out struct {
		Status string `json:"status"`
	}
	_, err := t.post(ctx, "Verifications", url.Values{"To": {phone}, "Channel": {"sms"}}, &out)
	return err
}

// Check reports whether code is right: "approved" is right, "pending" is wrong. 404, or any final status, is
// ErrCodeGone.
func (t *Twilio) Check(ctx context.Context, phone, code string) (bool, error) {
	var out struct {
		Status string `json:"status"`
	}
	status, err := t.post(ctx, "VerificationCheck", url.Values{"To": {phone}, "Code": {code}}, &out)
	if status == http.StatusNotFound {
		return false, ErrCodeGone
	}
	if err != nil {
		return false, err
	}
	switch out.Status {
	case "approved":
		return true, nil
	case "pending":
		return false, nil
	default: // canceled, expired, max_attempts_reached, deleted, failed
		return false, ErrCodeGone
	}
}

// twilioError is Twilio's error body ({"code": 20404, "message": "...", "status": 404}).
type twilioError struct {
	Status  int    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *twilioError) Error() string {
	return fmt.Sprintf("twilio: HTTP %d, code %d: %s", e.Status, e.Code, e.Message)
}

// post sends a form to one Verify endpoint and decodes a 2xx JSON body into out. It returns the HTTP status (0 if
// the request never got an answer) and, for a non-2xx answer, a *twilioError.
func (t *Twilio) post(ctx context.Context, endpoint string, form url.Values, out any) (int, error) {
	base := strings.TrimSuffix(t.BaseURL, "/")
	if base == "" {
		base = "https://verify.twilio.com"
	}
	client := t.Client
	if client == nil {
		client = defaultTwilioClient
	}
	u := base + "/v2/Services/" + url.PathEscape(t.ServiceSID) + "/" + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("twilio %s: %w", endpoint, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if err != nil {
		return res.StatusCode, fmt.Errorf("twilio %s: %w", endpoint, err)
	}
	if res.StatusCode/100 != 2 {
		e := &twilioError{Status: res.StatusCode}
		json.Unmarshal(body, e) // best effort: the status alone is enough to act on
		e.Status = res.StatusCode
		return res.StatusCode, e
	}
	if err := json.Unmarshal(body, out); err != nil {
		return res.StatusCode, fmt.Errorf("twilio %s: bad JSON: %w", endpoint, err)
	}
	return res.StatusCode, nil
}
