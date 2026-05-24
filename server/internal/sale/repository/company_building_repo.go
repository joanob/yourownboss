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

// CompanySaleBuildingRepositoryInterface defines the contract for company sale building operations
type CompanySaleBuildingRepositoryInterface interface {
	CreateCompanySaleBuilding(ctx context.Context, companyID, saleBuildingID string, constructionEndsAt time.Time) (*models.CompanySaleBuilding, error)
	GetCompanySaleBuildingByID(ctx context.Context, id string) (*models.CompanySaleBuilding, error)
	GetCompanySaleBuildingsByCompanyID(ctx context.Context, companyID string) ([]*models.CompanySaleBuilding, error)
	UpdateBuildingLevel(ctx context.Context, id string, level int64, constructionEndsAt time.Time) error
}

// CompanySaleBuildingRepository handles company sale building database operations
type CompanySaleBuildingRepository struct {
	queries *dbqueries.Queries
}

// NewCompanySaleBuildingRepository creates a new company sale building repository
func NewCompanySaleBuildingRepository(queries *dbqueries.Queries) *CompanySaleBuildingRepository {
	return &CompanySaleBuildingRepository{queries: queries}
}

// CreateCompanySaleBuilding inserts a new company sale building at level 1
func (r *CompanySaleBuildingRepository) CreateCompanySaleBuilding(ctx context.Context, companyID, saleBuildingID string, constructionEndsAt time.Time) (*models.CompanySaleBuilding, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	endsAt := constructionEndsAt

	err := r.queries.CreateCompanySaleBuilding(ctx, dbqueries.CreateCompanySaleBuildingParams{
		ID:                 id,
		CompanyID:          companyID,
		SaleBuildingID:     saleBuildingID,
		Level:              1,
		ConstructionEndsAt: &endsAt,
		CreatedAt:          now,
	})
	if err != nil {
		log.Error().Err(err).Str("company_id", companyID).Msg("Failed to create company sale building")
		return nil, err
	}

	return &models.CompanySaleBuilding{
		ID:                 id,
		CompanyID:          companyID,
		SaleBuildingID:     saleBuildingID,
		Level:              1,
		ConstructionEndsAt: &endsAt,
		CreatedAt:          now,
	}, nil
}

// GetCompanySaleBuildingByID retrieves a company sale building by ID
func (r *CompanySaleBuildingRepository) GetCompanySaleBuildingByID(ctx context.Context, id string) (*models.CompanySaleBuilding, error) {
	logger := log.With().Str("building_id", id).Logger()

	dbo, err := r.queries.GetCompanySaleBuildingByID(ctx, id)
	if err == sql.ErrNoRows {
		logger.Debug().Msg("Company sale building not found")
		return nil, nil
	}
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get company sale building")
		return nil, err
	}

	return models.FromDBCompanySaleBuilding(&dbo), nil
}

// GetCompanySaleBuildingsByCompanyID retrieves all sale buildings for a company
func (r *CompanySaleBuildingRepository) GetCompanySaleBuildingsByCompanyID(ctx context.Context, companyID string) ([]*models.CompanySaleBuilding, error) {
	logger := log.With().Str("company_id", companyID).Logger()

	dbos, err := r.queries.GetCompanySaleBuildingsByCompanyID(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get company sale buildings")
		return nil, err
	}

	buildings := make([]*models.CompanySaleBuilding, 0, len(dbos))
	for _, dbo := range dbos {
		d := dbo
		buildings = append(buildings, models.FromDBCompanySaleBuilding(&d))
	}

	logger.Debug().Int("count", len(buildings)).Msg("Company sale buildings retrieved")
	return buildings, nil
}

// UpdateBuildingLevel updates the level and construction_ends_at for a sale building
func (r *CompanySaleBuildingRepository) UpdateBuildingLevel(ctx context.Context, id string, level int64, constructionEndsAt time.Time) error {
	logger := log.With().Str("building_id", id).Int64("level", level).Logger()

	endsAt := constructionEndsAt
	err := r.queries.UpdateCompanySaleBuildingLevel(ctx, dbqueries.UpdateCompanySaleBuildingLevelParams{
		ID:                 id,
		Level:              level,
		ConstructionEndsAt: &endsAt,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update company sale building level")
		return err
	}

	logger.Debug().Msg("Company sale building level updated")
	return nil
}
