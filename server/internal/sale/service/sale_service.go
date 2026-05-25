package service

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"

	auditRepository "github.com/joanob/yourownboss/internal/audit/repository"
	companyModels "github.com/joanob/yourownboss/internal/company/models"
	companyRepo "github.com/joanob/yourownboss/internal/company/repository"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	saleModels "github.com/joanob/yourownboss/internal/sale/models"
	saleRepo "github.com/joanob/yourownboss/internal/sale/repository"
)

// Sale service domain errors
var (
	ErrBuildingNotFound   = errors.New("building_not_found")
	ErrBuildingNotIdle    = errors.New("building_not_idle")
	ErrBuildingNotSelling = errors.New("building_not_selling")
	ErrSaleNotComplete    = errors.New("sale_not_complete")
	ErrResourceNotFound   = errors.New("resource_not_found")
	ErrInvalidUnits       = errors.New("invalid_units")
	ErrInvalidLevels      = errors.New("invalid_levels")
)

// SaleService manages sale building operations
type SaleService struct {
	companyRepo   companyRepo.CompanyRepositoryInterface
	inventoryRepo companyRepo.InventoryRepositoryInterface
	buildingRepo  saleRepo.CompanySaleBuildingRepositoryInterface
	saleRunRepo   saleRepo.SaleRunRepositoryInterface
	gamedataCache *cache.GamedataCache
	auditRepo     auditRepository.AuditRepositoryInterface
}

// NewSaleService creates a new SaleService
func NewSaleService(
	companyRepo companyRepo.CompanyRepositoryInterface,
	inventoryRepo companyRepo.InventoryRepositoryInterface,
	buildingRepo saleRepo.CompanySaleBuildingRepositoryInterface,
	saleRunRepo saleRepo.SaleRunRepositoryInterface,
	gamedataCache *cache.GamedataCache,
	auditRepo auditRepository.AuditRepositoryInterface,
) *SaleService {
	return &SaleService{
		companyRepo:   companyRepo,
		inventoryRepo: inventoryRepo,
		buildingRepo:  buildingRepo,
		saleRunRepo:   saleRunRepo,
		gamedataCache: gamedataCache,
		auditRepo:     auditRepo,
	}
}

// GetBuildings returns all company sale buildings with their active runs
func (s *SaleService) GetBuildings(ctx context.Context, companyID string) ([]*saleModels.CompanySaleBuildingDTO, error) {
	buildings, err := s.buildingRepo.GetCompanySaleBuildingsByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*saleModels.CompanySaleBuildingDTO, 0, len(buildings))
	for _, b := range buildings {
		activeRun, err := s.saleRunRepo.GetActiveRunByBuildingID(ctx, b.ID)
		if err != nil {
			return nil, err
		}
		b.ActiveRun = activeRun
		dtos = append(dtos, b.ToDTO())
	}

	return dtos, nil
}

// BuildSaleBuilding creates a new company sale building, deducting construction cost
func (s *SaleService) BuildSaleBuilding(ctx context.Context, companyID, saleBuildingMasterID string) (*saleModels.CompanySaleBuildingDTO, error) {
	logger := log.With().Str("company_id", companyID).Str("master_id", saleBuildingMasterID).Logger()

	// Resolve master building from cache
	masterBuilding, found := s.gamedataCache.GetSaleBuilding(saleBuildingMasterID)
	if !found {
		logger.Warn().Msg("Sale building master not found in cache")
		return nil, ErrBuildingNotFound
	}

	// Get company
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil || company == nil {
		return nil, companyModels.ErrCompanyNotFound
	}

	// Check funds
	if !company.CanAfford(masterBuilding.ConstructionCost) {
		return nil, companyModels.ErrInsufficientFunds
	}

	// Deduct money
	if err := company.RemoveMoney(masterBuilding.ConstructionCost); err != nil {
		return nil, err
	}
	if _, err := s.companyRepo.UpdateCompanyMoney(ctx, companyID, company.Money); err != nil {
		return nil, err
	}

	// Create the building record
	constructionEndsAt := time.Now().UTC().Add(time.Duration(masterBuilding.ConstructionTimeS) * time.Second)
	building, err := s.buildingRepo.CreateCompanySaleBuilding(ctx, companyID, masterBuilding.ID, constructionEndsAt)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create company sale building")
		return nil, err
	}

	logger.Info().Str("building_id", building.ID).Time("ends_at", constructionEndsAt).Msg("Sale building construction started")
	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "BUILD_SALE_BUILDING", "building", building.ID, map[string]interface{}{
			"master_id": saleBuildingMasterID,
			"cost":      masterBuilding.ConstructionCost,
		})
	}
	return building.ToDTO(), nil
}

