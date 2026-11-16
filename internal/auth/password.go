package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Passwords are stored as PBKDF2-HMAC-SHA256 from the standard library, at OWASP's recommended work factor for it
// (600,000 iterations), with a 16-byte random salt and a 32-byte key:
//
//	pbkdf2-sha256$600000$<salt>$<key>      (salt and key in unpadded standard base64)
//
// The iteration count travels with each hash, so it can be raised later: a sign-in that finds an older count
// stores a fresh hash.
const (
	DefaultIterations = 600_000
	passwordScheme    = "pbkdf2-sha256"
	saltBytes         = 16
	keyBytes          = 32
)

var b64 = base64.RawStdEncoding

// HashPassword returns the stored form of password at the given iteration count.
func HashPassword(password string, iterations int) (string, error) {
	salt := make([]byte, saltBytes)
	rand.Read(salt) // never fails since Go 1.24
	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, keyBytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s$%d$%s$%s", passwordScheme, iterations, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

var errBadHash = errors.New("unreadable password hash")

func parseHash(stored string) (iterations int, salt, key []byte, err error) {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != passwordScheme {
		return 0, nil, nil, errBadHash
	}
	if iterations, err = strconv.Atoi(parts[1]); err != nil || iterations < 1 {
		return 0, nil, nil, errBadHash
	}
	if salt, err = b64.DecodeString(parts[2]); err != nil || len(salt) == 0 {
		return 0, nil, nil, errBadHash
	}
	if key, err = b64.DecodeString(parts[3]); err != nil || len(key) == 0 {
		return 0, nil, nil, errBadHash
	}
	return iterations, salt, key, nil
}

// CheckPassword reports whether password matches the stored hash, comparing in constant time, and whether the hash
// should be replaced because it was made with fewer than want iterations. An empty or unreadable hash (an account
// from the phone era has none) matches nothing.
func CheckPassword(stored, password string, want int) (ok, rehash bool) {
	iterations, salt, key, err := parseHash(stored)
	if err != nil {
		return false, false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(key))
	if err != nil || subtle.ConstantTimeCompare(got, key) != 1 {
		return false, false
	}
	return true, iterations < want
}

// Password rules, after NIST SP 800-63B: a length range and a check against the passwords everyone tries first, and
// no rules about mixing letters, digits and symbols. Spaces count like any other character, so a passphrase works.
const (
	PasswordMin = 8
	PasswordMax = 128
)

// PasswordProblem says what is wrong with a new password for the account at email, in words for the fan, or ""
// when it will do.
func PasswordProblem(password, email string) string {
	n := utf8.RuneCountInString(password)
	switch {
	case n < PasswordMin:
		return fmt.Sprintf("Use at least %d characters.", PasswordMin)
	case n > PasswordMax:
		return fmt.Sprintf("Use at most %d characters.", PasswordMax)
	case !utf8.ValidString(password):
		return "That password has characters we can't read. Try another."
	}
	lower := strings.ToLower(password)
	local, _, _ := strings.Cut(email, "@")
	if lower == email || (len(local) >= 4 && lower == local) {
		return "Your password can't be your email address."
	}
	if commonPasswords[lower] || repeated(lower) {
		return "That's one of the first passwords people guess. Pick something less common."
	}
	return ""
}

// repeated is true for one character over and over ("aaaaaaaa"), or a short run of the keyboard or the digits.
func repeated(p string) bool {
	if first, _ := utf8.DecodeRuneInString(p); strings.Count(p, string(first)) == utf8.RuneCountInString(p) {
		return true
	}
	for _, run := range []string{"0123456789", "9876543210", "abcdefghijklmnopqrstuvwxyz", "qwertyuiop", "asdfghjkl", "zxcvbnm"} {
		if strings.Contains(run, p) {
			return true
		}
	}
	return false
}

// commonPasswords are the most-used passwords of 8 characters or more from public breach lists, plus the obvious
// ones for a cricket app. Shorter ones already fail the length rule.
var commonPasswords = func() map[string]bool {
	m := map[string]bool{}
	for _, p := range strings.Fields(`
		password password1 password12 password123 password1234 passw0rd p@ssw0rd p@ssword pass1234 passpass
		12345678 123456789 1234567890 12341234 123123123 11111111 00000000 12121212 87654321 987654321 11223344
		qwerty12 qwerty123 qwerty1234 qwertyui qwertyuiop 1qaz2wsx qazwsxedc 1q2w3e4r 1q2w3e4r5t zaq12wsx
		abc12345 abcd1234 abcdefgh a1b2c3d4 aa123456 asdf1234 asdfghjk 123qweasd q1w2e3r4 iloveyou iloveyou1
		sunshine princess football baseball superman batman123 trustno1 welcome1 welcome123 letmein1 letmein123
		admin123 administrator monkey123 dragon123 shadow123 master123 michael1 charlie1 jennifer whatever
		starwars computer internet mercedes samsung1 babygirl lovely123 jordan23 liverpool chelsea1 arsenal1
		manchester football1 changeme changeme1 secret123 freedom1 hello123 hello1234 test1234 testing123
		cricket1 cricket123 cricket12 sachin10 sachin123 dhoni007 virat18 virat123 kohli18 india123 india1234
		mumbai123 bangalore chennai123 innings1 innings123
	`) {
		m[p] = true
	}
	return m
}()
