package jobboard

import (
	"context"
	"sync"
	"time"

	"github.com/samber/oops"
)

// RateLimiter enforces per-board submission rate limits to avoid triggering
// anti-automation defenses on ATS platforms. Each board is tracked independently
// using its slug as the key.
type RateLimiter struct {
	boards   map[string]boardState
	defaults RateConfig
	mu       sync.Mutex
}

// boardState tracks the last request time and active cooldown for a single board.
type boardState struct {
	lastRequest time.Time
	cooldownEnd time.Time
}

// NewRateLimiter creates a rate limiter with the given default configuration.
// Individual boards may override these defaults via their BoardMeta.RateLimit.
func NewRateLimiter(defaults RateConfig) *RateLimiter {
	return &RateLimiter{
		boards:   make(map[string]boardState),
		defaults: defaults,
	}
}

// Wait blocks until the rate limit window for the given board has passed.
// It uses the board's own RateConfig if provided, falling back to the limiter
// defaults. Returns an error only if the context is canceled while waiting.
func (rl *RateLimiter) Wait(ctx context.Context, board Board) error {
	slug := board.Meta().Slug
	cfg := rl.configFor(board)
	interval := rl.intervalFromConfig(cfg)

	rl.mu.Lock()
	state := rl.boards[slug]
	now := time.Now()

	// If a cooldown is active (from a previous error), wait for it.
	if now.Before(state.cooldownEnd) {
		wait := time.Until(state.cooldownEnd)
		rl.mu.Unlock()

		return rl.waitDuration(ctx, wait, slug)
	}

	// Normal rate limiting based on requests-per-minute.
	if now.Sub(state.lastRequest) >= interval {
		rl.boards[slug] = boardState{lastRequest: now, cooldownEnd: state.cooldownEnd}
		rl.mu.Unlock()

		return nil
	}

	wait := interval - now.Sub(state.lastRequest)
	rl.mu.Unlock()

	return rl.waitDuration(ctx, wait, slug)
}

// RecordError activates the cooldown period for a board after an error.
func (rl *RateLimiter) RecordError(board Board) {
	slug := board.Meta().Slug
	cfg := rl.configFor(board)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	state := rl.boards[slug]
	state.cooldownEnd = time.Now().Add(cfg.CooldownOnError)
	rl.boards[slug] = state
}

// configFor returns the rate config for a board, falling back to defaults
// when the board has no custom config.
func (rl *RateLimiter) configFor(board Board) RateConfig {
	cfg := board.Meta().RateLimit
	if cfg.RequestsPerMinute == 0 {
		return rl.defaults
	}

	return cfg
}

// intervalFromConfig calculates the minimum time between requests.
func (rl *RateLimiter) intervalFromConfig(cfg RateConfig) time.Duration {
	if cfg.RequestsPerMinute <= 0 {
		return 0
	}

	return time.Minute / time.Duration(cfg.RequestsPerMinute)
}

// waitDuration blocks for the specified duration or until the context is canceled.
func (rl *RateLimiter) waitDuration(ctx context.Context, wait time.Duration, slug string) error {
	select {
	case <-ctx.Done():
		return oops.
			In("jobboard").
			Tags("rate-limit", slug).
			Wrapf(ctx.Err(), "rate limiter wait canceled for board %s", slug)
	case <-time.After(wait):
	}

	rl.mu.Lock()
	state := rl.boards[slug]
	state.lastRequest = time.Now()
	rl.boards[slug] = state
	rl.mu.Unlock()

	return nil
}