// UpgradeBuilding upgrades an existing idle building by the given number of levels
func (s *SaleService) UpgradeBuilding(ctx context.Context, companyID, buildingID string, levels int64) (*saleModels.CompanySaleBuildingDTO, error) {
	logger := log.With().Str("company_id", companyID).Str("building_id", buildingID).Int64("levels", levels).Logger()

	if levels <= 0 {
		return nil, ErrInvalidLevels
	}

	// Get building
	building, err := s.buildingRepo.GetCompanySaleBuildingByID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	if building == nil || building.CompanyID != companyID {
		return nil, ErrBuildingNotFound
	}

	// Load active run
	activeRun, err := s.saleRunRepo.GetActiveRunByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	building.ActiveRun = activeRun

	// Building must be idle to upgrade
	if !building.IsIdle() {
		return nil, ErrBuildingNotIdle
	}

	// Resolve master building data from cache by DB ID
	masterBuilding, found := s.gamedataCache.GetSaleBuildingByDBID(building.SaleBuildingID)
	if !found {
		logger.Error().Str("sale_building_id", building.SaleBuildingID).Msg("Master building not found in cache")
		return nil, ErrBuildingNotFound
	}

	// Upgrade cost: base construction cost × number of levels
	upgradeCost := masterBuilding.ConstructionCost * levels

	// Get company
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil || company == nil {
		return nil, companyModels.ErrCompanyNotFound
	}
	if !company.CanAfford(upgradeCost) {
		return nil, companyModels.ErrInsufficientFunds
	}

	// Deduct money
	if err := company.RemoveMoney(upgradeCost); err != nil {
		return nil, err
	}
	if _, err := s.companyRepo.UpdateCompanyMoney(ctx, companyID, company.Money); err != nil {
		return nil, err
	}

	// Update level and construction time
	newLevel := building.Level + levels
	constructionEndsAt := time.Now().UTC().Add(time.Duration(masterBuilding.ConstructionTimeS*levels) * time.Second)
	if err := s.buildingRepo.UpdateBuildingLevel(ctx, buildingID, newLevel, constructionEndsAt); err != nil {
		return nil, err
	}

	building.Level = newLevel
	building.ConstructionEndsAt = &constructionEndsAt
	building.ActiveRun = nil

	logger.Info().Int64("new_level", newLevel).Time("ends_at", constructionEndsAt).Msg("Sale building upgrade started")
	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "UPGRADE_SALE_BUILDING", "building", buildingID, map[string]interface{}{
			"levels":    levels,
			"new_level": newLevel,
			"cost":      upgradeCost,
		})
	}
	return building.ToDTO(), nil
}

