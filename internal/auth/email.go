package auth

import (
	"net/mail"
	"strings"
)

// emailMax is the longest address SMTP can carry (RFC 5321 path limit, less the angle brackets).
const emailMax = 254

// NormEmail returns raw trimmed and lower-cased, or ok=false when it doesn't look like an address mail can reach:
// exactly one @, something before it, a domain with a dot and no empty labels, no spaces and no display name.
// Lower-casing the whole address is what every big mail provider does in practice, so "Swathi@Gmail.com" and
// "swathi@gmail.com" are one account. Dots and +tags are left alone: they can be different inboxes elsewhere.
func NormEmail(raw string) (string, bool) {
	e := strings.ToLower(strings.TrimSpace(raw))
	if e == "" || len(e) > emailMax || strings.ContainsAny(e, " \t\r\n<>\"(),;:\\[]") || strings.Count(e, "@") != 1 {
		return "", false
	}
	local, domain, _ := strings.Cut(e, "@")
	if local == "" || len(local) > 64 || !strings.Contains(domain, ".") {
		return "", false
	}
	for label := range strings.SplitSeq(domain, ".") {
		if label == "" || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return "", false
		}
	}
	if a, err := mail.ParseAddress(e); err != nil || a.Address != e || a.Name != "" {
		return "", false
	}
	return e, true
}
