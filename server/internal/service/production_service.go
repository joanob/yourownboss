package service

import (
	"context"

	"yourownboss/internal/db"
	"yourownboss/internal/repository"
)

// ProductionService handles production building queries.
type ProductionService interface {
	GetProductionBuildings(ctx context.Context) ([]ProductionBuildingDetails, error)
}

// ProductionBuildingDetails represents a building with its processes.
type ProductionBuildingDetails struct {
	ID        int64
	Name      string
	Cost      int64
	Processes []ProductionProcessDetails
}

// ProductionProcessDetails represents a process with its resources.
type ProductionProcessDetails struct {
	ID               int64
	Name             string
	ProcessingTimeMs int64
	WindowStartHour  *int64
	WindowEndHour    *int64
	Resources        []ProductionProcessResourceDetails
}

// ProductionProcessResourceDetails represents an input/output resource for a process.
type ProductionProcessResourceDetails struct {
	ResourceID   int64
	ResourceName string
	Direction    string
	Quantity     int64
}

type productionService struct {
	buildingRepo        repository.ProductionBuildingRepository
	processRepo         repository.ProductionProcessRepository
	processResourceRepo repository.ProductionProcessResourceRepository
	resourceRepo        repository.ResourceRepository
}

// NewProductionService creates a new production service.
func NewProductionService(
	buildingRepo repository.ProductionBuildingRepository,
	processRepo repository.ProductionProcessRepository,
	processResourceRepo repository.ProductionProcessResourceRepository,
	resourceRepo repository.ResourceRepository,
) ProductionService {
	return &productionService{
		buildingRepo:        buildingRepo,
		processRepo:         processRepo,
		processResourceRepo: processResourceRepo,
		resourceRepo:        resourceRepo,
	}
}

func (s *productionService) GetProductionBuildings(ctx context.Context) ([]ProductionBuildingDetails, error) {
	buildings, err := s.buildingRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	resources, err := s.resourceRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	resourceByID := make(map[int64]db.Resource, len(resources))
	for _, res := range resources {
		resourceByID[res.ID] = res
	}

	result := make([]ProductionBuildingDetails, 0, len(buildings))
	for _, building := range buildings {
		processes, err := s.processRepo.GetAllByBuilding(ctx, building.ID)
		if err != nil {
			return nil, err
		}

		processDetails := make([]ProductionProcessDetails, 0, len(processes))
		for _, process := range processes {
			processResources, err := s.processResourceRepo.GetAllByProcess(ctx, process.ID)
			if err != nil {
				return nil, err
			}

			resourcesDetails := make([]ProductionProcessResourceDetails, 0, len(processResources))
			for _, processResource := range processResources {
				resourceName := ""
				if res, ok := resourceByID[processResource.ResourceID]; ok {
					resourceName = res.Name
				}
				resourcesDetails = append(resourcesDetails, ProductionProcessResourceDetails{
					ResourceID:   processResource.ResourceID,
					ResourceName: resourceName,
					Direction:    processResource.Direction,
					Quantity:     processResource.Quantity,
				})
			}

			processDetails = append(processDetails, ProductionProcessDetails{
				ID:               process.ID,
				Name:             process.Name,
				ProcessingTimeMs: process.ProcessingTimeMs,
				WindowStartHour:  process.WindowStartHour,
				WindowEndHour:    process.WindowEndHour,
				Resources:        resourcesDetails,
			})
		}

		result = append(result, ProductionBuildingDetails{
			ID:        building.ID,
			Name:      building.Name,
			Cost:      building.Cost,
			Processes: processDetails,
		})
	}

	return result, nil
}
