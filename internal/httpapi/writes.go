package httpapi

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/swathivallabhaneni289/innings/internal/domain"
	"github.com/swathivallabhaneni289/innings/internal/store"
)

func (v *view) stamp() string { return v.now.UTC().Format(time.RFC3339) }

// ---- logs ----

func (s *Server) handleCreateLog(v *view, w http.ResponseWriter, r *http.Request) {
	var in domain.MatchLog // NewLog: id, userId, createdAt and stubNo are ignored
	if !decode(w, r, &in) {
		return
	}
	m, ok := v.match(in.MatchID)
	if !ok {
		writeErr(w, http.StatusBadRequest, "unknown-match", "no such match")
		return
	}
	if m.State == domain.Upcoming {
		writeErr(w, http.StatusBadRequest, "not-started", "you can't log a match before it starts")
		return
	}
	if !domain.ValidRating(in.Rating) {
		writeErr(w, http.StatusBadRequest, "bad-rating", "rating must be 0.5–5 in half-ball steps")
		return
	}
	if !domain.ValidWatchPlace(in.WatchedAt) {
		writeErr(w, http.StatusBadRequest, "bad-watched-at", "watchedAt must be a known place")
		return
	}
	xi := xiPlayers(m)
	for pid, rt := range in.PlayerRatings {
		if rt < 1 || rt > 10 {
			writeErr(w, http.StatusBadRequest, "bad-player-rating", "player ratings are 1–10")
			return
		}
		if len(xi) > 0 && !xi[pid] {
			writeErr(w, http.StatusBadRequest, "bad-player", "player "+pid+" isn't in this match")
			return
		}
	}
	if in.FanPotmID != nil && len(xi) > 0 && !xi[*in.FanPotmID] {
		writeErr(w, http.StatusBadRequest, "bad-player", "fan Player of the Match isn't in this match")
		return
	}
	if in.Companions != nil && *in.Companions < 0 {
		writeErr(w, http.StatusBadRequest, "bad-companions", "companions can't be negative")
		return
	}
	if utf8.RuneCountInString(in.Review) > 2000 || utf8.RuneCountInString(in.PlaceNote) > 200 {
		writeErr(w, http.StatusBadRequest, "too-long", "review max 2000 characters, place note max 200")
		return
	}
	in.ID, in.UserID, in.CreatedAt = store.NewID("log-"), v.me.ID, v.stamp()
	in.Review = strings.TrimSpace(in.Review)
	saved, _, err := s.Store.SaveLog(v.ctx, normLog(in))
	check(err)
	fi, _ := s.newView(v.ctx, v.me).feedItem(saved)
	writeJSON(w, http.StatusCreated, fi)
}

