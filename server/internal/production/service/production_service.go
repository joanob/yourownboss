package service

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"

	auditRepo "github.com/joanob/yourownboss/internal/audit/repository"
	companyModels "github.com/joanob/yourownboss/internal/company/models"
	companyRepo "github.com/joanob/yourownboss/internal/company/repository"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	productionModels "github.com/joanob/yourownboss/internal/production/models"
	productionRepo "github.com/joanob/yourownboss/internal/production/repository"
)

// Production service domain errors
var (
	ErrBuildingNotFound      = errors.New("building_not_found")
	ErrBuildingNotIdle       = errors.New("building_not_idle")
	ErrBuildingNotProducing  = errors.New("building_not_producing")
	ErrProductionNotComplete = errors.New("production_not_complete")
	ErrProcessNotFound       = errors.New("process_not_found")
	ErrProcessNotAvailable   = errors.New("process_not_available_now")
	ErrInvalidCycles         = errors.New("invalid_cycles")
)

// ProductionService manages production building operations
type ProductionService struct {
	companyRepo       companyRepo.CompanyRepositoryInterface
	inventoryRepo     companyRepo.InventoryRepositoryInterface
	buildingRepo      productionRepo.CompanyBuildingRepositoryInterface
	productionRunRepo productionRepo.ProductionRunRepositoryInterface
	gamedataCache     *cache.GamedataCache
	auditRepo         auditRepo.AuditRepositoryInterface
}

// NewProductionService creates a new ProductionService
func NewProductionService(
	companyRepo companyRepo.CompanyRepositoryInterface,
	inventoryRepo companyRepo.InventoryRepositoryInterface,
	buildingRepo productionRepo.CompanyBuildingRepositoryInterface,
	productionRunRepo productionRepo.ProductionRunRepositoryInterface,
	gamedataCache *cache.GamedataCache,
	auditRepo auditRepo.AuditRepositoryInterface,
) *ProductionService {
	return &ProductionService{
		companyRepo:       companyRepo,
		inventoryRepo:     inventoryRepo,
		buildingRepo:      buildingRepo,
		productionRunRepo: productionRunRepo,
		gamedataCache:     gamedataCache,
		auditRepo:         auditRepo,
	}
}

// GetBuildings returns all company production buildings with their active runs
func (s *ProductionService) GetBuildings(ctx context.Context, companyID string) ([]*productionModels.CompanyProductionBuildingDTO, error) {
	buildings, err := s.buildingRepo.GetCompanyProductionBuildingsByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*productionModels.CompanyProductionBuildingDTO, 0, len(buildings))
	for _, b := range buildings {
		activeRun, err := s.productionRunRepo.GetActiveRunByBuildingID(ctx, b.ID)
		if err != nil {
			return nil, err
		}
		b.ActiveRun = activeRun
		dtos = append(dtos, b.ToDTO())
	}

	return dtos, nil
}

// BuildProductionBuilding creates a new company production building, deducting construction cost
func (s *ProductionService) BuildProductionBuilding(ctx context.Context, companyID, productionBuildingMasterID string) (*productionModels.CompanyProductionBuildingDTO, error) {
	logger := log.With().Str("company_id", companyID).Str("master_id", productionBuildingMasterID).Logger()

	// Resolve master building from cache
	masterBuilding, found := s.gamedataCache.GetProductionBuilding(productionBuildingMasterID)
	if !found {
		logger.Warn().Msg("Production building master not found in cache")
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
	building, err := s.buildingRepo.CreateCompanyProductionBuilding(ctx, companyID, masterBuilding.ID, constructionEndsAt)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create company production building")
		return nil, err
	}

	logger.Info().Str("building_id", building.ID).Time("ends_at", constructionEndsAt).Msg("Production building construction started")
	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "BUILD_PRODUCTION_BUILDING", "building", building.ID, map[string]interface{}{
			"master_id": productionBuildingMasterID,
			"cost":      masterBuilding.ConstructionCost,
		})
	}
	return building.ToDTO(), nil
}

// UpgradeBuilding upgrades an existing idle building by the given number of levels
func (s *ProductionService) UpgradeBuilding(ctx context.Context, companyID, buildingID string, levels int64) (*productionModels.CompanyProductionBuildingDTO, error) {
	logger := log.With().Str("company_id", companyID).Str("building_id", buildingID).Int64("levels", levels).Logger()

	if levels <= 0 {
		return nil, ErrInvalidCycles
	}

	// Get building
	building, err := s.buildingRepo.GetCompanyProductionBuildingByID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	if building == nil || building.CompanyID != companyID {
		return nil, ErrBuildingNotFound
	}

	// Load active run
	activeRun, err := s.productionRunRepo.GetActiveRunByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	building.ActiveRun = activeRun

	// Building must be idle to upgrade
	if !building.IsIdle() {
		return nil, ErrBuildingNotIdle
	}

	// Resolve master building data from cache by DB ID
	masterBuilding, found := s.gamedataCache.GetProductionBuildingByDBID(building.ProductionBuildingID)
	if !found {
		logger.Error().Str("production_building_id", building.ProductionBuildingID).Msg("Master building not found in cache")
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

	logger.Info().Int64("new_level", newLevel).Time("ends_at", constructionEndsAt).Msg("Building upgrade started")
	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "UPGRADE_PRODUCTION_BUILDING", "building", buildingID, map[string]interface{}{
			"levels":    levels,
			"new_level": newLevel,
			"cost":      upgradeCost,
		})
	}
	return building.ToDTO(), nil
}

