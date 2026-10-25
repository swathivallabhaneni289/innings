package domain

import (
	"regexp"
	"strings"
)

// @username rules (docs/API.md → Auth, "Usernames"): 3 to 20 characters of lower-case letters, digits, "_" and ".",
// starting with a letter and not ending with ".". Uniqueness against other fans is checked by the API.
var handleRE = regexp.MustCompile(`^[a-z][a-z0-9_.]{1,18}[a-z0-9_]$`)

// reservedHandles can't be chosen: they would read as the app, its staff or its own screens.
var reservedHandles = map[string]bool{
	"innings": true, "inningsapp": true, "admin": true, "administrator": true, "support": true, "help": true,
	"me": true, "settings": true, "official": true, "team": true, "staff": true, "moderator": true, "mod": true,
	"root": true, "system": true, "api": true, "null": true, "undefined": true, "account": true, "login": true,
	"signin": true, "signup": true, "logout": true, "privacy": true, "terms": true, "security": true, "verify": true,
	"everyone": true,
}

// NormHandle is a handle as it would be saved: trimmed and lower-cased.
func NormHandle(h string) string { return strings.ToLower(strings.TrimSpace(h)) }

// ValidHandle reports whether h (already normalised) follows the format rules.
func ValidHandle(h string) bool { return handleRE.MatchString(h) }

// ReservedHandle reports whether h (already normalised) is kept back for the app.
func ReservedHandle(h string) bool { return reservedHandles[h] }

// HandleRulesText is the human explanation shown when a handle breaks the format rules.
const HandleRulesText = "Usernames are 3–20 characters: lower-case letters, numbers, _ and ., starting with a letter and not ending with ."
