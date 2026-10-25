package auth

import (
	"net"
	"net/http"
	"net/netip"
	"sync"
	"time"
)

// windowLimiter allows at most n events per key in any rolling window. It lives in memory, so a restart forgets
// it. That is fine for the per-IP cap, which only slows one source down; the per-number cap that protects people's
// phones is stored in the database.
type windowLimiter struct {
	mu     sync.Mutex
	n      int
	window time.Duration
	hits   map[string][]time.Time // oldest first
}

func newWindowLimiter(n int, window time.Duration) *windowLimiter {
	return &windowLimiter{n: n, window: window, hits: map[string][]time.Time{}}
}

// sweepAt is how many keys may pile up before stale ones are cleared, so the map can't grow without bound.
const sweepAt = 4096

// allow records an event for key at now if fewer than n happened in the window before it. Otherwise it records
// nothing and returns how long until the oldest event ages out.
func (l *windowLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cut := now.Add(-l.window)
	hs := l.hits[key]
	i := 0
	for i < len(hs) && !hs[i].After(cut) {
		i++
	}
	hs = hs[i:]
	if len(hs) >= l.n {
		l.hits[key] = hs
		return false, hs[0].Add(l.window).Sub(now)
	}
	l.hits[key] = append(hs, now)
	if len(l.hits) > sweepAt {
		for k, hs := range l.hits {
			if !hs[len(hs)-1].After(cut) {
				delete(l.hits, k)
			}
		}
	}
	return true, 0
}

// ClientIP is the host part of r.RemoteAddr. X-Forwarded-For is ignored on purpose: any client can set it. Behind a
// trusted reverse proxy this would have to read the proxy's header instead.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ipKey is the rate-limit key for an address. IPv6 addresses are grouped by /64, the block a single subscriber
// (often a single phone) is usually given, so stepping through it doesn't dodge the cap.
func ipKey(ip string) string {
	a, err := netip.ParseAddr(ip)
	if err != nil {
		return ip
	}
	a = a.Unmap()
	if a.Is6() {
		if p, err := a.WithZone("").Prefix(64); err == nil {
			return p.String()
		}
	}
	return a.String()
}
