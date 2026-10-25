package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/auth"
)

// errCode decodes {"error": ...} from a body.
func errCode(b []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	json.Unmarshal(b, &e)
	return e.Error
}

func (c *client) verify(phone, code string) (int, []byte) {
	c.t.Helper()
	return c.do("POST", "/v1/auth/verify", map[string]string{"phone": phone, "code": code})
}

func wrong(code string) string {
	if code == "000000" {
		return "111111"
	}
	return "000000"
}

func TestSignInAndSecondDevice(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	phone := nextPhone()

	// The app may send the number as typed; it comes back as E.164.
	status, b, _ := c.req("POST", "/v1/auth/start", map[string]string{"phone": phone[:3] + " " + phone[3:8] + "-" + phone[8:]})
	var start struct {
		Phone            string `json:"phone"`
		ResendInSeconds  int    `json:"resendInSeconds"`
		ExpiresInSeconds int    `json:"expiresInSeconds"`
		DevCode          string `json:"devCode"`
	}
	json.Unmarshal(b, &start)
	if status != 200 || start.Phone != phone || start.ResendInSeconds != 30 || start.ExpiresInSeconds != 600 || len(start.DevCode) != 6 {
		t.Fatalf("start: %d %s", status, b)
	}
	status, b = c.verify(phone, start.DevCode)
	if status != 200 {
		t.Fatalf("verify: %d %s", status, b)
	}
	var first struct {
		Token   string         `json:"token"`
		Me      map[string]any `json:"me"`
		Account map[string]any `json:"account"`
		IsNew   bool           `json:"isNew"`
	}
	json.Unmarshal(b, &first)
	if !first.IsNew || len(first.Token) != 64 || first.Me["name"] != "" || first.Me["handle"] != "" ||
		first.Me["avatarColour"] != "#3B3F6B" || first.Me["favouriteTeamId"] != "ind" {
		t.Fatalf("new account: %s", b)
	}
	if first.Account["phone"] != phone || first.Account["onboarded"] != false || first.Account["createdAt"] == "" {
		t.Fatalf("account: %v", first.Account)
	}
	if _, ok := first.Me["phone"]; ok {
		t.Fatal("phone must never be part of User")
	}
	c.token = first.Token

	// A new account starts empty: no logs, predictions or follows.
	me := c.getJSON("/v1/users/me")
	if len(list(me["logs"])) != 0 || me["stats"].(map[string]any)["logged"] != 0.0 {
		t.Fatalf("new account has history: %v", me["logs"])
	}
	for _, f := range list(c.getJSON("/v1/friends")["friends"]) {
		if f.(map[string]any)["following"] == true {
			t.Fatalf("new account follows %v", str(f, "user", "id"))
		}
	}
	// Someone else looking me up never sees my number.
	other := c.another()
	if _, b := other.do("GET", "/v1/users/"+str(first.Me, "id"), nil); strings.Contains(string(b), phone[1:]) {
		t.Fatalf("phone leaked in a profile: %s", b)
	}

	// Same number on a second device: same fan, new session, both work.
	c.clock.advance(auth.ResendGap)
	second := &client{t: t, base: c.base, api: c.api, clock: c.clock}
	got := second.signIn(phone)
	if got["isNew"] != false || second.id != str(first.Me, "id") || second.token == c.token {
		t.Fatalf("second device: %v", got)
	}
	if str(c.getJSON("/v1/users/me")["user"], "id") != second.id {
		t.Fatal("first device lost its session")
	}

	if status, b := c.do("POST", "/v1/auth/start", map[string]string{"phone": "9876543210"}); status != 400 || errCode(b) != "bad-phone" {
		t.Fatalf("bad phone: %d %s", status, b)
	}
	if status, b := c.verify("+12", "123456"); status != 400 || errCode(b) != "bad-phone" {
		t.Fatalf("verify bad phone: %d %s", status, b)
	}
}

