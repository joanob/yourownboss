package repository

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
	"github.com/joanob/yourownboss/internal/production/models"
)

// ProductionProcessRepository handles all production process-related database operations
type ProductionProcessRepository struct {
	queries *dbqueries.Queries
}

// NewProductionProcessRepository creates a new production process repository
func NewProductionProcessRepository(queries *dbqueries.Queries) *ProductionProcessRepository {
	return &ProductionProcessRepository{
		queries: queries,
	}
}

// CreateProductionProcess creates a new production process in the database
func (r *ProductionProcessRepository) CreateProductionProcess(ctx context.Context, process *models.ProductionProcess) (*models.ProductionProcess, error) {
	logger := log.With().Str("process_id", process.ID).Str("master_id", process.MasterID).Logger()

	dbo := process.ToProductionProcessDBO()
	err := r.queries.CreateProductionProcess(ctx, dbqueries.CreateProductionProcessParams{
		ID:                   dbo.ID,
		MasterID:             dbo.MasterID,
		ProductionBuildingID: dbo.ProductionBuildingID,
		Name:                 dbo.Name,
		CycleTimeS:           dbo.CycleTimeSec,
		WindowStartHour:      dbo.WindowStartHour,
		WindowEndHour:        dbo.WindowEndHour,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create production process")
		return nil, err
	}

	logger.Debug().Msg("Production process created successfully")
	return process, nil
}

// GetProcessByID retrieves a production process by its ID with all its resources
func (r *ProductionProcessRepository) GetProcessByID(ctx context.Context, processID string) (*models.ProductionProcess, error) {
	logger := log.With().Str("process_id", processID).Logger()

	dbo, err := r.queries.GetProductionProcessByID(ctx, processID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get production process")
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Debug().Msg("Production process not found")
		return nil, nil
	}

	dboDbo := models.FromDBQueriesProductionProcess(&dbo)
	process := models.NewProductionProcess(dboDbo)

	// Load process resources (input and output)
	resources, err := r.queries.GetProductionProcessResources(ctx, processID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get production process resources")
		return nil, err
	}

	for _, res := range resources {
		if res.IsOutput {
			process.AddOutputResource(res.ResourceID, res.Quantity)
		} else {
			process.AddInputResource(res.ResourceID, res.Quantity)
		}
	}

	logger.Debug().Msg("Production process retrieved successfully")
	return process, nil
}

// GetProcessByMasterID retrieves a production process by its master ID
func (r *ProductionProcessRepository) GetProcessByMasterID(ctx context.Context, masterID string) (*models.ProductionProcess, error) {
	logger := log.With().Str("master_id", masterID).Logger()

	dbo, err := r.queries.GetProductionProcessByMasterID(ctx, masterID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get production process by master_id")
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Debug().Msg("Production process not found")
		return nil, nil
	}

	dboDbo := models.FromDBQueriesProductionProcess(&dbo)
	process := models.NewProductionProcess(dboDbo)

	// Load process resources
	resources, err := r.queries.GetProductionProcessResources(ctx, dbo.ID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get production process resources")
		return nil, err
	}

	for _, res := range resources {
		if res.IsOutput {
			process.AddOutputResource(res.ResourceID, res.Quantity)
		} else {
			process.AddInputResource(res.ResourceID, res.Quantity)
		}
	}

	logger.Debug().Msg("Production process retrieved successfully")
	return process, nil
}

// GetProcessesByBuildingID retrieves all processes for a production building
func (r *ProductionProcessRepository) GetProcessesByBuildingID(ctx context.Context, buildingID string) ([]*models.ProductionProcess, error) {
	logger := log.With().Str("building_id", buildingID).Logger()

	dbos, err := r.queries.GetProductionProcessesByBuildingID(ctx, buildingID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get production processes by building_id")
		return nil, err
	}

	processes := make([]*models.ProductionProcess, 0, len(dbos))
	for _, dbo := range dbos {
		dboDbo := models.FromDBQueriesProductionProcess(&dbo)
		process := models.NewProductionProcess(dboDbo)

		// Load resources for each process
		resources, err := r.queries.GetProductionProcessResources(ctx, dbo.ID)
		if err != nil && err != sql.ErrNoRows {
			logger.Error().Err(err).Str("process_id", dbo.ID).Msg("Failed to get process resources")
			return nil, err
		}

		for _, res := range resources {
			if res.IsOutput {
				process.AddOutputResource(res.ResourceID, res.Quantity)
			} else {
				process.AddInputResource(res.ResourceID, res.Quantity)
			}
		}

		processes = append(processes, process)
	}

	logger.Debug().Int("count", len(processes)).Msg("Production processes retrieved successfully")
	return processes, nil
}

// UpsertProductionProcess creates or updates a production process (used for importing gamedata)
func (r *ProductionProcessRepository) UpsertProductionProcess(ctx context.Context, process *models.ProductionProcess) (*models.ProductionProcess, error) {
	logger := log.With().Str("process_id", process.ID).Str("master_id", process.MasterID).Logger()

	dbo := process.ToProductionProcessDBO()
	err := r.queries.UpsertProductionProcess(ctx, dbqueries.UpsertProductionProcessParams{
		ID:                   dbo.ID,
		MasterID:             dbo.MasterID,
		ProductionBuildingID: dbo.ProductionBuildingID,
		Name:                 dbo.Name,
		CycleTimeS:           dbo.CycleTimeSec,
		WindowStartHour:      dbo.WindowStartHour,
		WindowEndHour:        dbo.WindowEndHour,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to upsert production process")
		return nil, err
	}

	logger.Debug().Msg("Production process upserted successfully")
	return process, nil
}

// CreateProcessResource creates a resource relationship for a process
func (r *ProductionProcessRepository) CreateProcessResource(ctx context.Context, processID, resourceID string, isOutput int64, quantity int64) error {
	logger := log.With().
		Str("process_id", processID).
		Str("resource_id", resourceID).
		Int64("is_output", isOutput).
		Logger()

	err := r.queries.CreateProductionProcessResource(ctx, dbqueries.CreateProductionProcessResourceParams{
		ProcessID:  processID,
		ResourceID: resourceID,
		IsOutput:   isOutput == 1,
		Quantity:   quantity,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create process resource")
		return err
	}

	logger.Debug().Msg("Process resource created successfully")
	return nil
}

// DeleteProcessResources deletes all resources for a process
func (r *ProductionProcessRepository) DeleteProcessResources(ctx context.Context, processID string) error {
	logger := log.With().Str("process_id", processID).Logger()

	err := r.queries.DeleteProductionProcessResources(ctx, processID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete process resources")
		return err
	}

	logger.Debug().Msg("Process resources deleted successfully")
	return nil
}
