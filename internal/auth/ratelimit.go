package auth

import (
	"net"
	"net/http"
	"net/netip"
	"sync"
	"time"
)

// windowLimiter allows at most n events per key in any rolling window. It lives in memory, so a restart forgets
// it. That is fine for the per-IP and per-account caps here, which only slow one source down; the per-address cap
// that protects people's inboxes is stored in the database.
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
	if wait := l.waitLocked(key, now); wait > 0 {
		return false, wait
	}
	l.recordLocked(key, now)
	return true, 0
}

// wait is how long until key may have another event: 0 while it is under the cap. It records nothing, for caps
// that count only failures (see record).
func (l *windowLimiter) wait(key string, now time.Time) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.waitLocked(key, now)
}

// record counts an event for key at now, whether or not the cap is reached.
func (l *windowLimiter) record(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.recordLocked(key, now)
}

// forget clears key's events (a right password wipes the slate for that account).
func (l *windowLimiter) forget(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.hits, key)
}

func (l *windowLimiter) waitLocked(key string, now time.Time) time.Duration {
	cut := now.Add(-l.window)
	hs := l.hits[key]
	i := 0
	for i < len(hs) && !hs[i].After(cut) {
		i++
	}
	if hs = hs[i:]; len(hs) == 0 {
		delete(l.hits, key)
		return 0
	}
	l.hits[key] = hs
	if len(hs) >= l.n {
		return hs[len(hs)-l.n].Add(l.window).Sub(now)
	}
	return 0
}

func (l *windowLimiter) recordLocked(key string, now time.Time) {
	l.hits[key] = append(l.hits[key], now)
	if len(l.hits) > sweepAt {
		cut := now.Add(-l.window)
		for k, hs := range l.hits {
			if !hs[len(hs)-1].After(cut) {
				delete(l.hits, k)
			}
		}
	}
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
