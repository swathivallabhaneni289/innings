package httpapi

import (
	"net/http"
	"slices"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// Live calls (predictions during a match) come from the World, which opens, locks and settles them from the live
// score; real fans' picks on them live in the store.

func (v *view) callPicks() []domain.CallPick {
	if v.callPickList == nil {
		v.callPickList = must(v.s.Store.CallPicks(v.ctx))
	}
	return v.callPickList
}

func (v *view) callView(c domain.LiveCall) domain.CallView {
	cv := domain.CallView{Call: normCall(c), Counts: map[string]int{}}
	for _, o := range c.Options {
		cv.Counts[o.ID] = 0
	}
	for _, p := range v.callPicks() {
		if p.CallID != c.ID {
			continue
		}
		cv.Counts[p.OptionID]++
		if p.UserID == v.me.ID {
			mine := p.OptionID
			cv.Mine = &mine
		}
	}
	if cv.Mine != nil && c.State == domain.CallSettled && c.AnswerID != nil && *c.AnswerID == *cv.Mine {
		cv.Points = domain.PointsLiveCall
	}
	return cv
}

func (v *view) callViews(m domain.Match) []domain.CallView {
	out := []domain.CallView{}
	for _, c := range v.s.World.Calls(m.ID) {
		out = append(out, v.callView(c))
	}
	return out
}

// callPoints sums userID's points from live calls on matches that finished at or after since.
func (v *view) callPoints(userID string, since time.Time) int {
	n := 0
	for _, p := range v.callPicks() {
		if p.UserID != userID {
			continue
		}
		c, ok := v.s.World.Call(p.CallID)
		if !ok || c.State != domain.CallSettled || c.AnswerID == nil || *c.AnswerID != p.OptionID {
			continue
		}
		if m, ok := v.match(c.MatchID); ok && !v.endOf(m).Before(since) {
			n += domain.PointsLiveCall
		}
	}
	return n
}

func normCall(c domain.LiveCall) domain.LiveCall {
	c.Options = nonNil(c.Options)
	return c
}

func (s *Server) handleCallPick(v *view, w http.ResponseWriter, r *http.Request) {
	m, ok := v.match(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "not-found", "no such match")
		return
	}
	c, ok := s.World.Call(r.PathValue("callId"))
	if !ok || c.MatchID != m.ID {
		writeErr(w, http.StatusNotFound, "not-found", "no such call on this match")
		return
	}
	var in struct {
		OptionID string `json:"optionId"`
	}
	if !decode(w, r, &in) {
		return
	}
	if !slices.ContainsFunc(c.Options, func(o domain.CallOption) bool { return o.ID == in.OptionID }) {
		writeErr(w, http.StatusBadRequest, "bad-option", "pick one of the call's options")
		return
	}
	// The call as it stands now: a pick that arrives after it locks is refused.
	if c.State != domain.CallOpen {
		writeErr(w, http.StatusConflict, "locked", "this call has locked")
		return
	}
	check(s.Store.SaveCallPick(v.ctx, domain.CallPick{UserID: v.me.ID, CallID: c.ID, MatchID: m.ID, OptionID: in.OptionID, CreatedAt: v.stamp()}))
	v.callPickList = nil
	writeJSON(w, http.StatusOK, v.callView(c))
}
