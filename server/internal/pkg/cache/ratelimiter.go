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

// Allow returns true if the user has not exceeded `limit` successful operations
// in the last minute for the given action.  It prunes stale entries on each call.
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
