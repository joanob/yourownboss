package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
	"github.com/joanob/yourownboss/internal/production/models"
)

// ProductionRunRepositoryInterface defines the interface for production run operations
type ProductionRunRepositoryInterface interface {
	CreateProductionRun(ctx context.Context, companyBuildingID, processID string, cycles int64, startedAt, endsAt time.Time) (*models.ProductionRun, error)
	GetActiveRunByBuildingID(ctx context.Context, buildingID string) (*models.ProductionRun, error)
	GetActiveRunsByBuildingIDs(ctx context.Context, buildingIDs []string) ([]*models.ProductionRun, error)
	GetProductionRunByID(ctx context.Context, runID string) (*models.ProductionRun, error)
	MarkRunCollected(ctx context.Context, runID string) error
}

// ProductionRunRepository handles database operations for production runs
type ProductionRunRepository struct {
	queries *dbqueries.Queries
}

// NewProductionRunRepository creates a new ProductionRunRepository
func NewProductionRunRepository(queries *dbqueries.Queries) *ProductionRunRepository {
	return &ProductionRunRepository{queries: queries}
}

// CreateProductionRun inserts a new production run record
func (r *ProductionRunRepository) CreateProductionRun(ctx context.Context, companyBuildingID, processID string, cycles int64, startedAt, endsAt time.Time) (*models.ProductionRun, error) {
	id := uuid.New().String()

	err := r.queries.CreateProductionRun(ctx, dbqueries.CreateProductionRunParams{
		ID:                id,
		CompanyBuildingID: companyBuildingID,
		ProcessID:         processID,
		ProductionCycles:  cycles,
		StartedAt:         startedAt,
		EndsAt:            endsAt,
	})
	if err != nil {
		log.Error().Err(err).Str("building_id", companyBuildingID).Msg("Failed to create production run")
		return nil, err
	}

	return &models.ProductionRun{
		ID:                id,
		CompanyBuildingID: companyBuildingID,
		ProcessID:         processID,
		ProductionCycles:  cycles,
		StartedAt:         startedAt,
		EndsAt:            endsAt,
		IsCollected:       false,
	}, nil
}

// GetActiveRunByBuildingID returns the active (uncollected) run for a building, or nil
func (r *ProductionRunRepository) GetActiveRunByBuildingID(ctx context.Context, buildingID string) (*models.ProductionRun, error) {
	dbo, err := r.queries.GetActiveProductionRunByBuildingID(ctx, buildingID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Error().Err(err).Str("building_id", buildingID).Msg("Failed to get active production run")
		return nil, err
	}
	return models.FromDBProductionRun(&dbo), nil
}

// GetActiveRunsByBuildingIDs returns all active (uncollected) runs for the given buildings in one query.
func (r *ProductionRunRepository) GetActiveRunsByBuildingIDs(ctx context.Context, buildingIDs []string) ([]*models.ProductionRun, error) {
	dbos, err := r.queries.GetActiveProductionRunsByBuildingIDs(ctx, buildingIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get active production runs by building IDs")
		return nil, err
	}
	runs := make([]*models.ProductionRun, 0, len(dbos))
	for i := range dbos {
		runs = append(runs, models.FromDBProductionRun(&dbos[i]))
	}
	return runs, nil
}

// GetProductionRunByID retrieves a production run by its ID
func (r *ProductionRunRepository) GetProductionRunByID(ctx context.Context, runID string) (*models.ProductionRun, error) {
	dbo, err := r.queries.GetProductionRunByID(ctx, runID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Error().Err(err).Str("run_id", runID).Msg("Failed to get production run")
		return nil, err
	}
	return models.FromDBProductionRun(&dbo), nil
}

// MarkRunCollected marks a production run as collected with the current UTC time
func (r *ProductionRunRepository) MarkRunCollected(ctx context.Context, runID string) error {
	now := time.Now().UTC()
	err := r.queries.MarkProductionRunCollected(ctx, dbqueries.MarkProductionRunCollectedParams{
		ID:          runID,
		CollectedAt: &now,
	})
	if err != nil {
		log.Error().Err(err).Str("run_id", runID).Msg("Failed to mark production run as collected")
	}
	return err
}
