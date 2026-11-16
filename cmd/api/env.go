package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"
)

var envKeyRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// loadEnvFile reads KEY=VALUE lines from path into the environment, skipping keys that are already set, so a value
// given on the command line wins. It returns the keys it set. A missing file sets nothing. The file is only parsed,
// never run, so a value can hold any character. Lines may start with "export "; blank lines and lines starting with
// # are skipped; a value may be wrapped in single or double quotes; an unquoted value ends at " #". Errors name the
// line, never its value: the file holds passwords.
func loadEnvFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var set []string
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(strings.TrimPrefix(line, "export "), "=")
		key = strings.TrimSpace(key)
		if !ok || !envKeyRE.MatchString(key) {
			return set, fmt.Errorf("%s line %d: expected KEY=VALUE", path, n)
		}
		value = strings.TrimSpace(value)
		switch {
		case len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0]:
			value = value[1 : len(value)-1]
		case strings.HasPrefix(value, `"`) || strings.HasPrefix(value, "'"):
			return set, fmt.Errorf("%s line %d: the quote around %s's value isn't closed", path, n, key)
		default:
			if i := strings.Index(value, " #"); i >= 0 {
				value = strings.TrimSpace(value[:i])
			}
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return set, err
		}
		set = append(set, key)
	}
	return set, sc.Err()
}