func TestWrongCodeCountsDown(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	phone := nextPhone()
	code := c.startCode(phone)
	for _, want := range []string{"4 tries left", "3 tries left", "2 tries left", "1 try left", "last try"} {
		status, b := c.verify(phone, wrong(code))
		if status != 400 || errCode(b) != "bad-code" || !strings.Contains(string(b), want) {
			t.Fatalf("want 400 bad-code %q, got %d %s", want, status, b)
		}
	}
	// Out of tries: even the right code is refused now.
	if status, b := c.verify(phone, code); status != http.StatusGone || errCode(b) != "code-expired" {
		t.Fatalf("after 5 wrong tries: %d %s", status, b)
	}
	// A malformed guess also spends a try.
	c.clock.advance(auth.ResendGap)
	code = c.startCode(phone)
	if status, b := c.verify(phone, "12ab"); status != 400 || !strings.Contains(string(b), "4 tries left") {
		t.Fatalf("malformed code: %d %s", status, b)
	}
	if status, b := c.verify(phone, code); status != 200 {
		t.Fatalf("right code after one wrong: %d %s", status, b)
	}
}

func TestCodeExpires(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	phone := nextPhone()
	code := c.startCode(phone)
	c.clock.advance(auth.CodeTTL - time.Second)
	if status, b := c.verify(phone, wrong(code)); status != 400 {
		t.Fatalf("just before expiry: %d %s", status, b)
	}
	c.clock.advance(time.Second)
	if status, b := c.verify(phone, code); status != http.StatusGone || errCode(b) != "code-expired" {
		t.Fatalf("expired code: %d %s", status, b)
	}
	if status, b := c.verify(nextPhone(), "123456"); status != http.StatusGone {
		t.Fatalf("never sent: %d %s", status, b)
	}
}

func TestCodeCantBeReused(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	phone := nextPhone()
	code := c.startCode(phone)
	if status, b := c.verify(phone, code); status != 200 {
		t.Fatalf("verify: %d %s", status, b)
	}
	if status, b := c.verify(phone, code); status != http.StatusGone || errCode(b) != "code-expired" {
		t.Fatalf("reused code: %d %s", status, b)
	}
}

func TestResendTooSoon(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	phone := nextPhone()
	c.startCode(phone)
	status, b, h := c.req("POST", "/v1/auth/start", map[string]string{"phone": phone})
	if status != http.StatusTooManyRequests || errCode(b) != "too-soon" || h.Get("Retry-After") != "30" {
		t.Fatalf("immediate resend: %d %s Retry-After=%q", status, b, h.Get("Retry-After"))
	}
	c.clock.advance(10 * time.Second)
	if _, _, h := c.req("POST", "/v1/auth/start", map[string]string{"phone": phone}); h.Get("Retry-After") != "20" {
		t.Fatalf("Retry-After after 10 s = %q", h.Get("Retry-After"))
	}
	c.clock.advance(20 * time.Second)
	code := c.startCode(phone) // a new code replaces the old one
	if status, b := c.verify(phone, code); status != 200 {
		t.Fatalf("new code: %d %s", status, b)
	}
}

func TestHourlyCapPerNumber(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	phone := nextPhone()
	for i := range auth.SendsPerHour {
		if i > 0 {
			c.clock.advance(auth.ResendGap)
		}
		c.startCode(phone)
	}
	c.clock.advance(auth.ResendGap)
	status, b, h := c.req("POST", "/v1/auth/start", map[string]string{"phone": phone})
	// The first of the five went out 150 s ago, so it ages out in 3450 s.
	if status != http.StatusTooManyRequests || errCode(b) != "too-many" || h.Get("Retry-After") != "3450" {
		t.Fatalf("sixth code this hour: %d %s Retry-After=%q", status, b, h.Get("Retry-After"))
	}
	c.startCode(nextPhone()) // another number is unaffected
	c.clock.advance(58 * time.Minute)
	c.startCode(phone)
}

func TestIPCap(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	for range auth.StartsPerIPPerHour {
		c.startCode(nextPhone())
	}
	for _, xff := range []string{"", "198.51.100.7"} {
		req, _ := http.NewRequest("POST", c.base+"/v1/auth/start", strings.NewReader(`{"phone":"`+nextPhone()+`"}`))
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff) // must not buy a fresh allowance
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var b [256]byte
		n, _ := res.Body.Read(b[:])
		res.Body.Close()
		if res.StatusCode != http.StatusTooManyRequests || errCode(b[:n]) != "too-many" || res.Header.Get("Retry-After") == "" {
			t.Fatalf("21st start from one IP (XFF %q): %d %s", xff, res.StatusCode, b[:n])
		}
	}
	c.clock.advance(time.Hour)
	c.startCode(nextPhone())
}

