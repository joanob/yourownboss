package repository

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
	"github.com/joanob/yourownboss/internal/sale/models"
)

// SaleBuildingRepository handles all sale building-related database operations
type SaleBuildingRepository struct {
	queries *dbqueries.Queries
}

// NewSaleBuildingRepository creates a new sale building repository
func NewSaleBuildingRepository(queries *dbqueries.Queries) *SaleBuildingRepository {
	return &SaleBuildingRepository{
		queries: queries,
	}
}

// CreateSaleBuilding creates a new sale building in the database
func (r *SaleBuildingRepository) CreateSaleBuilding(ctx context.Context, building *models.SaleBuilding) (*models.SaleBuilding, error) {
	logger := log.With().Str("building_id", building.ID).Str("master_id", building.MasterID).Logger()

	dbo := building.ToSaleBuildingDBO()
	err := r.queries.CreateSaleBuilding(ctx, dbqueries.CreateSaleBuildingParams{
		ID:                dbo.ID,
		MasterID:          dbo.MasterID,
		Name:              dbo.Name,
		ConstructionCost:  dbo.ConstructionCost,
		ConstructionTimeS: dbo.ConstructionTimeSec,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create sale building")
		return nil, err
	}

	logger.Debug().Msg("Sale building created successfully")
	return building, nil
}

// GetSaleBuildingByID retrieves a sale building by its ID
func (r *SaleBuildingRepository) GetSaleBuildingByID(ctx context.Context, buildingID string) (*models.SaleBuilding, error) {
	logger := log.With().Str("building_id", buildingID).Logger()

	dbo, err := r.queries.GetSaleBuildingByID(ctx, buildingID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get sale building")
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Debug().Msg("Sale building not found")
		return nil, nil
	}

	dboDbo := models.FromDBQueriesSaleBuilding(&dbo)
	building := models.NewSaleBuilding(dboDbo)

	// Load sale resources for this building
	resources, err := r.queries.GetSaleResourcesByBuildingID(ctx, buildingID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get sale building resources")
		return nil, err
	}

	for _, res := range resources {
		building.AddResource(res.ResourceID, res.PricePerUnit, res.UnitsSoldPerSecond)
	}

	logger.Debug().Msg("Sale building retrieved successfully")
	return building, nil
}

// GetSaleBuildingByMasterID retrieves a sale building by its master ID
func (r *SaleBuildingRepository) GetSaleBuildingByMasterID(ctx context.Context, masterID string) (*models.SaleBuilding, error) {
	logger := log.With().Str("master_id", masterID).Logger()

	dbo, err := r.queries.GetSaleBuildingByMasterID(ctx, masterID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get sale building by master_id")
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Debug().Msg("Sale building not found")
		return nil, nil
	}

	dboDbo := models.FromDBQueriesSaleBuilding(&dbo)
	building := models.NewSaleBuilding(dboDbo)

	// Load sale resources
	resources, err := r.queries.GetSaleResourcesByBuildingID(ctx, dbo.ID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get sale building resources")
		return nil, err
	}

	for _, res := range resources {
		building.AddResource(res.ResourceID, res.PricePerUnit, res.UnitsSoldPerSecond)
	}

	logger.Debug().Msg("Sale building retrieved successfully")
	return building, nil
}

// GetAllSaleBuildings retrieves all sale buildings
func (r *SaleBuildingRepository) GetAllSaleBuildings(ctx context.Context) ([]*models.SaleBuilding, error) {
	logger := log.Logger

	dbos, err := r.queries.GetAllSaleBuildings(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all sale buildings")
		return nil, err
	}

	buildings := make([]*models.SaleBuilding, 0, len(dbos))
	for _, dbo := range dbos {
		dboDbo := models.FromDBQueriesSaleBuilding(&dbo)
		building := models.NewSaleBuilding(dboDbo)

		// Load resources for each building
		resources, err := r.queries.GetSaleResourcesByBuildingID(ctx, dbo.ID)
		if err != nil && err != sql.ErrNoRows {
			logger.Error().Err(err).Str("building_id", dbo.ID).Msg("Failed to get sale building resources")
			return nil, err
		}

		for _, res := range resources {
			building.AddResource(res.ResourceID, res.PricePerUnit, res.UnitsSoldPerSecond)
		}

		buildings = append(buildings, building)
	}

	logger.Debug().Int("count", len(buildings)).Msg("All sale buildings retrieved successfully")
	return buildings, nil
}

// UpsertSaleBuilding creates or updates a sale building (used for importing gamedata)
func (r *SaleBuildingRepository) UpsertSaleBuilding(ctx context.Context, building *models.SaleBuilding) (*models.SaleBuilding, error) {
	logger := log.With().Str("building_id", building.ID).Str("master_id", building.MasterID).Logger()

	dbo := building.ToSaleBuildingDBO()
	err := r.queries.UpsertSaleBuilding(ctx, dbqueries.UpsertSaleBuildingParams{
		ID:                dbo.ID,
		MasterID:          dbo.MasterID,
		Name:              dbo.Name,
		ConstructionCost:  dbo.ConstructionCost,
		ConstructionTimeS: dbo.ConstructionTimeSec,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to upsert sale building")
		return nil, err
	}

	logger.Debug().Msg("Sale building upserted successfully")
	return building, nil
}