// StartProduction begins a production run for an idle building
func (s *ProductionService) StartProduction(ctx context.Context, companyID, buildingID, processMasterID string, cycles int64) (*productionModels.CompanyProductionBuildingDTO, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("building_id", buildingID).
		Str("process_master_id", processMasterID).
		Int64("cycles", cycles).
		Logger()

	if cycles <= 0 {
		return nil, ErrInvalidCycles
	}

	// Get building
	building, err := s.buildingRepo.GetCompanyProductionBuildingByID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	if building == nil || building.CompanyID != companyID {
		return nil, ErrBuildingNotFound
	}

	// Load active run
	activeRun, err := s.productionRunRepo.GetActiveRunByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	building.ActiveRun = activeRun

	if !building.IsIdle() {
		return nil, ErrBuildingNotIdle
	}

	// Find the process in the cache: it must belong to the master building
	process, err := s.findProcess(building.ProductionBuildingID, processMasterID)
	if err != nil {
		return nil, err
	}

	// Validate time window (if defined for this process)
	if process.WindowStartHour != nil && process.WindowEndHour != nil {
		if err := validateProcessTimeWindow(process, time.Now().UTC(), cycles); err != nil {
			return nil, err
		}
	}

	// Validate and consume input resources (quantity × cycles × level)
	level := building.Level
	for _, res := range process.Resources {
		if !res.IsOutput {
			required := res.Quantity * cycles * level
			item, err := s.inventoryRepo.GetInventoryItem(ctx, companyID, res.ResourceID)
			if err != nil {
				return nil, err
			}
			if item == nil || item.Quantity < required {
				return nil, companyModels.ErrInsufficientInventory
			}
		}
	}
	for _, res := range process.Resources {
		if !res.IsOutput {
			required := res.Quantity * cycles * level
			if _, err := s.inventoryRepo.RemoveFromInventory(ctx, companyID, res.ResourceID, required); err != nil {
				return nil, err
			}
		}
	}

	// Create production run
	now := time.Now().UTC()
	endsAt := now.Add(time.Duration(process.CycleTimeS*cycles) * time.Second)
	run, err := s.productionRunRepo.CreateProductionRun(ctx, buildingID, process.ID, cycles, now, endsAt)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create production run")
		return nil, err
	}

	building.ActiveRun = run
	logger.Info().Str("run_id", run.ID).Time("ends_at", endsAt).Msg("Production started")
	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "START_PRODUCTION", "building", buildingID, map[string]interface{}{
			"run_id":     run.ID,
			"process_id": processMasterID,
			"cycles":     cycles,
		})
	}
	return building.ToDTO(), nil
}

// CollectProduction collects outputs from a completed production run
func (s *ProductionService) CollectProduction(ctx context.Context, companyID, buildingID string) (*productionModels.CompanyProductionBuildingDTO, error) {
	logger := log.With().Str("company_id", companyID).Str("building_id", buildingID).Logger()

	// Get building
	building, err := s.buildingRepo.GetCompanyProductionBuildingByID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	if building == nil || building.CompanyID != companyID {
		return nil, ErrBuildingNotFound
	}

	// Get active run
	activeRun, err := s.productionRunRepo.GetActiveRunByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	building.ActiveRun = activeRun

	if !building.IsProducing() {
		return nil, ErrBuildingNotProducing
	}

	if !activeRun.IsCompleted() {
		return nil, ErrProductionNotComplete
	}

	// Get the process from cache to know its output resources
	process, found := s.gamedataCache.GetProductionProcessByID(activeRun.ProcessID)
	if !found {
		logger.Error().Str("process_id", activeRun.ProcessID).Msg("Process not found in cache during collect")
		return nil, ErrProcessNotFound
	}

	// Add output resources to inventory (quantity × cycles × level)
	level := building.Level
	for _, res := range process.Resources {
		if res.IsOutput {
			qty := res.Quantity * activeRun.ProductionCycles * level
			if _, err := s.inventoryRepo.AddToInventory(ctx, companyID, res.ResourceID, qty); err != nil {
				return nil, err
			}
		}
	}

	// Mark run as collected
	if err := s.productionRunRepo.MarkRunCollected(ctx, activeRun.ID); err != nil {
		return nil, err
	}

	building.ActiveRun = nil
	logger.Info().Str("run_id", activeRun.ID).Msg("Production collected")
	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "COLLECT_PRODUCTION", "building", buildingID, map[string]interface{}{
			"run_id": activeRun.ID,
		})
	}
	return building.ToDTO(), nil
}

// findProcess locates a process by its masterID within the master building identified by productionBuildingDBID
func (s *ProductionService) findProcess(productionBuildingDBID, processMasterID string) (*cache.ProductionProcess, error) {
	masterBuilding, found := s.gamedataCache.GetProductionBuildingByDBID(productionBuildingDBID)
	if !found {
		return nil, ErrBuildingNotFound
	}

	for _, p := range masterBuilding.Processes {
		if p.MasterID == processMasterID {
			pp := p
			return &pp, nil
		}
	}
	return nil, ErrProcessNotFound
}

// validateProcessTimeWindow checks the current time is inside the process window and that the
// production run (cycles × cycleTimeS seconds) fits entirely within the window on the same UTC day.
func validateProcessTimeWindow(process *cache.ProductionProcess, now time.Time, cycles int64) error {
	windowStart := *process.WindowStartHour // seconds from UTC midnight
	windowEnd := *process.WindowEndHour

	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	secondsNow := int64(now.Sub(midnight).Seconds())

	if secondsNow < windowStart || secondsNow >= windowEnd {
		return ErrProcessNotAvailable
	}

	// The entire run must also end within the window on the same day
	runDurationSec := process.CycleTimeS * cycles
	if secondsNow+runDurationSec > windowEnd {
		return ErrProcessNotAvailable
	}

	return nil
}
