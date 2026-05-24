package repository

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
	"github.com/joanob/yourownboss/internal/production/models"
)

// ProductionBuildingRepository handles all production building-related database operations
type ProductionBuildingRepository struct {
	queries *dbqueries.Queries
}

// NewProductionBuildingRepository creates a new production building repository
func NewProductionBuildingRepository(queries *dbqueries.Queries) *ProductionBuildingRepository {
	return &ProductionBuildingRepository{
		queries: queries,
	}
}

// CreateProductionBuilding creates a new production building in the database
func (r *ProductionBuildingRepository) CreateProductionBuilding(ctx context.Context, building *models.ProductionBuilding) (*models.ProductionBuilding, error) {
	logger := log.With().Str("building_id", building.ID).Str("master_id", building.MasterID).Logger()

	dbo := building.ToProductionBuildingDBO()
	err := r.queries.CreateProductionBuilding(ctx, dbqueries.CreateProductionBuildingParams{
		ID:                dbo.ID,
		MasterID:          dbo.MasterID,
		Name:              dbo.Name,
		ConstructionCost:  dbo.ConstructionCost,
		ConstructionTimeS: dbo.ConstructionTimeSec,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create production building")
		return nil, err
	}

	logger.Debug().Msg("Production building created successfully")
	return building, nil
}

// GetProductionBuildingByID retrieves a production building by its ID
func (r *ProductionBuildingRepository) GetProductionBuildingByID(ctx context.Context, buildingID string) (*models.ProductionBuilding, error) {
	logger := log.With().Str("building_id", buildingID).Logger()

	dbo, err := r.queries.GetProductionBuildingByID(ctx, buildingID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get production building")
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Debug().Msg("Production building not found")
		return nil, nil
	}

	dboDbo := models.FromDBQueriesProductionBuilding(&dbo)
	logger.Debug().Msg("Production building retrieved successfully")
	return models.NewProductionBuilding(dboDbo), nil
}

// GetProductionBuildingByMasterID retrieves a production building by its master ID
func (r *ProductionBuildingRepository) GetProductionBuildingByMasterID(ctx context.Context, masterID string) (*models.ProductionBuilding, error) {
	logger := log.With().Str("master_id", masterID).Logger()

	dbo, err := r.queries.GetProductionBuildingByMasterID(ctx, masterID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get production building by master_id")
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Debug().Msg("Production building not found")
		return nil, nil
	}

	dboDbo := models.FromDBQueriesProductionBuilding(&dbo)
	logger.Debug().Msg("Production building retrieved successfully")
	return models.NewProductionBuilding(dboDbo), nil
}

// GetAllProductionBuildings retrieves all production buildings
func (r *ProductionBuildingRepository) GetAllProductionBuildings(ctx context.Context) ([]*models.ProductionBuilding, error) {
	logger := log.Logger

	dbos, err := r.queries.GetAllProductionBuildings(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all production buildings")
		return nil, err
	}

	buildings := make([]*models.ProductionBuilding, 0, len(dbos))
	for _, dbo := range dbos {
		dboDbo := models.FromDBQueriesProductionBuilding(&dbo)
		buildings = append(buildings, models.NewProductionBuilding(dboDbo))
	}

	logger.Debug().Int("count", len(buildings)).Msg("All production buildings retrieved successfully")
	return buildings, nil
}

// UpsertProductionBuilding creates or updates a production building (used for importing gamedata)
func (r *ProductionBuildingRepository) UpsertProductionBuilding(ctx context.Context, building *models.ProductionBuilding) (*models.ProductionBuilding, error) {
	logger := log.With().Str("building_id", building.ID).Str("master_id", building.MasterID).Logger()

	dbo := building.ToProductionBuildingDBO()
	err := r.queries.UpsertProductionBuilding(ctx, dbqueries.UpsertProductionBuildingParams{
		ID:                dbo.ID,
		MasterID:          dbo.MasterID,
		Name:              dbo.Name,
		ConstructionCost:  dbo.ConstructionCost,
		ConstructionTimeS: dbo.ConstructionTimeSec,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to upsert production building")
		return nil, err
	}

	logger.Debug().Msg("Production building upserted successfully")
	return building, nil
}
