package repository

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
	"github.com/joanob/yourownboss/internal/resources/models"
)

// ResourceRepository handles all resource-related database operations
type ResourceRepository struct {
	queries *dbqueries.Queries
}

// NewResourceRepository creates a new resource repository
func NewResourceRepository(queries *dbqueries.Queries) *ResourceRepository {
	return &ResourceRepository{
		queries: queries,
	}
}

// CreateResource creates a new resource in the database
func (r *ResourceRepository) CreateResource(ctx context.Context, resource *models.Resource) (*models.Resource, error) {
	logger := log.With().Str("resource_id", resource.ID).Str("master_id", resource.MasterID).Logger()

	dbo := resource.ToResourceDBO()
	err := r.queries.CreateResource(ctx, dbqueries.CreateResourceParams{
		ID:            dbo.ID,
		MasterID:      dbo.MasterID,
		Name:          dbo.Name,
		MarketPrice:   dbo.MarketPrice,
		MarketSaleQty: dbo.MarketSaleQty,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create resource")
		return nil, err
	}

	logger.Debug().Msg("Resource created successfully")
	return resource, nil
}

// GetResourceByID retrieves a resource by its ID
func (r *ResourceRepository) GetResourceByID(ctx context.Context, resourceID string) (*models.Resource, error) {
	logger := log.With().Str("resource_id", resourceID).Logger()

	dbo, err := r.queries.GetResourceByID(ctx, resourceID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get resource")
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Debug().Msg("Resource not found")
		return nil, nil
	}

	dboDbo := models.FromDBQueriesResource(&dbo)
	logger.Debug().Msg("Resource retrieved successfully")
	return models.NewResource(dboDbo), nil
}

// GetResourceByMasterID retrieves a resource by its master ID
func (r *ResourceRepository) GetResourceByMasterID(ctx context.Context, masterID string) (*models.Resource, error) {
	logger := log.With().Str("master_id", masterID).Logger()

	dbo, err := r.queries.GetResourceByMasterID(ctx, masterID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get resource by master_id")
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Debug().Msg("Resource not found")
		return nil, nil
	}

	dboDbo := models.FromDBQueriesResource(&dbo)
	logger.Debug().Msg("Resource retrieved successfully")
	return models.NewResource(dboDbo), nil
}

// GetAllResources retrieves all resources
func (r *ResourceRepository) GetAllResources(ctx context.Context) ([]*models.Resource, error) {
	logger := log.Logger

	dbos, err := r.queries.GetAllResources(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all resources")
		return nil, err
	}

	resources := make([]*models.Resource, 0, len(dbos))
	for _, dbo := range dbos {
		dboDbo := models.FromDBQueriesResource(&dbo)
		resources = append(resources, models.NewResource(dboDbo))
	}

	logger.Debug().Int("count", len(resources)).Msg("All resources retrieved successfully")
	return resources, nil
}

// UpsertResource creates or updates a resource (used for importing gamedata)
func (r *ResourceRepository) UpsertResource(ctx context.Context, resource *models.Resource) (*models.Resource, error) {
	logger := log.With().Str("resource_id", resource.ID).Str("master_id", resource.MasterID).Logger()

	dbo := resource.ToResourceDBO()
	err := r.queries.UpsertResource(ctx, dbqueries.UpsertResourceParams{
		ID:            dbo.ID,
		MasterID:      dbo.MasterID,
		Name:          dbo.Name,
		MarketPrice:   dbo.MarketPrice,
		MarketSaleQty: dbo.MarketSaleQty,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to upsert resource")
		return nil, err
	}

	logger.Debug().Msg("Resource upserted successfully")
	return resource, nil
}
