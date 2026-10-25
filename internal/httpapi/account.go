package httpapi

import (
	"cmp"
	"errors"
	"math"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/swathivallabhaneni289/innings/internal/auth"
	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// ---- sign-in (docs/API.md → Auth) ----

// authFail writes err if there is one: an *auth.Error as-is (with Retry-After when it has a wait), anything else
// as a 500. It reports whether it wrote.
func (s *Server) authFail(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	var ae *auth.Error
	if !errors.As(err, &ae) {
		s.Log.Error("auth", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal", "something went wrong")
		return true
	}
	if ae.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(ae.RetryAfter.Seconds()))))
	}
	writeErr(w, ae.Status, ae.Code, ae.Message)
	return true
}

func (s *Server) handleAuthStart(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Phone string `json:"phone"`
	}
	if !decode(w, r, &in) {
		return
	}
	resp, err := s.Auth.Start(r.Context(), in.Phone, auth.ClientIP(r))
	if s.authFail(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if !decode(w, r, &in) {
		return
	}
	si, err := s.Auth.Verify(r.Context(), in.Phone, in.Code)
	if s.authFail(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, domain.AuthVerifyResponse{Token: si.Token, Me: si.User, Account: si.Account, IsNew: si.IsNew})
}

// handleLogout ends only the session this request came in on; the fan's other devices stay signed in.
func (s *Server) handleLogout(v *view, w http.ResponseWriter, r *http.Request) {
	check(s.Auth.Logout(v.ctx, auth.SessionFrom(v.ctx)))
	w.WriteHeader(http.StatusNoContent)
}

// handleBootstrap needs a session now (the middleware answers 401 without one) and echoes its token.
func (s *Server) handleBootstrap(v *view, w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, domain.BootstrapResponse{
		Token:      auth.BearerToken(r),
		Me:         v.me,
		Account:    must(s.Store.Account(v.ctx, v.me.ID)),
		Teams:      v.teams(),
		Players:    nonNilMap(s.World.Players()),
		ServerTime: v.stamp(),
	})
}

// ---- profile set-up ----

// Reasons a handle can't be had (HandleCheckResponse.reason).
const (
	handleInvalid  = "invalid"
	handleReserved = "reserved"
	handleTaken    = "taken"
)

// handleProblem says why h (normalised) can't become my handle, with a human message, or "" when it can. My own
// current handle counts as free. Simulated fans' handles are taken just like real users', ignoring case.
func (s *Server) handleProblem(v *view, h string) (reason, msg string) {
	switch {
	case !domain.ValidHandle(h):
		return handleInvalid, domain.HandleRulesText
	case domain.ReservedHandle(h):
		return handleReserved, "@" + h + " is reserved."
	case strings.EqualFold(v.me.Handle, h):
		return "", ""
	}
	for _, u := range append(slices.Clone(s.World.Friends()), s.World.PublicFans()...) {
		if strings.EqualFold(u.Handle, h) {
			return handleTaken, "@" + h + " is taken."
		}
	}
	if must(s.Store.HandleTaken(v.ctx, h, v.me.ID)) {
		return handleTaken, "@" + h + " is taken."
	}
	return "", ""
}

func (s *Server) handleHandleCheck(v *view, w http.ResponseWriter, r *http.Request) {
	h := domain.NormHandle(r.PathValue("handle"))
	resp := domain.HandleCheckResponse{Handle: h, Available: true}
	if reason, msg := s.handleProblem(v, h); reason != "" {
		resp.Available, resp.Reason, resp.Message = false, &reason, &msg
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleOnboarded finishes sign-up once a name and @username are saved; from then on the fan can be found.
func (s *Server) handleOnboarded(v *view, w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(v.me.Name) == "" || v.me.Handle == "" {
		writeErr(w, http.StatusBadRequest, "profile-incomplete", "Choose a name and a @username first.")
		return
	}
	check(s.Store.SetOnboarded(v.ctx, v.me.ID, s.Auth.Now()))
	writeJSON(w, http.StatusOK, map[string]any{"account": must(s.Store.Account(v.ctx, v.me.ID))})
}

// ---- contacts ----

const maxContactHashes = 5000

var sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// handleContactsMatch finds onboarded fans among the caller's contacts, given SHA-256 hex hashes of the contacts'
// E.164 numbers. The hashes are only compared, never stored or logged.
func (s *Server) handleContactsMatch(v *view, w http.ResponseWriter, r *http.Request) {
	var in struct {
		Hashes []string `json:"hashes"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.Hashes) > maxContactHashes {
		writeErr(w, http.StatusBadRequest, "bad-hashes", "send at most 5000 hashes")
		return
	}
	seen := map[string]bool{}
	uniq := []string{}
	for _, h := range in.Hashes {
		if !sha256Hex.MatchString(h) {
			writeErr(w, http.StatusBadRequest, "bad-hashes", "each hash must be 64 lower-case hex characters (SHA-256 of an E.164 number)")
			return
		}
		if !seen[h] {
			seen[h] = true
			uniq = append(uniq, h)
		}
	}
	rows := []domain.FriendRow{}
	for _, u := range must(s.Store.ListedUsersByPhoneHash(v.ctx, uniq)) {
		if u.ID != v.me.ID {
			rows = append(rows, domain.FriendRow{User: u, Following: v.iFollow(u.ID)})
		}
	}
	slices.SortStableFunc(rows, func(a, b domain.FriendRow) int {
		return cmp.Or(cmp.Compare(strings.ToLower(a.User.Name), strings.ToLower(b.User.Name)), cmp.Compare(a.User.ID, b.User.ID))
	})
	writeJSON(w, http.StatusOK, domain.ContactsMatchResponse{Matches: rows})
}

// ---- deleting an account ----

// handleDeleteMe deletes the account and everything it made (see Store.DeleteUser), signing out every device.
func (s *Server) handleDeleteMe(v *view, w http.ResponseWriter, r *http.Request) {
	check(s.Store.DeleteUser(v.ctx, v.me.ID))
	w.WriteHeader(http.StatusNoContent)
}