// StartSale begins a sale run for an idle building
func (s *SaleService) StartSale(ctx context.Context, companyID, buildingID, resourceID string, units int64) (*saleModels.CompanySaleBuildingDTO, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("building_id", buildingID).
		Str("resource_id", resourceID).
		Int64("units", units).
		Logger()

	if units <= 0 {
		return nil, ErrInvalidUnits
	}

	// Get building
	building, err := s.buildingRepo.GetCompanySaleBuildingByID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	if building == nil || building.CompanyID != companyID {
		return nil, ErrBuildingNotFound
	}

	// Load active run
	activeRun, err := s.saleRunRepo.GetActiveRunByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	building.ActiveRun = activeRun

	if !building.IsIdle() {
		return nil, ErrBuildingNotIdle
	}

	// Resolve master building from cache and find the resource
	masterBuilding, found := s.gamedataCache.GetSaleBuildingByDBID(building.SaleBuildingID)
	if !found {
		logger.Error().Str("sale_building_id", building.SaleBuildingID).Msg("Master building not found in cache")
		return nil, ErrBuildingNotFound
	}

	saleResource, err := findSaleResource(masterBuilding, resourceID)
	if err != nil {
		return nil, err
	}

	// Verify the company has enough inventory
	item, err := s.inventoryRepo.GetInventoryItem(ctx, companyID, resourceID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.Quantity < units {
		return nil, companyModels.ErrInsufficientInventory
	}

	// Remove units from inventory
	if _, err := s.inventoryRepo.RemoveFromInventory(ctx, companyID, resourceID, units); err != nil {
		return nil, err
	}

	// Calculate sale duration: units / (units_sold_per_second * level)
	effectiveRate := saleResource.UnitsSoldPerSecond * building.Level
	durationSeconds := units / effectiveRate
	if durationSeconds < 1 {
		durationSeconds = 1
	}

	now := time.Now().UTC()
	endsAt := now.Add(time.Duration(durationSeconds) * time.Second)

	run, err := s.saleRunRepo.CreateSaleRun(ctx, buildingID, resourceID, units, now, endsAt)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create sale run")
		return nil, err
	}

	building.ActiveRun = run
	logger.Info().Str("run_id", run.ID).Time("ends_at", endsAt).Msg("Sale started")
	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "START_SALE", "building", buildingID, map[string]interface{}{
			"run_id":      run.ID,
			"resource_id": resourceID,
			"units":       units,
		})
	}
	return building.ToDTO(), nil
}

// CollectSale collects revenue from a completed sale run
func (s *SaleService) CollectSale(ctx context.Context, companyID, buildingID string) (*saleModels.CompanySaleBuildingDTO, error) {
	logger := log.With().Str("company_id", companyID).Str("building_id", buildingID).Logger()

	// Get building
	building, err := s.buildingRepo.GetCompanySaleBuildingByID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	if building == nil || building.CompanyID != companyID {
		return nil, ErrBuildingNotFound
	}

	// Get active run
	activeRun, err := s.saleRunRepo.GetActiveRunByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	building.ActiveRun = activeRun

	if !building.IsSelling() {
		return nil, ErrBuildingNotSelling
	}

	if !activeRun.IsCompleted() {
		return nil, ErrSaleNotComplete
	}

	// Resolve master building and find current price for the resource
	masterBuilding, found := s.gamedataCache.GetSaleBuildingByDBID(building.SaleBuildingID)
	if !found {
		logger.Error().Str("sale_building_id", building.SaleBuildingID).Msg("Master building not found in cache")
		return nil, ErrBuildingNotFound
	}

	saleResource, err := findSaleResource(masterBuilding, activeRun.ResourceID)
	if err != nil {
		return nil, err
	}

	// Revenue = units * current price_per_unit from cache
	revenue := activeRun.UnitsToSell * saleResource.PricePerUnit

	// Add money to company
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil || company == nil {
		return nil, companyModels.ErrCompanyNotFound
	}
	if err := company.AddMoney(revenue); err != nil {
		return nil, err
	}
	if _, err := s.companyRepo.UpdateCompanyMoney(ctx, companyID, company.Money); err != nil {
		return nil, err
	}

	// Mark run as collected
	now := time.Now().UTC()
	if err := s.saleRunRepo.MarkRunCollected(ctx, activeRun.ID, now); err != nil {
		return nil, err
	}

	building.ActiveRun = nil
	logger.Info().Int64("revenue", revenue).Msg("Sale collected")
	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "COLLECT_SALE", "building", buildingID, map[string]interface{}{
			"run_id":  activeRun.ID,
			"revenue": revenue,
		})
	}
	return building.ToDTO(), nil
}

// findSaleResource finds a resource by ID in the given sale building's resource list
func findSaleResource(masterBuilding cache.SaleBuilding, resourceID string) (cache.SaleResource, error) {
	for _, r := range masterBuilding.Resources {
		if r.ResourceID == resourceID {
			return r, nil
		}
	}
	return cache.SaleResource{}, ErrResourceNotFound
}
