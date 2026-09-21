package cloudconnexa

import (
	"context"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// DefaultMaxRetries is how many times a request rejected with HTTP 429 Too Many Requests
// is retried before the error is returned to the caller.
const DefaultMaxRetries = 10

// Bounds for the wait between retries. Variables rather than constants so tests can shrink them.
var (
	retryBaseWait = 500 * time.Millisecond
	retryMaxWait  = 20 * time.Second
)

// retryPause schedules retries for one server-side rate-limit bucket. Every request rejected
// with 429 reserves the next free slot, one wait after the previous slot, so retries reach the
// server spaced at its replenish rate instead of all at once. A new request issued while slots
// are outstanding waits for the last one, queueing behind the retries rather than competing
// with them for the same token.
type retryPause struct {
	mu   sync.Mutex
	last time.Time
}

// reserve claims the slot wait after the last reserved slot (or after now, if that is later)
// and returns it.
func (p *retryPause) reserve(wait time.Duration) time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()
	start := time.Now()
	if p.last.After(start) {
		start = p.last
	}
	p.last = start.Add(wait)
	return p.last
}

// deadline returns the last reserved slot.
func (p *retryPause) deadline() time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.last
}

// waitUntil blocks until t has passed or ctx is done.
func waitUntil(ctx context.Context, t time.Time) error {
	wait := time.Until(t)
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// retryWait returns how long to wait before retrying a request rejected with 429.
// It waits at least as long as the server asks through Retry-After or the
// X-RateLimit-Replenish-* headers, and never less than an exponential backoff with jitter,
// so the wait keeps growing under sustained contention. attempt is zero-based.
func retryWait(h http.Header, attempt int) time.Duration {
	wait := jitteredBackoff(attempt)
	if hint, ok := serverRetryHint(h); ok && hint > wait {
		wait = hint
	}
	return min(wait, retryMaxWait)
}

// serverRetryHint extracts the wait the server asks for from a 429 response.
// It reads Retry-After (seconds or HTTP-date) first, then the time to replenish one token
// from X-RateLimit-Replenish-Rate and X-RateLimit-Replenish-Time.
func serverRetryHint(h http.Header) (time.Duration, bool) {
	if v := h.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second, true
		}
		if t, err := http.ParseTime(v); err == nil {
			return max(time.Until(t), 0), true
		}
	}
	rateValue, errRate := strconv.Atoi(h.Get("X-RateLimit-Replenish-Rate"))
	timeValue, errTime := strconv.Atoi(h.Get("X-RateLimit-Replenish-Time"))
	if errRate == nil && errTime == nil && rateValue > 0 && timeValue > 0 {
		return time.Duration(timeValue) * time.Second / time.Duration(rateValue), true
	}
	return 0, false
}

// jitteredBackoff returns retryBaseWait doubled attempt times and capped at retryMaxWait,
// keeping the lower half and randomizing the upper half so concurrent callers spread out.
func jitteredBackoff(attempt int) time.Duration {
	d := retryBaseWait
	for i := 0; i < attempt && d < retryMaxWait; i++ {
		d *= 2
	}
	d = min(d, retryMaxWait)
	half := d / 2
	return half + rand.N(half+1) // #nosec G404 -- jitter does not need cryptographic randomness
}
