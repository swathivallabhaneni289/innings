package auth

import (
	"net/http/httptest"
	"testing"
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
