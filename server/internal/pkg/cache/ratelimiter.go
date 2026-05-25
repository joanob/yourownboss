package cache

import (
	"sync"
	"time"
)

const rateLimitWindow = time.Minute

// RateLimiter implements an in-memory sliding-window rate limiter.
// It tracks successful operations per (userID, action) pair.
type RateLimiter struct {
	mu      sync.Mutex
	records map[string][]time.Time // key: "userID:action" → timestamps of successful ops
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		records: make(map[string][]time.Time),
	}
}

// AllowAndRecord atomically checks the rate limit and, if allowed, records the operation.
// Returns true if the operation is allowed (the operation has been counted).
// PERF-04: combining Allow+Record in one lock avoids the TOCTOU race between two separate calls.
func (rl *RateLimiter) AllowAndRecord(userID, action string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := userID + ":" + action
	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)

	current := rl.records[key]
	var recent []time.Time
	for _, t := range current {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= limit {
		rl.records[key] = recent // store pruned list even on rejection
		return false
	}

	rl.records[key] = append(recent, now)
	return true
}

// Allow returns true if the user has not exceeded `limit` successful operations
// in the last minute for the given action. It prunes stale entries on each call.
// Prefer AllowAndRecord when the operation should be counted on the same call.
func (rl *RateLimiter) Allow(userID, action string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := userID + ":" + action
	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)

	current := rl.records[key]
	var recent []time.Time
	for _, t := range current {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	rl.records[key] = recent

	return len(recent) < limit
}

// Record registers one successful operation for the given (userID, action) pair.
// Call this only after the operation has actually succeeded.
func (rl *RateLimiter) Record(userID, action string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := userID + ":" + action
	rl.records[key] = append(rl.records[key], time.Now())
}

// StartCleanup starts a background goroutine that periodically removes entries
// for users who have stopped making requests, preventing unbounded memory growth.
// SEC-05: without this, idle user entries accumulate forever.
// Call once after NewRateLimiter(). A cleanup interval of 5 minutes is recommended.
func (rl *RateLimiter) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rateLimitWindow)
	for key, timestamps := range rl.records {
		var recent []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(rl.records, key)
		} else {
			rl.records[key] = recent
		}
	}
}
