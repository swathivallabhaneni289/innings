package domain

import "testing"

func TestHandleRules(t *testing.T) {
	for h, want := range map[string]bool{
		"abc":                   true,
		"a_b":                   true,
		"rohan.d":               true,
		"fan_2026":              true,
		"abcdefghijklmnopqrst":  true, // 20
		"ab":                    false,
		"abcdefghijklmnopqrstu": false, // 21
		"1abc":                  false,
		"_abc":                  false,
		".abc":                  false,
		"abc.":                  false,
		"ab-c":                  false,
		"Abc":                   false, // must be normalised first
		"ab c":                  false,
		"":                      false,
		"émile":                 false,
	} {
		if got := ValidHandle(h); got != want {
			t.Errorf("ValidHandle(%q) = %v, want %v", h, got, want)
		}
	}
	if NormHandle("  Rohan.D ") != "rohan.d" {
		t.Errorf("NormHandle: %q", NormHandle("  Rohan.D "))
	}
	for _, h := range []string{"innings", "admin", "support", "help", "me", "settings", "official", "team"} {
		if !ReservedHandle(h) {
			t.Errorf("%q should be reserved", h)
		}
	}
	if ReservedHandle("aditi") {
		t.Error("aditi should not be reserved")
	}
}
