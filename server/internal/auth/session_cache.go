package auth

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// SessionCache stores active sessions in memory with optional TTL cleanup.
type SessionCache struct {
	sessions map[string]SessionData // key: session_id, value: SessionData
	mu       sync.RWMutex
}

// NewSessionCache creates a new session cache.
func NewSessionCache() *SessionCache {
	return &SessionCache{
		sessions: make(map[string]SessionData),
	}
}

// Store saves a session in the cache.
func (sc *SessionCache) Store(sessionID string, data SessionData) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.sessions[sessionID] = data
	log.Debug().
		Str("session_id", sessionID).
		Str("user_id", data.UserID).
		Time("expires_at", data.ExpiresAt).
		Msg("Session stored in cache")
}

// Get retrieves a session from the cache by session ID.
// Returns nil if session not found or has expired.
func (sc *SessionCache) Get(sessionID string) *SessionData {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	data, exists := sc.sessions[sessionID]
	if !exists {
		log.Debug().Str("session_id", sessionID).Msg("Session not found in cache")
		return nil
	}

	// Check if session has expired
	if time.Now().After(data.ExpiresAt) {
		log.Debug().Str("session_id", sessionID).Msg("Session has expired")
		return nil
	}

	// Check if session was revoked
	if data.RevokedAt != nil {
		log.Debug().Str("session_id", sessionID).Msg("Session has been revoked")
		return nil
	}

	return &data
}

// Revoke marks a session as revoked in the cache.
func (sc *SessionCache) Revoke(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	data, exists := sc.sessions[sessionID]
	if !exists {
		log.Debug().Str("session_id", sessionID).Msg("Session not found in cache to revoke")
		return
	}

	now := time.Now().UTC()
	data.RevokedAt = &now
	sc.sessions[sessionID] = data

	log.Debug().Str("session_id", sessionID).Msg("Session revoked in cache")
}

// Delete removes a session from the cache.
func (sc *SessionCache) Delete(sessionID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.sessions, sessionID)
	log.Debug().Str("session_id", sessionID).Msg("Session deleted from cache")
}

// DeleteExpired removes all expired or revoked sessions from the cache.
// This should be called periodically (e.g., every 6 hours).
func (sc *SessionCache) DeleteExpired() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now().UTC()
	deletedCount := 0

	for sessionID, data := range sc.sessions {
		// Delete if expired or revoked
		if now.After(data.ExpiresAt) || data.RevokedAt != nil {
			delete(sc.sessions, sessionID)
			deletedCount++
		}
	}

	if deletedCount > 0 {
		log.Info().
			Int("deleted_count", deletedCount).
			Int("remaining_sessions", len(sc.sessions)).
			Msg("Expired sessions cleaned from cache")
	}

	return deletedCount
}

// Count returns the number of active sessions in the cache.
func (sc *SessionCache) Count() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	now := time.Now().UTC()
	count := 0

	for _, data := range sc.sessions {
		// Count only non-expired, non-revoked sessions
		if now.Before(data.ExpiresAt) && data.RevokedAt == nil {
			count++
		}
	}

	return count
}
