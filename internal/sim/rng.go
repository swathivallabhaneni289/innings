package sim

import (
	"hash/fnv"
	"math"
)

// rng is a tiny deterministic PRNG (splitmix64) seeded from strings. Never use math/rand's global source.
type rng struct{ s uint64 }

func hash64(parts ...string) uint64 {
	h := fnv.New64a()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return h.Sum64()
}

func newRNG(parts ...string) *rng { return &rng{s: hash64(parts...)} }

func (r *rng) next() uint64 {
	r.s += 0x9e3779b97f4a7c15
	z := r.s
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// intn returns 0..n-1 (0 when n <= 0).
func (r *rng) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next() % uint64(n))
}

// between returns lo..hi inclusive.
func (r *rng) between(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + r.intn(hi-lo+1)
}

func (r *rng) float() float64 { return float64(r.next()>>11) / float64(1<<53) }

func (r *rng) frange(lo, hi float64) float64 { return lo + (hi-lo)*r.float() }

func (r *rng) chance(p float64) bool { return r.float() < p }

func pick[T any](r *rng, xs []T) T { return xs[r.intn(len(xs))] }

// shuffled returns a deterministic permutation copy of xs.
func shuffled[T any](r *rng, xs []T) []T {
	out := append([]T(nil), xs...)
	for i := len(out) - 1; i > 0; i-- {
		j := r.intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func clampF(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func clampI(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// halfStep rounds to the nearest 0.5 within 0.5..5.
func halfStep(v float64) float64 { return clampF(math.Round(v*2)/2, 0.5, 5) }
