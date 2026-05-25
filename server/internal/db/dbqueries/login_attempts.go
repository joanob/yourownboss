package dbqueries

import (
	"context"
	"time"
)

const countRecentFailedAttempts = `
SELECT COUNT(*) FROM login_attempts
WHERE username = ? AND failed_at > ? AND is_deleted = 0
`

// CountRecentFailedAttempts returns the number of failed login attempts
// for a given username since the `since` timestamp.
func (q *Queries) CountRecentFailedAttempts(ctx context.Context, username string, since time.Time) (int64, error) {
	row := q.db.QueryRowContext(ctx, countRecentFailedAttempts, username, since)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

const insertLoginAttempt = `
INSERT INTO login_attempts (id, username, failed_at, is_deleted)
VALUES (?, ?, CURRENT_TIMESTAMP, 0)
`

// InsertLoginAttempt records a failed login attempt for the given username.
func (q *Queries) InsertLoginAttempt(ctx context.Context, id, username string) error {
	_, err := q.db.ExecContext(ctx, insertLoginAttempt, id, username)
	return err
}
