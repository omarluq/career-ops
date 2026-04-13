package applicator

import (
	"context"
	"sync"
	"time"

	"github.com/samber/oops"
)

// RateLimiter enforces per-portal submission rate limits to avoid
// triggering anti-automation defenses on ATS platforms.
type RateLimiter struct {
	portals  map[PortalType]time.Time
	interval time.Duration
	mu       sync.Mutex
}

// NewRateLimiter creates a rate limiter that enforces at least interval
// between consecutive submissions to the same portal.
func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		portals:  make(map[PortalType]time.Time),
		interval: interval,
	}
}

// Wait blocks until the portal's rate limit window has passed.
// Returns an error only if the context is canceled while waiting.
func (rl *RateLimiter) Wait(ctx context.Context, portal PortalType) error {
	rl.mu.Lock()
	last, ok := rl.portals[portal]
	now := time.Now()

	if !ok || now.Sub(last) >= rl.interval {
		rl.portals[portal] = now
		rl.mu.Unlock()
		return nil
	}

	wait := rl.interval - now.Sub(last)
	rl.mu.Unlock()

	select {
	case <-ctx.Done():
		return oops.Wrapf(ctx.Err(), "rate limiter wait canceled for portal %s", portal)
	case <-time.After(wait):
	}

	rl.mu.Lock()
	rl.portals[portal] = time.Now()
	rl.mu.Unlock()

	return nil
}
