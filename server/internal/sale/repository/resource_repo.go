package repository

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// SaleResourceRepository handles sale building resources-related database operations
type SaleResourceRepository struct {
	queries *dbqueries.Queries
}

// NewSaleResourceRepository creates a new sale resource repository
func NewSaleResourceRepository(queries *dbqueries.Queries) *SaleResourceRepository {
	return &SaleResourceRepository{
		queries: queries,
	}
}

// CreateSaleResource creates a resource relationship for a sale building
func (r *SaleResourceRepository) CreateSaleResource(ctx context.Context, saleBuildingID, resourceID string, pricePerUnit, unitsSoldPerSecond int64) error {
	logger := log.With().
		Str("sale_building_id", saleBuildingID).
		Str("resource_id", resourceID).
		Int64("price_per_unit", pricePerUnit).
		Int64("units_sold_per_second", unitsSoldPerSecond).
		Logger()

	err := r.queries.CreateSaleResource(ctx, dbqueries.CreateSaleResourceParams{
		SaleBuildingID:     saleBuildingID,
		ResourceID:         resourceID,
		PricePerUnit:       pricePerUnit,
		UnitsSoldPerSecond: unitsSoldPerSecond,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create sale resource")
		return err
	}

	logger.Debug().Msg("Sale resource created successfully")
	return nil
}

// GetSaleResourcesByBuildingID retrieves all resources for a sale building
func (r *SaleResourceRepository) GetSaleResourcesByBuildingID(ctx context.Context, saleBuildingID string) ([]dbqueries.SaleResource, error) {
	logger := log.With().Str("sale_building_id", saleBuildingID).Logger()

	resources, err := r.queries.GetSaleResourcesByBuildingID(ctx, saleBuildingID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get sale building resources")
		return nil, err
	}

	logger.Debug().Int("count", len(resources)).Msg("Sale building resources retrieved successfully")
	return resources, nil
}

// DeleteSaleResources deletes all resources for a sale building
func (r *SaleResourceRepository) DeleteSaleResources(ctx context.Context, saleBuildingID string) error {
	logger := log.With().Str("sale_building_id", saleBuildingID).Logger()

	err := r.queries.DeleteSaleResources(ctx, saleBuildingID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete sale building resources")
		return err
	}

	logger.Debug().Msg("Sale building resources deleted successfully")
	return nil
}
