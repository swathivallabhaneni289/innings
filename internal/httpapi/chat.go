package httpapi

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

// The match chat: one room per fixture, open from a day before the first ball until a day after the match's
// usual end, so fans can talk through the build-up, the game and the post-match. Vault classics have none.
const (
	chatOpensBefore = 24 * time.Hour
	chatStaysAfter  = 24 * time.Hour
	chatPage        = 200         // messages per read
	chatGap         = time.Second // one message a second per fan
	chatStamp       = "2006-01-02T15:04:05.000Z07:00"
)

// typicalEnd is when a match usually finishes: its start plus the format's usual length.
func typicalEnd(m domain.Match) time.Time {
	d := 3*time.Hour + 30*time.Minute
	switch m.Format {
	case domain.FormatHundred:
		d = 2*time.Hour + 45*time.Minute
	case domain.FormatODI:
		d = 8 * time.Hour
	case domain.FormatTest:
		d = 5 * 24 * time.Hour
	}
	return startOf(m).Add(d)
}

// chatWindow says whether m's chat takes messages now and, before it opens, when it will.
func (v *view) chatWindow(m domain.Match) (open bool, opensAt string) {
	if m.Kind == "classic" {
		return false, ""
	}
	if opens := startOf(m).Add(-chatOpensBefore); v.now.Before(opens) {
		return false, opens.UTC().Format(time.RFC3339)
	}
	if m.State != domain.Completed {
		return true, ""
	}
	return v.now.Before(typicalEnd(m).Add(chatStaysAfter)), ""
}

func (s *Server) handleChat(v *view, w http.ResponseWriter, r *http.Request) {
	m, ok := v.match(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such match")
		return
	}
	after := r.URL.Query().Get("after")
	if after != "" {
		if _, err := time.Parse(time.RFC3339, after); err != nil {
			writeErr(w, http.StatusBadRequest, "bad-after", "after must be a message's createdAt")
			return
		}
	}
	msgs := must(s.Store.ChatMessages(v.ctx, m.ID, after, chatPage))
	users := map[string]domain.User{}
	for _, msg := range msgs {
		if u, ok := v.user(msg.UserID); ok {
			users[u.ID] = u
		}
	}
	open, opensAt := v.chatWindow(m)
	writeJSON(w, http.StatusOK, domain.ChatResponse{Messages: msgs, Users: users, Open: open, OpensAt: opensAt})
}

func (s *Server) handleChatPost(v *view, w http.ResponseWriter, r *http.Request) {
	m, ok := v.match(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such match")
		return
	}
	if open, _ := v.chatWindow(m); !open {
		writeErr(w, http.StatusConflict, "chat-closed", "the chat for this match isn't open")
		return
	}
	var in struct {
		Text string `json:"text"`
	}
	if !decode(w, r, &in) {
		return
	}
	text := strings.TrimSpace(in.Text)
	if text == "" || utf8.RuneCountInString(text) > 500 {
		writeErr(w, http.StatusBadRequest, "bad-text", "a message must be 1–500 characters")
		return
	}
	if last := must(s.Store.LastChatAt(v.ctx, v.me.ID)); last != "" && v.now.Sub(parseTime(last)) < chatGap {
		w.Header().Set("Retry-After", "1")
		writeErr(w, http.StatusTooManyRequests, "slow-down", "one message a second")
		return
	}
	msg := domain.ChatMessage{ID: store.NewID("c-"), MatchID: m.ID, UserID: v.me.ID, Text: text, CreatedAt: v.now.UTC().Format(chatStamp)}
	check(s.Store.AddChatMessage(v.ctx, msg))
	writeJSON(w, http.StatusCreated, msg)
}
