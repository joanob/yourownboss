package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/pkg/cache"
	"github.com/joanob/yourownboss/internal/production/models"
	"github.com/joanob/yourownboss/internal/production/repository"
	resourceModels "github.com/joanob/yourownboss/internal/resources/models"
	resourceRepo "github.com/joanob/yourownboss/internal/resources/repository"
	saleModels "github.com/joanob/yourownboss/internal/sale/models"
	saleRepo "github.com/joanob/yourownboss/internal/sale/repository"
)

// GamedataService handles gamedata operations
type GamedataService struct {
	resourceRepo           *resourceRepo.ResourceRepository
	productionBuildingRepo *repository.ProductionBuildingRepository
	productionProcessRepo  *repository.ProductionProcessRepository
	saleBuildingRepo       *saleRepo.SaleBuildingRepository
	saleResourceRepo       *saleRepo.SaleResourceRepository
	gamedataCache          *cache.GamedataCache
}

// NewGamedataService creates a new gamedata service
func NewGamedataService(
	resourceRepo *resourceRepo.ResourceRepository,
	productionBuildingRepo *repository.ProductionBuildingRepository,
	productionProcessRepo *repository.ProductionProcessRepository,
	saleBuildingRepo *saleRepo.SaleBuildingRepository,
	saleResourceRepo *saleRepo.SaleResourceRepository,
	gamedataCache *cache.GamedataCache,
) *GamedataService {
	return &GamedataService{
		resourceRepo:           resourceRepo,
		productionBuildingRepo: productionBuildingRepo,
		productionProcessRepo:  productionProcessRepo,
		saleBuildingRepo:       saleBuildingRepo,
		saleResourceRepo:       saleResourceRepo,
		gamedataCache:          gamedataCache,
	}
}

// GetGameData returns all gamedata from cache
func (s *GamedataService) GetGameData(ctx context.Context) (*GamedataResponse, error) {
	logger := log.Logger

	// Get all cached resources
	cachedResources := s.gamedataCache.GetAllResources()
	resources := make([]*resourceModels.Resource, 0, len(cachedResources))
	for _, cr := range cachedResources {
		resources = append(resources, &resourceModels.Resource{
			ID:            cr.ID,
			MasterID:      cr.MasterID,
			Name:          cr.Name,
			MarketPrice:   cr.MarketPrice,
			MarketSaleQty: cr.MarketSaleQty,
		})
	}

	logger.Debug().
		Int("resources", len(resources)).
		Msg("Gamedata retrieved from cache")

	return &GamedataResponse{
		Resources:           resources,
		ProductionBuildings: []*models.ProductionBuilding{},
		SaleBuildings:       []*saleModels.SaleBuilding{},
	}, nil
}

