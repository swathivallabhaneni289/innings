package sim

import (
	"fmt"
	"math"
	"time"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

// partial interpolates a scripted innings to fraction frac (0..1) of its balls.
func partial(in inn, frac float64) inn {
	frac = clampF(frac, 0, 1)
	if frac >= 1 {
		return in
	}
	return inn{
		t: in.t,
		r: int(math.Round(float64(in.r) * (0.8*frac + 0.2*frac*frac))),
		w: int(float64(in.w) * math.Pow(frac, 1.1)),
		b: int(float64(in.b) * frac),
	}
}

func scoreText(f *fixture, in inn) string {
	return fmt.Sprintf("%s %d/%d (%s)", teamShort(f.teams[in.t]), in.r, in.w, oversText(f.format, in.b))
}

func (w *World) liveScore(f *fixture, now time.Time) *domain.LiveScore {
	e := now.Sub(w.start(f)).Minutes()
	if f.format == domain.FormatTest {
		return f.liveTest(e)
	}
	return f.liveLimited(e)
}

func (f *fixture) liveLimited(e float64) *domain.LiveScore {
	brk := 15.0
	if f.format == domain.FormatODI {
		brk = 20
	}
	l := (float64(f.dur) - brk - 10) / 2
	in1 := f.inns[0]
	ls := &domain.LiveScore{BattingTeamID: f.teams[in1.t]}

	if f.noResult {
		balls := min(int(e/l*float64(f.maxBalls)), in1.b)
		cur := partial(in1, float64(balls)/float64(in1.b))
		ls.Innings = f.inningsScores([]inn{cur})
		ls.StatusText = scoreText(f, cur)
		if balls >= in1.b {
			ls.StatusText = "Rain stops play · " + ls.StatusText
		}
		return ls
	}

	if e < l {
		balls := int(e / l * float64(f.maxBalls))
		cur := partial(in1, float64(balls)/float64(in1.b))
		cur.b = min(balls, in1.b)
		ls.Innings = f.inningsScores([]inn{cur})
		ls.StatusText = scoreText(f, cur)
		if f.format != domain.FormatHundred && cur.b > 0 {
			ls.StatusText += fmt.Sprintf(" · CRR %.2f", float64(cur.r)*6/float64(cur.b))
		}
		return ls
	}

	in2 := f.inns[1]
	target := in1.r + 1
	ls.BattingTeamID = f.teams[in2.t]
	chaser := teamShort(f.teams[in2.t])
	if e < l+brk {
		ls.Innings = f.inningsScores([]inn{in1})
		ls.StatusText = fmt.Sprintf("Innings break · %s need %d off %d balls", chaser, target, f.maxBalls)
		return ls
	}
	balls := min(int((e-l-brk)/l*float64(f.maxBalls)), in2.b)
	cur := partial(in2, float64(balls)/float64(in2.b))
	cur.b = balls
	if balls < in2.b && cur.r >= target {
		cur.r = target - 1
	}
	ls.Innings = f.inningsScores([]inn{in1, cur})
	if balls >= in2.b {
		ls.StatusText = f.text
		return ls
	}
	ls.StatusText = fmt.Sprintf("%s need %d off %d balls", chaser, target-cur.r, f.maxBalls-cur.b)
	return ls
}

// testBalls returns balls bowled after e minutes (90 overs a day in three sessions) and the session label.
func testBalls(e float64) (balls int, dayNo int, session string) {
	d := int(e / day)
	m := e - float64(d*day)
	played := math.Min(m, 120) + clampF(m-160, 0, 120) + clampF(m-300, 0, 120)
	switch {
	case m < 120:
		session = "Session 1"
	case m < 160:
		session = "Lunch"
	case m < 280:
		session = "Session 2"
	case m < 300:
		session = "Tea"
	case m < 420:
		session = "Session 3"
	default:
		session = "Stumps"
	}
	return d*540 + int(played*1.5), d + 1, session
}

func (f *fixture) liveTest(e float64) *domain.LiveScore {
	bowled, dayNo, session := testBalls(e)
	prefix := fmt.Sprintf("Day %d, %s · ", dayNo, session)
	ls := &domain.LiveScore{}
	cum := 0
	done := []inn{}
	for k, in := range f.inns {
		if bowled < cum+in.b {
			cur := partial(in, float64(bowled-cum)/float64(in.b))
			cur.b = bowled - cum
			ls.Innings = f.inningsScores(append(done, cur))
			ls.BattingTeamID = f.teams[in.t]
			tot := [2]int{}
			for _, d := range done {
				tot[d.t] += d.r
			}
			tot[cur.t] += cur.r
			txt := prefix + fmt.Sprintf("%s %d/%d", teamShort(f.teams[cur.t]), cur.r, cur.w)
			diff := tot[cur.t] - tot[1-cur.t]
			switch {
			case k == 0:
			case k == len(f.inns)-1 && k >= 3:
				txt += fmt.Sprintf(" · need %d to win", -diff+1)
			case diff > 0:
				txt += fmt.Sprintf(" · lead by %d", diff)
			case diff < 0:
				txt += fmt.Sprintf(" · trail by %d", -diff)
			default:
				txt += " · scores level"
			}
			ls.StatusText = txt
			return ls
		}
		cum += in.b
		done = append(done, in)
	}
	ls.Innings = f.inningsScores(f.inns)
	ls.BattingTeamID = f.teams[f.inns[len(f.inns)-1].t]
	ls.StatusText = fmt.Sprintf("Day %d · %s", dayNo, f.text)
	return ls
}
