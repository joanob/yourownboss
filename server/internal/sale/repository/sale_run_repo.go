package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
	"github.com/joanob/yourownboss/internal/sale/models"
)

// SaleRunRepositoryInterface defines the contract for sale run operations
type SaleRunRepositoryInterface interface {
	CreateSaleRun(ctx context.Context, companySaleBuildingID, resourceID string, unitsToSell int64, startedAt, endsAt time.Time) (*models.SaleRun, error)
	GetSaleRunByID(ctx context.Context, id string) (*models.SaleRun, error)
	GetActiveRunByBuildingID(ctx context.Context, companySaleBuildingID string) (*models.SaleRun, error)
	GetActiveRunsByBuildingIDs(ctx context.Context, buildingIDs []string) ([]*models.SaleRun, error)
	MarkRunCollected(ctx context.Context, id string, collectedAt time.Time) error
}

// SaleRunRepository handles sale run database operations
type SaleRunRepository struct {
	queries *dbqueries.Queries
}

// NewSaleRunRepository creates a new sale run repository
func NewSaleRunRepository(queries *dbqueries.Queries) *SaleRunRepository {
	return &SaleRunRepository{queries: queries}
}

// CreateSaleRun inserts a new sale run record
func (r *SaleRunRepository) CreateSaleRun(ctx context.Context, companySaleBuildingID, resourceID string, unitsToSell int64, startedAt, endsAt time.Time) (*models.SaleRun, error) {
	id := uuid.New().String()

	err := r.queries.CreateSaleRun(ctx, dbqueries.CreateSaleRunParams{
		ID:                    id,
		CompanySaleBuildingID: companySaleBuildingID,
		ResourceID:            resourceID,
		UnitsToSell:           unitsToSell,
		StartedAt:             startedAt,
		EndsAt:                endsAt,
	})
	if err != nil {
		log.Error().Err(err).Str("building_id", companySaleBuildingID).Msg("Failed to create sale run")
		return nil, err
	}

	return &models.SaleRun{
		ID:                    id,
		CompanySaleBuildingID: companySaleBuildingID,
		ResourceID:            resourceID,
		UnitsToSell:           unitsToSell,
		StartedAt:             startedAt,
		EndsAt:                endsAt,
		IsCollected:           false,
	}, nil
}

// GetSaleRunByID retrieves a sale run by ID
func (r *SaleRunRepository) GetSaleRunByID(ctx context.Context, id string) (*models.SaleRun, error) {
	logger := log.With().Str("run_id", id).Logger()

	dbo, err := r.queries.GetSaleRunByID(ctx, id)
	if err == sql.ErrNoRows {
		logger.Debug().Msg("Sale run not found")
		return nil, nil
	}
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get sale run")
		return nil, err
	}

	return models.FromDBSaleRun(&dbo), nil
}

// GetActiveRunByBuildingID retrieves the active (non-collected) sale run for a building
func (r *SaleRunRepository) GetActiveRunByBuildingID(ctx context.Context, companySaleBuildingID string) (*models.SaleRun, error) {
	logger := log.With().Str("building_id", companySaleBuildingID).Logger()

	dbo, err := r.queries.GetActiveSaleRunByBuildingID(ctx, companySaleBuildingID)
	if err == sql.ErrNoRows {
		logger.Debug().Msg("No active sale run found")
		return nil, nil
	}
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get active sale run")
		return nil, err
	}

	return models.FromDBSaleRun(&dbo), nil
}

// GetActiveRunsByBuildingIDs returns all active sale runs for the given buildings in one query.
func (r *SaleRunRepository) GetActiveRunsByBuildingIDs(ctx context.Context, buildingIDs []string) ([]*models.SaleRun, error) {
	dbos, err := r.queries.GetActiveSaleRunsByBuildingIDs(ctx, buildingIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get active sale runs by building IDs")
		return nil, err
	}
	runs := make([]*models.SaleRun, 0, len(dbos))
	for i := range dbos {
		runs = append(runs, models.FromDBSaleRun(&dbos[i]))
	}
	return runs, nil
}

// MarkRunCollected marks a sale run as collected
func (r *SaleRunRepository) MarkRunCollected(ctx context.Context, id string, collectedAt time.Time) error {
	logger := log.With().Str("run_id", id).Logger()

	t := collectedAt
	err := r.queries.MarkSaleRunCollected(ctx, dbqueries.MarkSaleRunCollectedParams{
		ID:          id,
		CollectedAt: &t,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to mark sale run collected")
		return err
	}

	logger.Debug().Msg("Sale run marked as collected")
	return nil
}