// ImportGameData imports gamedata from a request and updates both database and cache
func (s *GamedataService) ImportGameData(ctx context.Context, data *GamedataImportRequest) error {
	logger := log.Logger.With().Int("resources_count", len(data.Resources)).
		Int("production_buildings_count", len(data.ProductionBuildings)).
		Int("sale_buildings_count", len(data.SaleBuildings)).Logger()

	logger.Info().Msg("Starting gamedata import")

	// Validate input
	if err := s.validateGamedataImport(data); err != nil {
		logger.Error().Err(err).Msg("Gamedata validation failed")
		return err
	}

	// Import resources
	for _, res := range data.Resources {
		if _, err := s.resourceRepo.UpsertResource(ctx, res); err != nil {
			logger.Error().Err(err).Str("resource_id", res.ID).Msg("Failed to import resource")
			return err
		}
	}
	logger.Debug().Int("count", len(data.Resources)).Msg("Resources imported")

	// Import production buildings and their processes
	for _, building := range data.ProductionBuildings {
		if _, err := s.productionBuildingRepo.UpsertProductionBuilding(ctx, building); err != nil {
			logger.Error().Err(err).Str("building_id", building.ID).Msg("Failed to import production building")
			return err
		}

		// Import processes for this building
		for _, process := range building.Processes {
			if _, err := s.productionProcessRepo.UpsertProductionProcess(ctx, process); err != nil {
				logger.Error().Err(err).Str("process_id", process.ID).Msg("Failed to import production process")
				return err
			}

			// Delete existing resources for this process
			if err := s.productionProcessRepo.DeleteProcessResources(ctx, process.ID); err != nil {
				logger.Error().Err(err).Str("process_id", process.ID).Msg("Failed to delete existing process resources")
				return err
			}

			// Import input resources
			for _, input := range process.InputResources {
				if err := s.productionProcessRepo.CreateProcessResource(ctx, process.ID, input.ResourceID, 0, input.Quantity); err != nil {
					logger.Error().Err(err).Str("process_id", process.ID).Str("resource_id", input.ResourceID).Msg("Failed to import process input resource")
					return err
				}
			}

			// Import output resources
			for _, output := range process.OutputResources {
				if err := s.productionProcessRepo.CreateProcessResource(ctx, process.ID, output.ResourceID, 1, output.Quantity); err != nil {
					logger.Error().Err(err).Str("process_id", process.ID).Str("resource_id", output.ResourceID).Msg("Failed to import process output resource")
					return err
				}
			}
		}
	}
	logger.Debug().Int("count", len(data.ProductionBuildings)).Msg("Production buildings imported")

	// Import sale buildings and their resources
	for _, building := range data.SaleBuildings {
		if _, err := s.saleBuildingRepo.UpsertSaleBuilding(ctx, building); err != nil {
			logger.Error().Err(err).Str("building_id", building.ID).Msg("Failed to import sale building")
			return err
		}

		// Delete existing resources for this sale building
		if err := s.saleResourceRepo.DeleteSaleResources(ctx, building.ID); err != nil {
			logger.Error().Err(err).Str("building_id", building.ID).Msg("Failed to delete existing sale building resources")
			return err
		}

		// Import resources for this building
		for _, resource := range building.Resources {
			if err := s.saleResourceRepo.CreateSaleResource(ctx, building.ID, resource.ResourceID, resource.PricePerUnit, resource.UnitsSoldPerSecond); err != nil {
				logger.Error().Err(err).Str("building_id", building.ID).Str("resource_id", resource.ResourceID).Msg("Failed to import sale building resource")
				return err
			}
		}
	}
	logger.Debug().Int("count", len(data.SaleBuildings)).Msg("Sale buildings imported")

	// Refresh cache with newly imported data
	if err := s.RefreshCache(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to refresh cache after import")
		return err
	}

	logger.Info().Msg("Gamedata import completed successfully")
	return nil
}

// RefreshCache reloads all gamedata from database into cache
func (s *GamedataService) RefreshCache(ctx context.Context) error {
	logger := log.Logger

	// Load all resources
	resources, err := s.resourceRepo.GetAllResources(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load resources for cache")
		return err
	}

	cacheResources := make([]cache.Resource, 0, len(resources))
	for _, r := range resources {
		cacheResources = append(cacheResources, cache.Resource{
			ID:            r.ID,
			MasterID:      r.MasterID,
			Name:          r.Name,
			MarketPrice:   r.MarketPrice,
			MarketSaleQty: r.MarketSaleQty,
		})
	}
	s.gamedataCache.SetResources(cacheResources)
	logger.Debug().Int("count", len(resources)).Msg("Resources loaded for cache")

	// Load all production buildings with their processes and resources
	buildings, err := s.productionBuildingRepo.GetAllProductionBuildings(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load production buildings for cache")
		return err
	}

	cacheBuildings := make([]cache.ProductionBuilding, 0, len(buildings))
	for _, b := range buildings {
		processes, err := s.productionProcessRepo.GetProcessesByBuildingID(ctx, b.ID)
		if err != nil {
			logger.Error().Err(err).Str("building_id", b.ID).Msg("Failed to load processes for building")
			return err
		}

		cacheProcesses := make([]cache.ProductionProcess, 0, len(processes))
		for _, p := range processes {
			cacheResources := make([]cache.ProductionProcessResource, 0, len(p.InputResources)+len(p.OutputResources))
			for _, r := range p.InputResources {
				cacheResources = append(cacheResources, cache.ProductionProcessResource{
					ResourceID: r.ResourceID,
					IsOutput:   false,
					Quantity:   r.Quantity,
				})
			}
			for _, r := range p.OutputResources {
				cacheResources = append(cacheResources, cache.ProductionProcessResource{
					ResourceID: r.ResourceID,
					IsOutput:   true,
					Quantity:   r.Quantity,
				})
			}
			cacheProcesses = append(cacheProcesses, cache.ProductionProcess{
				ID:                   p.ID,
				MasterID:             p.MasterID,
				ProductionBuildingID: p.ProductionBuildingID,
				Name:                 p.Name,
				CycleTimeS:           p.CycleTimeSec,
				WindowStartHour:      p.WindowStartHour,
				WindowEndHour:        p.WindowEndHour,
				Resources:            cacheResources,
			})
		}

		cacheBuildings = append(cacheBuildings, cache.ProductionBuilding{
			ID:                b.ID,
			MasterID:          b.MasterID,
			Name:              b.Name,
			ConstructionCost:  b.ConstructionCost,
			ConstructionTimeS: b.ConstructionTimeSec,
			Processes:         cacheProcesses,
		})
	}
	s.gamedataCache.SetProductionBuildings(cacheBuildings)
	logger.Debug().Int("count", len(buildings)).Msg("Production buildings loaded for cache")

	// Load all sale buildings with their resources
	saleBuildings, err := s.saleBuildingRepo.GetAllSaleBuildings(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load sale buildings for cache")
		return err
	}

	cacheSaleBuildings := make([]cache.SaleBuilding, 0, len(saleBuildings))
	for _, sb := range saleBuildings {
		cacheSaleResources := make([]cache.SaleResource, 0, len(sb.Resources))
		for _, r := range sb.Resources {
			cacheSaleResources = append(cacheSaleResources, cache.SaleResource{
				ResourceID:         r.ResourceID,
				PricePerUnit:       r.PricePerUnit,
				UnitsSoldPerSecond: r.UnitsSoldPerSecond,
			})
		}
		cacheSaleBuildings = append(cacheSaleBuildings, cache.SaleBuilding{
			ID:                sb.ID,
			MasterID:          sb.MasterID,
			Name:              sb.Name,
			ConstructionCost:  sb.ConstructionCost,
			ConstructionTimeS: sb.ConstructionTimeSec,
			Resources:         cacheSaleResources,
		})
	}
	s.gamedataCache.SetSaleBuildings(cacheSaleBuildings)
	logger.Debug().Int("count", len(saleBuildings)).Msg("Sale buildings loaded for cache")

	logger.Info().
		Int("resources", len(resources)).
		Int("production_buildings", len(buildings)).
		Int("sale_buildings", len(saleBuildings)).
		Msg("Cache refreshed successfully")

	return nil
}

