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

// CompanyBuildingRepositoryInterface defines the interface for company production building operations
type CompanyBuildingRepositoryInterface interface {
	CreateCompanyProductionBuilding(ctx context.Context, companyID, productionBuildingID string, constructionEndsAt time.Time) (*models.CompanyProductionBuilding, error)
	GetCompanyProductionBuildingByID(ctx context.Context, buildingID string) (*models.CompanyProductionBuilding, error)
	GetCompanyProductionBuildingsByCompanyID(ctx context.Context, companyID string) ([]*models.CompanyProductionBuilding, error)
	UpdateBuildingLevel(ctx context.Context, buildingID string, newLevel int64, constructionEndsAt time.Time) error
}

// CompanyBuildingRepository handles database operations for company production buildings
type CompanyBuildingRepository struct {
	queries *dbqueries.Queries
}

// NewCompanyBuildingRepository creates a new CompanyBuildingRepository
func NewCompanyBuildingRepository(queries *dbqueries.Queries) *CompanyBuildingRepository {
	return &CompanyBuildingRepository{queries: queries}
}

// CreateCompanyProductionBuilding inserts a new company production building at level 1
func (r *CompanyBuildingRepository) CreateCompanyProductionBuilding(ctx context.Context, companyID, productionBuildingID string, constructionEndsAt time.Time) (*models.CompanyProductionBuilding, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	endsAt := constructionEndsAt

	err := r.queries.CreateCompanyProductionBuilding(ctx, dbqueries.CreateCompanyProductionBuildingParams{
		ID:                   id,
		CompanyID:            companyID,
		ProductionBuildingID: productionBuildingID,
		Level:                1,
		ConstructionEndsAt:   &endsAt,
		CreatedAt:            now,
	})
	if err != nil {
		log.Error().Err(err).Str("company_id", companyID).Msg("Failed to create company production building")
		return nil, err
	}

	return &models.CompanyProductionBuilding{
		ID:                   id,
		CompanyID:            companyID,
		ProductionBuildingID: productionBuildingID,
		Level:                1,
		ConstructionEndsAt:   &endsAt,
		CreatedAt:            now,
	}, nil
}

// GetCompanyProductionBuildingByID retrieves a company production building by its ID
func (r *CompanyBuildingRepository) GetCompanyProductionBuildingByID(ctx context.Context, buildingID string) (*models.CompanyProductionBuilding, error) {
	dbo, err := r.queries.GetCompanyProductionBuildingByID(ctx, buildingID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Error().Err(err).Str("building_id", buildingID).Msg("Failed to get company production building")
		return nil, err
	}
	return models.FromDBCompanyProductionBuilding(&dbo), nil
}

// GetCompanyProductionBuildingsByCompanyID retrieves all non-deleted buildings for a company
func (r *CompanyBuildingRepository) GetCompanyProductionBuildingsByCompanyID(ctx context.Context, companyID string) ([]*models.CompanyProductionBuilding, error) {
	dbos, err := r.queries.GetCompanyProductionBuildingsByCompanyID(ctx, companyID)
	if err != nil {
		log.Error().Err(err).Str("company_id", companyID).Msg("Failed to get company production buildings")
		return nil, err
	}

	buildings := make([]*models.CompanyProductionBuilding, 0, len(dbos))
	for _, dbo := range dbos {
		d := dbo
		buildings = append(buildings, models.FromDBCompanyProductionBuilding(&d))
	}
	return buildings, nil
}

// UpdateBuildingLevel updates the level and construction end time of a building
func (r *CompanyBuildingRepository) UpdateBuildingLevel(ctx context.Context, buildingID string, newLevel int64, constructionEndsAt time.Time) error {
	endsAt := constructionEndsAt
	err := r.queries.UpdateCompanyProductionBuildingLevel(ctx, dbqueries.UpdateCompanyProductionBuildingLevelParams{
		ID:                 buildingID,
		Level:              newLevel,
		ConstructionEndsAt: &endsAt,
	})
	if err != nil {
		log.Error().Err(err).Str("building_id", buildingID).Int64("new_level", newLevel).Msg("Failed to update building level")
	}
	return err
}