func TestNoSMSProvider(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	c.api.Auth.Verifier = nil
	status, b := c.do("POST", "/v1/auth/start", map[string]string{"phone": nextPhone()})
	if status != http.StatusServiceUnavailable || errCode(b) != "sms-unavailable" {
		t.Fatalf("no provider: %d %s", status, b)
	}
}

// fakeTwilio mocks the two Twilio Verify endpoints (not real Twilio). Its code is always 424242, a number ending
// in 99 fails to send, and expire() forgets a verification the way Twilio does after 10 minutes.
type fakeTwilio struct {
	mu      sync.Mutex
	pending map[string]bool
}

func (f *fakeTwilio) expire(phone string) { f.mu.Lock(); defer f.mu.Unlock(); delete(f.pending, phone) }

func (f *fakeTwilio) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r.ParseForm()
	to := r.PostForm.Get("To")
	w.Header().Set("Content-Type", "application/json")
	switch {
	case strings.HasSuffix(r.URL.Path, "/Verifications"):
		if strings.HasSuffix(to, "99") {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"code":60200,"message":"Invalid parameter: To","status":400}`)
			return
		}
		f.pending[to] = true
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"status":"pending"}`)
	case strings.HasSuffix(r.URL.Path, "/VerificationCheck"):
		if !f.pending[to] {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"code":20404,"message":"not found","status":404}`)
			return
		}
		if r.PostForm.Get("Code") != "424242" {
			fmt.Fprint(w, `{"status":"pending"}`)
			return
		}
		delete(f.pending, to)
		fmt.Fprint(w, `{"status":"approved"}`)
	default:
		http.NotFound(w, r)
	}
}

// TestTwilioMode runs sign-in end to end with the Twilio verifier pointed at a local mock. It shows no devCode
// leaves the server and that Twilio's answers map to our errors. It does not exercise real Twilio.
func TestTwilioMode(t *testing.T) {
	c := newTestServer(t, newFakeWorld(testNow))
	mock := &fakeTwilio{pending: map[string]bool{}}
	tw := httptest.NewServer(mock)
	t.Cleanup(tw.Close)
	c.api.Auth.Verifier = &auth.Twilio{AccountSID: "AC1", AuthToken: "tok", ServiceSID: "VA1", BaseURL: tw.URL}

	phone := nextPhone()
	status, b := c.do("POST", "/v1/auth/start", map[string]string{"phone": phone})
	var raw map[string]any
	json.Unmarshal(b, &raw)
	if status != 200 {
		t.Fatalf("start: %d %s", status, b)
	}
	if _, ok := raw["devCode"]; ok {
		t.Fatalf("devCode returned with Twilio configured: %s", b)
	}
	if status, b := c.verify(phone, "000000"); status != 400 || !strings.Contains(string(b), "4 tries left") {
		t.Fatalf("wrong code: %d %s", status, b)
	}
	status, b = c.verify(phone, "424242")
	if status != 200 || !strings.Contains(string(b), `"isNew":true`) {
		t.Fatalf("right code: %d %s", status, b)
	}
	if status, b := c.verify(phone, "424242"); status != http.StatusGone {
		t.Fatalf("reuse: %d %s", status, b)
	}

	// Twilio forgets the verification (404 on check): code-expired.
	phone2 := nextPhone()
	c.do("POST", "/v1/auth/start", map[string]string{"phone": phone2})
	mock.expire(phone2)
	if status, b := c.verify(phone2, "424242"); status != http.StatusGone || errCode(b) != "code-expired" {
		t.Fatalf("Twilio 404: %d %s", status, b)
	}
	// Twilio refuses to send: sms-failed, and no live code is left behind.
	failing := nextPhone()
	failing = failing[:len(failing)-2] + "99"
	if status, b := c.do("POST", "/v1/auth/start", map[string]string{"phone": failing}); status != http.StatusBadGateway || errCode(b) != "sms-failed" {
		t.Fatalf("send refused: %d %s", status, b)
	}
	if status, _ := c.verify(failing, "424242"); status != http.StatusGone {
		t.Fatalf("after a failed send: %d", status)
	}
}

func TestLogoutEndsOnlyThisSession(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	other := c.device()
	if status, _ := c.do("POST", "/v1/auth/logout", nil); status != http.StatusNoContent {
		t.Fatalf("logout: %d", status)
	}
	if status, _ := c.do("GET", "/v1/home", nil); status != http.StatusUnauthorized {
		t.Fatalf("signed-out token still works: %d", status)
	}
	if status, _ := c.do("POST", "/v1/auth/logout", nil); status != http.StatusUnauthorized {
		t.Fatalf("logout twice: %d", status)
	}
	other.getJSON("/v1/home") // the same fan's other device is still signed in
}

func TestHandles(t *testing.T) { eachWorld(t, testHandles) }

func testHandles(t *testing.T, c *client) {
	check := func(h, reason string) {
		t.Helper()
		r := c.getJSON("/v1/handles/"+h, "handle", "available", "reason", "message")
		if reason == "" && (r["available"] != true || r["reason"] != nil || r["message"] != nil) {
			t.Errorf("%s: want available, got %v", h, r)
		}
		if reason != "" && (r["available"] != false || r["reason"] != reason || r["message"] == nil) {
			t.Errorf("%s: want %s, got %v", h, reason, r)
		}
	}
	if r := c.getJSON("/v1/handles/Swathi.R"); r["handle"] != "swathi.r" || r["available"] != true {
		t.Fatalf("lower-cased check: %v", r)
	}
	for _, h := range []string{"ab", "1abc", "abc.", "ab-c", "abcdefghijklmnopqrstu"} {
		check(h, "invalid")
	}
	for _, h := range []string{"innings", "Admin", "support", "help", "settings", "official", "team"} {
		check(h, "reserved")
	}
	check("me", "invalid") // too short as well as reserved
	friends := list(c.getJSON("/v1/friends")["friends"])
	if len(friends) == 0 {
		t.Fatal("world has no simulated friends")
	}
	simHandle := str(friends[0], "user", "handle")
	check(strings.ToUpper(simHandle), "taken")

	if status, b := c.do("PUT", "/v1/me", map[string]any{"handle": "Swathi.R"}); status != 200 || !strings.Contains(string(b), `"handle":"swathi.r"`) {
		t.Fatalf("set handle: %d %s", status, b)
	}
	check("SWATHI.R", "") // my own handle counts as available
	if status, b := c.do("PUT", "/v1/me", map[string]any{"handle": "swathi.r"}); status != 200 {
		t.Fatalf("re-save own handle: %d %s", status, b)
	}

	other := c.another()
	if r := other.getJSON("/v1/handles/swathi.r"); r["reason"] != "taken" {
		t.Fatalf("someone else's handle: %v", r)
	}
	for h, want := range map[string]string{"Swathi.R": "handle-taken", simHandle: "handle-taken", "admin": "bad-handle", "x": "bad-handle", "": "bad-handle"} {
		code := map[string]int{"handle-taken": 409, "bad-handle": 400}[want]
		if status, b := other.do("PUT", "/v1/me", map[string]any{"handle": h}); status != code || errCode(b) != want {
			t.Errorf("PUT handle %q: %d %s", h, status, b)
		}
	}
	if u := other.getJSON("/v1/users/me")["user"]; str(u, "handle") != "" {
		t.Fatalf("a refused handle was saved: %v", u)
	}
}

// onboard saves a name and handle and finishes sign-up.
func (c *client) onboard(name, handle string) {
	c.t.Helper()
	if status, b := c.do("PUT", "/v1/me", map[string]any{"name": name, "handle": handle}); status != 200 {
		c.t.Fatalf("profile: %d %s", status, b)
	}
	if status, b := c.do("POST", "/v1/me/onboarded", nil); status != 200 {
		c.t.Fatalf("onboarded: %d %s", status, b)
	}
}

func TestOnboarding(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	if status, b := c.do("POST", "/v1/me/onboarded", nil); status != 400 || errCode(b) != "profile-incomplete" {
		t.Fatalf("blank profile: %d %s", status, b)
	}
	c.do("PUT", "/v1/me", map[string]any{"name": "Swathi"})
	if status, b := c.do("POST", "/v1/me/onboarded", nil); status != 400 || errCode(b) != "profile-incomplete" {
		t.Fatalf("no handle yet: %d %s", status, b)
	}
	c.do("PUT", "/v1/me", map[string]any{"handle": "swathi"})
	status, b := c.do("POST", "/v1/me/onboarded", nil)
	var out struct {
		Account map[string]any `json:"account"`
	}
	json.Unmarshal(b, &out)
	if status != 200 || out.Account["onboarded"] != true || out.Account["phone"] != c.phone {
		t.Fatalf("onboarded: %d %s", status, b)
	}
	if _, boot := c.bootstrapStatus(c.token); boot["account"].(map[string]any)["onboarded"] != true {
		t.Fatalf("bootstrap account: %v", boot["account"])
	}
}

func TestNotOnboardedIsHidden(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	shown, hidden := c.another(), c.another()
	shown.onboard("Zebedee Shown", "zshown")
	hidden.do("PUT", "/v1/me", map[string]any{"name": "Zebedee Hidden", "handle": "zhidden"}) // never finishes sign-up

	users := list(c.getJSON("/v1/search?q=zebedee")["users"])
	if len(users) != 1 || str(users[0], "id") != shown.id {
		t.Fatalf("search: %v", users)
	}
	c.do("POST", "/v1/users/"+shown.id+"/follow", nil)
	c.do("POST", "/v1/users/"+hidden.id+"/follow", nil)
	ids := map[string]bool{}
	for _, f := range list(c.getJSON("/v1/friends")["friends"]) {
		ids[str(f, "user", "id")] = true
	}
	if !ids[shown.id] || ids[hidden.id] {
		t.Fatalf("friends list: shown=%v hidden=%v", ids[shown.id], ids[hidden.id])
	}
}

func phoneHash(phone string) string {
	sum := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(sum[:])
}

func TestContactsMatch(t *testing.T) {
	c := newTestAPI(t, newFakeWorld(testNow))
	c.onboard("Me", "me_myself")
	friend, halfway := c.another(), c.another()
	friend.onboard("Aarav", "aarav")
	match := func(hashes []string) (int, []byte) {
		return c.do("POST", "/v1/contacts/match", map[string]any{"hashes": hashes})
	}
	hashes := []string{phoneHash(friend.phone), phoneHash(halfway.phone), phoneHash(c.phone), phoneHash("+447700900123"), phoneHash(friend.phone)}
	status, b := match(hashes)
	var out struct {
		Matches []struct {
			User      map[string]any `json:"user"`
			Following bool           `json:"following"`
		} `json:"matches"`
	}
	json.Unmarshal(b, &out)
	if status != 200 || len(out.Matches) != 1 || out.Matches[0].User["id"] != friend.id || out.Matches[0].Following {
		t.Fatalf("match: %d %s", status, b)
	}
	if strings.Contains(string(b), friend.phone[1:]) {
		t.Fatalf("phone in contact match: %s", b)
	}
	c.do("POST", "/v1/users/"+friend.id+"/follow", nil)
	if _, b := match(hashes[:1]); !strings.Contains(string(b), `"following":true`) {
		t.Fatalf("following after follow: %s", b)
	}
	if status, b := match([]string{}); status != 200 || !strings.Contains(string(b), `"matches":[]`) {
		t.Fatalf("no hashes: %d %s", status, b)
	}
	tooMany := make([]string, 5001)
	for i := range tooMany {
		tooMany[i] = phoneHash(fmt.Sprint(i))
	}
	for _, bad := range [][]string{{strings.ToUpper(phoneHash(friend.phone))}, {"abc"}, {friend.phone}, tooMany} {
		if status, b := match(bad); status != 400 || errCode(b) != "bad-hashes" {
			t.Errorf("bad hashes (%d): %d %s", len(bad), status, b)
		}
	}
	if status, b := match(tooMany[:5000]); status != 200 {
		t.Fatalf("5000 hashes: %d %s", status, b)
	}
}

func TestDeleteAccount(t *testing.T) {
	gone := newTestAPI(t, newFakeWorld(testNow))
	gone.onboard("Leaving Soon", "leaving")
	stays := gone.another()
	stays.onboard("Staying Put", "staying")
	goneDevice2 := gone.device()

	newLog := func(c *client, matchID string) string {
		status, b := c.do("POST", "/v1/logs", map[string]any{"matchId": matchID, "rating": 4, "watchedAt": "home", "review": "Good one"})
		if status != 201 {
			t.Fatalf("log: %d %s", status, b)
		}
		var fi struct {
			Log struct {
				ID string `json:"id"`
			} `json:"log"`
		}
		json.Unmarshal(b, &fi)
		return fi.Log.ID
	}
	goneLog, staysLog := newLog(gone, "done"), newLog(stays, "old")
	gone.do("POST", "/v1/users/"+stays.id+"/follow", nil)
	stays.do("POST", "/v1/users/"+gone.id+"/follow", nil)
	for _, pair := range []struct {
		c   *client
		log string
	}{{gone, staysLog}, {stays, goneLog}} {
		pair.c.do("POST", "/v1/logs/"+pair.log+"/like", nil)
		pair.c.do("POST", "/v1/logs/"+pair.log+"/there-too", nil)
		if status, b := pair.c.do("POST", "/v1/logs/"+pair.log+"/replies", map[string]string{"text": "Was there!"}); status != 201 {
			t.Fatalf("reply: %d %s", status, b)
		}
	}
	gone.do("PUT", "/v1/matches/soon/prediction", map[string]string{"winnerId": "mi", "topRunScorerId": "mi-p0", "topWicketTakerId": "csk-p1"})
	gone.do("PUT", "/v1/matches/live/watching", map[string]any{"place": "cafe"})
	gone.do("PUT", "/v1/me/top5", map[string]any{"matchIds": []string{"done"}})
	// Before: stays sees gone everywhere.
	if fi := stays.getJSON("/v1/logs/" + staysLog); fi["likes"] != 1.0 || fi["thereToo"] != 1.0 || len(list(fi["replies"])) != 1 {
		t.Fatalf("before delete: %v", fi)
	}
	if !strings.Contains(fmt.Sprint(stays.getJSON("/v1/notifications")["items"]), "Leaving Soon") {
		t.Fatal("before delete: no notification from the leaving fan")
	}

	if status, b := gone.do("DELETE", "/v1/me", nil); status != http.StatusNoContent {
		t.Fatalf("delete: %d %s", status, b)
	}
	for _, d := range []*client{gone, goneDevice2} {
		if status, _ := d.do("GET", "/v1/home", nil); status != http.StatusUnauthorized {
			t.Fatalf("deleted account's session still works: %d", status)
		}
	}

	// Everything stays's screens load, with no trace of the deleted fan.
	home := stays.getJSON("/v1/home", "feed", "board")
	for _, row := range list(home["board"]) {
		if str(row, "user", "id") == gone.id {
			t.Error("deleted fan still on the board")
		}
	}
	for _, fi := range list(home["feed"]) {
		if str(fi, "user", "id") == gone.id {
			t.Error("deleted fan still in the feed")
		}
	}
	if notes := stays.getJSON("/v1/notifications", "items"); strings.Contains(fmt.Sprint(notes["items"]), "Leaving Soon") {
		t.Errorf("notifications still mention the deleted fan: %v", notes["items"])
	}
	fi := stays.getJSON("/v1/logs/"+staysLog, "log", "replies")
	if fi["likes"] != 0.0 || fi["thereToo"] != 0.0 || len(list(fi["replies"])) != 0 {
		t.Fatalf("stays's stub after delete: %v", fi)
	}
	if status, _ := stays.do("GET", "/v1/logs/"+goneLog, nil); status != http.StatusNotFound {
		t.Fatalf("deleted fan's log: %d", status)
	}
	if status, _ := stays.do("GET", "/v1/users/"+gone.id, nil); status != http.StatusNotFound {
		t.Fatalf("deleted fan's profile: %d", status)
	}
	for _, f := range list(stays.getJSON("/v1/friends")["friends"]) {
		if str(f, "user", "id") == gone.id {
			t.Error("deleted fan still in friends")
		}
	}
	if users := list(stays.getJSON("/v1/search?q=leaving")["users"]); len(users) != 0 {
		t.Errorf("deleted fan still in search: %v", users)
	}
	stays.getJSON("/v1/users/me", "stats", "logs")
	stays.getJSON("/v1/matches/done", "fans")
	if sp := stays.getJSON("/v1/matches/soon")["split"].(map[string]any); sp["total"] != 0.0 {
		t.Errorf("deleted fan's prediction still counted: %v", sp)
	}
	if w := stays.getJSON("/v1/matches/live")["friendsWatching"]; strings.Contains(fmt.Sprint(w), gone.id) {
		t.Errorf("deleted fan still watching: %v", w)
	}

	// The number is free again: signing in makes a brand-new account.
	again := &client{t: t, base: gone.base, api: gone.api, clock: gone.clock}
	if got := again.signIn(gone.phone); got["isNew"] != true || again.id == gone.id {
		t.Fatalf("re-sign-up: %v", got)
	}
}