// validateGamedataImport validates the gamedata import request
func (s *GamedataService) validateGamedataImport(data *GamedataImportRequest) error {
	if data == nil {
		return fmt.Errorf("gamedata import request is nil")
	}

	// Check for duplicate resource master_ids
	resourceMasterIDs := make(map[string]bool)
	for _, res := range data.Resources {
		if resourceMasterIDs[res.MasterID] {
			return fmt.Errorf("duplicate resource master_id: %s", res.MasterID)
		}
		resourceMasterIDs[res.MasterID] = true

		if res.ID == "" || res.MasterID == "" || res.Name == "" {
			return fmt.Errorf("invalid resource: missing required fields")
		}
	}

	// Check for duplicate production building master_ids
	buildingMasterIDs := make(map[string]bool)
	for _, building := range data.ProductionBuildings {
		if buildingMasterIDs[building.MasterID] {
			return fmt.Errorf("duplicate production building master_id: %s", building.MasterID)
		}
		buildingMasterIDs[building.MasterID] = true

		if building.ID == "" || building.MasterID == "" || building.Name == "" {
			return fmt.Errorf("invalid production building: missing required fields")
		}
	}

	// Check for duplicate sale building master_ids
	saleBuildingMasterIDs := make(map[string]bool)
	for _, building := range data.SaleBuildings {
		if saleBuildingMasterIDs[building.MasterID] {
			return fmt.Errorf("duplicate sale building master_id: %s", building.MasterID)
		}
		saleBuildingMasterIDs[building.MasterID] = true

		if building.ID == "" || building.MasterID == "" || building.Name == "" {
			return fmt.Errorf("invalid sale building: missing required fields")
		}
	}

	return nil
}

// GamedataResponse represents the response for gamedata endpoints
type GamedataResponse struct {
	Resources           []*resourceModels.Resource   `json:"resources"`
	ProductionBuildings []*models.ProductionBuilding `json:"production_buildings"`
	SaleBuildings       []*saleModels.SaleBuilding   `json:"sale_buildings"`
}

// GamedataImportRequest represents the request body for importing gamedata
type GamedataImportRequest struct {
	Resources           []*resourceModels.Resource   `json:"resources"`
	ProductionBuildings []*models.ProductionBuilding `json:"production_buildings"`
	SaleBuildings       []*saleModels.SaleBuilding   `json:"sale_buildings"`
}
