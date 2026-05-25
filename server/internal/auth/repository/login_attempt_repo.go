package repository

import (
	"context"
	"time"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

type loginAttemptRepository struct {
	queries *dbqueries.Queries
}

// NewLoginAttemptRepository creates a new LoginAttemptRepository backed by SQLite.
func NewLoginAttemptRepository(queries *dbqueries.Queries) LoginAttemptRepository {
	return &loginAttemptRepository{queries: queries}
}

// CountRecentFailed returns the number of failed login attempts for username since `since`.
func (r *loginAttemptRepository) CountRecentFailed(ctx context.Context, username string, since time.Time) (int64, error) {
	return r.queries.CountRecentFailedAttempts(ctx, username, since)
}

// RecordFailure inserts a failed login attempt record.
func (r *loginAttemptRepository) RecordFailure(ctx context.Context, id, username string) error {
	return r.queries.InsertLoginAttempt(ctx, id, username)
}
