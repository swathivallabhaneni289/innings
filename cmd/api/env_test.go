package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte(`# Innings settings
export INNINGS_TEST_A=plain
INNINGS_TEST_B = "quoted value with spaces # and a hash"
INNINGS_TEST_C='single $quoted & ;special'
INNINGS_TEST_D=trailing comment # ignored
INNINGS_TEST_E=
INNINGS_TEST_KEEP=from-file

INNINGS_TEST_F=a=b=c
`), 0o600)
	t.Setenv("INNINGS_TEST_KEEP", "from-env")
	for _, k := range []string{"INNINGS_TEST_A", "INNINGS_TEST_B", "INNINGS_TEST_C", "INNINGS_TEST_D", "INNINGS_TEST_E", "INNINGS_TEST_F"} {
		t.Setenv(k, "") // registers a cleanup that restores the variable
		os.Unsetenv(k)
	}

	set, err := loadEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"INNINGS_TEST_A":    "plain",
		"INNINGS_TEST_B":    "quoted value with spaces # and a hash",
		"INNINGS_TEST_C":    "single $quoted & ;special",
		"INNINGS_TEST_D":    "trailing comment",
		"INNINGS_TEST_E":    "",
		"INNINGS_TEST_F":    "a=b=c",
		"INNINGS_TEST_KEEP": "from-env", // the environment wins
	}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if slices.Contains(set, "INNINGS_TEST_KEEP") || len(set) != 6 {
		t.Fatalf("keys set: %v", set)
	}

	if set, err := loadEnvFile(filepath.Join(t.TempDir(), "missing")); err != nil || len(set) != 0 {
		t.Fatalf("missing file: %v %v", set, err)
	}
	for _, bad := range []string{"NOT A LINE", "1BAD=x", `INNINGS_TEST_G="unclosed secret`} {
		p := filepath.Join(t.TempDir(), ".env")
		os.WriteFile(p, []byte(bad+"\n"), 0o600)
		_, err := loadEnvFile(p)
		if err == nil || !strings.Contains(err.Error(), "line 1") || strings.Contains(err.Error(), "secret") {
			t.Errorf("%q: %v (errors name the line, never the value)", bad, err)
		}
	}
}
