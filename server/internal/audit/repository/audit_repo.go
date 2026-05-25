package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// AuditRepositoryInterface defines the contract for audit logging operations
type AuditRepositoryInterface interface {
	Log(ctx context.Context, userID, companyID, action, resourceType, resourceID string, changes interface{}) error
}

// AuditRepository handles audit log database operations
type AuditRepository struct {
	queries *dbqueries.Queries
}

// NewAuditRepository creates a new AuditRepository
func NewAuditRepository(queries *dbqueries.Queries) *AuditRepository {
	return &AuditRepository{queries: queries}
}

// Log creates a new audit log entry. Errors are logged but never propagated to
// avoid blocking normal business operations.
func (r *AuditRepository) Log(ctx context.Context, userID, companyID, action, resourceType, resourceID string, changes interface{}) error {
	var userIDPtr *string
	if userID != "" {
		v := userID
		userIDPtr = &v
	}

	var companyIDPtr *string
	if companyID != "" {
		v := companyID
		companyIDPtr = &v
	}

	var resourceTypePtr *string
	if resourceType != "" {
		v := resourceType
		resourceTypePtr = &v
	}

	var resourceIDPtr *string
	if resourceID != "" {
		v := resourceID
		resourceIDPtr = &v
	}

	var changesPtr *string
	if changes != nil {
		b, err := json.Marshal(changes)
		if err != nil {
			log.Warn().Err(err).Msg("audit: failed to marshal changes")
		} else {
			s := string(b)
			changesPtr = &s
		}
	}

	err := r.queries.CreateAuditLog(ctx, dbqueries.CreateAuditLogParams{
		ID:           uuid.New().String(),
		UserID:       userIDPtr,
		CompanyID:    companyIDPtr,
		Action:       action,
		ResourceType: resourceTypePtr,
		ResourceID:   resourceIDPtr,
		Changes:      changesPtr,
		Timestamp:    time.Now().UTC(),
	})
	if err != nil {
		log.Error().Err(err).Str("action", action).Msg("audit: failed to write log entry")
	}
	return err
}
