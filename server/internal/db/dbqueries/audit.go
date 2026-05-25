package dbqueries

import (
	"context"
	"time"
)

// CreateAuditLogParams holds the parameters for inserting an audit log entry
type CreateAuditLogParams struct {
	ID           string
	UserID       *string
	CompanyID    *string
	Action       string
	ResourceType *string
	ResourceID   *string
	Changes      *string
	Timestamp    time.Time
}

// CreateAuditLog inserts a new audit log entry into the audit_log table
func (q *Queries) CreateAuditLog(ctx context.Context, arg CreateAuditLogParams) error {
	const stmt = `
		INSERT INTO audit_log (id, user_id, company_id, action, resource_type, resource_id, changes, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := q.db.ExecContext(ctx, stmt,
		arg.ID,
		arg.UserID,
		arg.CompanyID,
		arg.Action,
		arg.ResourceType,
		arg.ResourceID,
		arg.Changes,
		arg.Timestamp,
	)
	return err
}