func (s *Server) handleDeleteLog(v *view, w http.ResponseWriter, r *http.Request) {
	if !must(s.Store.DeleteLog(v.ctx, v.me.ID, r.PathValue("id"))) {
		writeErr(w, http.StatusNotFound, "not-found", "no such log of yours")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- predictions & watching ----

func (s *Server) handlePrediction(v *view, w http.ResponseWriter, r *http.Request) {
	m, ok := v.match(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such match")
		return
	}
	if domain.IsLocked(m, v.now) {
		writeErr(w, http.StatusConflict, "locked", "predictions lock at the first ball")
		return
	}
	var in struct {
		WinnerID         string `json:"winnerId"`
		TopRunScorerID   string `json:"topRunScorerId"`
		TopWicketTakerID string `json:"topWicketTakerId"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.WinnerID != m.TeamIDs[0] && in.WinnerID != m.TeamIDs[1] {
		writeErr(w, http.StatusBadRequest, "bad-pick", "winner must be one of the two teams")
		return
	}
	inMatch := func(pid string) bool {
		if xi := xiPlayers(m); len(xi) > 0 {
			return xi[pid]
		}
		p, ok := s.World.Players()[pid]
		return ok && (p.TeamID == m.TeamIDs[0] || p.TeamID == m.TeamIDs[1])
	}
	if !inMatch(in.TopRunScorerID) || !inMatch(in.TopWicketTakerID) {
		writeErr(w, http.StatusBadRequest, "bad-pick", "top scorer and top wicket-taker must play in this match")
		return
	}
	p := domain.Prediction{UserID: v.me.ID, MatchID: m.ID, WinnerID: in.WinnerID, TopRunScorerID: in.TopRunScorerID, TopWicketTakerID: in.TopWicketTakerID, CreatedAt: v.stamp()}
	check(s.Store.SavePrediction(v.ctx, p))
	writeJSON(w, http.StatusOK, map[string]any{"prediction": p})
}

func (s *Server) handleWatching(v *view, w http.ResponseWriter, r *http.Request) {
	m, ok := v.match(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such match")
		return
	}
	var in struct {
		Place *domain.WatchPlace `json:"place"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Place != nil && !domain.ValidWatchPlace(*in.Place) {
		writeErr(w, http.StatusBadRequest, "bad-place", "unknown watch place")
		return
	}
	check(s.Store.SetWatching(v.ctx, v.me.ID, m.ID, in.Place, v.now))
	writeJSON(w, http.StatusOK, map[string]any{"watching": in.Place})
}

// ---- reactions & replies ----

func (s *Server) toggleReaction(v *view, w http.ResponseWriter, r *http.Request, table string) (on bool, count int, ok bool) {
	id := r.PathValue("id")
	if _, found := v.logByID(id); !found {
		writeErr(w, http.StatusNotFound, "not-found", "no such log")
		return false, 0, false
	}
	on = must(s.Store.ToggleReaction(v.ctx, table, id, v.me.ID, v.now))
	count, _ = s.newView(v.ctx, v.me).reactionsOn(table, id)
	return on, count, true
}

func (s *Server) handleLike(v *view, w http.ResponseWriter, r *http.Request) {
	if on, n, ok := s.toggleReaction(v, w, r, store.Likes); ok {
		writeJSON(w, http.StatusOK, map[string]any{"liked": on, "likes": n})
	}
}

func (s *Server) handleThereToo(v *view, w http.ResponseWriter, r *http.Request) {
	if on, n, ok := s.toggleReaction(v, w, r, store.ThereToo); ok {
		writeJSON(w, http.StatusOK, map[string]any{"iWasThere": on, "thereToo": n})
	}
}

func (s *Server) handleReply(v *view, w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, found := v.logByID(id); !found {
		writeErr(w, http.StatusNotFound, "not-found", "no such log")
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
		writeErr(w, http.StatusBadRequest, "bad-text", "reply must be 1–500 characters")
		return
	}
	rp := domain.Reply{ID: store.NewID("r-"), LogID: id, UserID: v.me.ID, Text: text, CreatedAt: v.stamp()}
	check(s.Store.AddReply(v.ctx, rp))
	writeJSON(w, http.StatusCreated, rp)
}

// ---- follows & profile ----

func (s *Server) handleFollow(v *view, w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "me" || id == v.me.ID {
		writeErr(w, http.StatusBadRequest, "self", "you can't follow yourself")
		return
	}
	if _, ok := v.user(id); !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such user")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"following": must(s.Store.ToggleFollow(v.ctx, v.me.ID, id, v.now))})
}

func (s *Server) handleUpdateMe(v *view, w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name            *string `json:"name"`
		Handle          *string `json:"handle"`
		Bio             *string `json:"bio"`
		FavouriteTeamID *string `json:"favouriteTeamId"`
		AvatarColour    *string `json:"avatarColour"`
		PhotoURI        *string `json:"photoUri"`
	}
	if !decode(w, r, &in) {
		return
	}
	u := v.me
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" || utf8.RuneCountInString(name) > 40 {
			writeErr(w, http.StatusBadRequest, "bad-name", "name must be 1–40 characters")
			return
		}
		u.Name = name
	}
	if in.Handle != nil {
		h := domain.NormHandle(*in.Handle)
		switch reason, msg := s.handleProblem(v, h); reason {
		case "":
			u.Handle = h
		case handleTaken:
			writeErr(w, http.StatusConflict, "handle-taken", msg)
			return
		default:
			writeErr(w, http.StatusBadRequest, "bad-handle", msg)
			return
		}
	}
	if in.Bio != nil {
		if utf8.RuneCountInString(*in.Bio) > 160 {
			writeErr(w, http.StatusBadRequest, "bad-bio", "bio max 160 characters")
			return
		}
		u.Bio = strings.TrimSpace(*in.Bio)
	}
	if in.FavouriteTeamID != nil {
		teams := v.teams()
		if _, ok := teams[*in.FavouriteTeamID]; len(teams) > 0 && !ok {
			writeErr(w, http.StatusBadRequest, "bad-team", "unknown team")
			return
		}
		u.FavouriteTeamID = *in.FavouriteTeamID
	}
	if in.AvatarColour != nil {
		u.AvatarColour = *in.AvatarColour
	}
	if in.PhotoURI != nil {
		u.PhotoURI = *in.PhotoURI
	}
	err := s.Store.UpdateUser(v.ctx, u)
	if errors.Is(err, store.ErrHandleTaken) { // someone took it between the check above and this write
		writeErr(w, http.StatusConflict, "handle-taken", "@"+u.Handle+" was just taken.")
		return
	}
	check(err)
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleTop5(v *view, w http.ResponseWriter, r *http.Request) {
	var in struct {
		MatchIDs []string `json:"matchIds"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.MatchIDs) > 5 {
		writeErr(w, http.StatusBadRequest, "too-many", "top 5 holds at most 5 matches")
		return
	}
	out := []domain.Match{}
	for i, id := range in.MatchIDs {
		if slices.Contains(in.MatchIDs[:i], id) {
			writeErr(w, http.StatusBadRequest, "duplicate", "a match can only be pinned once")
			return
		}
		m, ok := v.match(id)
		if !ok {
			writeErr(w, http.StatusBadRequest, "unknown-match", "no such match: "+id)
			return
		}
		out = append(out, m)
	}
	check(s.Store.SetTop5(v.ctx, v.me.ID, in.MatchIDs))
	writeJSON(w, http.StatusOK, map[string]any{"top5": out})
}
