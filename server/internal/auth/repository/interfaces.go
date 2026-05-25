package repository

import (
	"context"
	"time"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// UserSessionRepository defines the interface for session persistence operations.
type UserSessionRepository interface {
	// CreateSession creates a new user session.
	// Returns the created session or error.
	CreateSession(ctx context.Context, params *dbqueries.CreateSessionParams) (*dbqueries.UserSession, error)

	// GetByID retrieves a session by its ID.
	// Returns nil if not found.
	GetByID(ctx context.Context, id string) (*dbqueries.UserSession, error)

	// GetBySessionID retrieves a session by session_id (the unique session identifier).
	// Returns nil if not found.
	GetBySessionID(ctx context.Context, sessionID string) (*dbqueries.UserSession, error)

	// GetByTokenHash retrieves a session by token_hash.
	// Returns nil if not found.
	GetByTokenHash(ctx context.Context, tokenHash string) (*dbqueries.UserSession, error)

	// RevokeSession marks a session as revoked (sets revoked_at).
	// Returns error if operation fails.
	RevokeSession(ctx context.Context, sessionID string) error

	// DeleteExpiredSessions soft-deletes expired or revoked sessions.
	// Returns error if operation fails.
	DeleteExpiredSessions(ctx context.Context) error

	// CountActiveSessions returns the count of active (non-revoked, non-expired) sessions for a user.
	CountActiveSessions(ctx context.Context, userID string) (int64, error)

	// GetUserSessions retrieves all sessions for a user (not soft-deleted).
	// Returns empty slice if none found.
	GetUserSessions(ctx context.Context, userID string) ([]*dbqueries.UserSession, error)
}

// LoginAttemptRepository defines the interface for tracking failed login attempts.
type LoginAttemptRepository interface {
	// CountRecentFailed returns the number of failed login attempts for a username
	// that occurred after `since`.
	CountRecentFailed(ctx context.Context, username string, since time.Time) (int64, error)

	// RecordFailure records a failed login attempt for a username.
	// `id` is a new unique UUID for the record.
	RecordFailure(ctx context.Context, id, username string) error

	// DeleteOldAttempts removes records older than 7 days.
	DeleteOldAttempts(ctx context.Context) error
}
