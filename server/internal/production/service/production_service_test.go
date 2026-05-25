package service

import (
	"context"
	"testing"
	"time"

	companyModels "github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	productionModels "github.com/joanob/yourownboss/internal/production/models"
)

// ============================================================================
// Mock CompanyRepository
// ============================================================================

type mockCompanyRepo struct {
	companies map[string]*companyModels.Company
	updateErr error
}

func newMockCompanyRepo() *mockCompanyRepo {
	return &mockCompanyRepo{companies: make(map[string]*companyModels.Company)}
}

func (m *mockCompanyRepo) CreateCompany(_ context.Context, userID, name string, initialMoney int64) (*companyModels.Company, error) {
	c := &companyModels.Company{ID: "co-" + name, UserID: userID, Name: name, Money: initialMoney, CreatedAt: time.Now().UTC()}
	m.companies[c.ID] = c
	return c, nil
}
func (m *mockCompanyRepo) GetCompanyByID(_ context.Context, id string) (*companyModels.Company, error) {
	return m.companies[id], nil
}
func (m *mockCompanyRepo) GetCompanyByUserID(_ context.Context, userID string) (*companyModels.Company, error) {
	return nil, nil
}
func (m *mockCompanyRepo) UpdateCompanyName(_ context.Context, companyID, name string) (*companyModels.Company, error) {
	return m.companies[companyID], nil
}
func (m *mockCompanyRepo) UpdateCompanyMoney(_ context.Context, companyID string, newMoney int64) (*companyModels.Company, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	c := m.companies[companyID]
	if c == nil {
		return nil, companyModels.ErrCompanyNotFound
	}
	c.Money = newMoney
	return c, nil
}
func (m *mockCompanyRepo) CheckCompanyExists(_ context.Context, companyID string) (bool, error) {
	_, ok := m.companies[companyID]
	return ok, nil
}
func (m *mockCompanyRepo) SoftDeleteCompany(_ context.Context, companyID string) error {
	delete(m.companies, companyID)
	return nil
}

// ============================================================================
// Mock InventoryRepository
// ============================================================================

type mockInventoryRepo struct {
	items     map[string]map[string]*companyModels.CompanyInventoryItem
	removeErr error
}

func newMockInventoryRepo() *mockInventoryRepo {
	return &mockInventoryRepo{items: make(map[string]map[string]*companyModels.CompanyInventoryItem)}
}

func (m *mockInventoryRepo) companyItems(companyID string) map[string]*companyModels.CompanyInventoryItem {
	if m.items[companyID] == nil {
		m.items[companyID] = make(map[string]*companyModels.CompanyInventoryItem)
	}
	return m.items[companyID]
}

func (m *mockInventoryRepo) GetInventory(_ context.Context, companyID string) (*companyModels.CompanyInventory, error) {
	inv := companyModels.NewCompanyInventory(companyID)
	for _, item := range m.companyItems(companyID) {
		inv.AddItem(item)
	}
	return inv, nil
}
func (m *mockInventoryRepo) GetInventoryItem(_ context.Context, companyID, resourceID string) (*companyModels.CompanyInventoryItem, error) {
	return m.companyItems(companyID)[resourceID], nil
}
func (m *mockInventoryRepo) AddToInventory(_ context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
	cm := m.companyItems(companyID)
	item, ok := cm[resourceID]
	if !ok {
		item = &companyModels.CompanyInventoryItem{ID: "inv-" + resourceID, CompanyID: companyID, ResourceID: resourceID}
		cm[resourceID] = item
	}
	item.Quantity += quantity
	return item, nil
}
func (m *mockInventoryRepo) RemoveFromInventory(_ context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
	if m.removeErr != nil {
		return nil, m.removeErr
	}
	item := m.companyItems(companyID)[resourceID]
	if item == nil || item.Quantity < quantity {
		return nil, companyModels.ErrInsufficientInventory
	}
	item.Quantity -= quantity
	return item, nil
}
func (m *mockInventoryRepo) UpdateInventoryQuantity(_ context.Context, itemID string, newQty int64) (*companyModels.CompanyInventoryItem, error) {
	return nil, nil
}
func (m *mockInventoryRepo) UpsertInventoryItem(_ context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
	return m.AddToInventory(context.Background(), companyID, resourceID, quantity)
}

// ============================================================================
// Mock CompanyBuildingRepository
// ============================================================================

type mockBuildingRepo struct {
	buildings      map[string]*productionModels.CompanyProductionBuilding
	createErr      error
	updateLevelErr error
	nextBuildingID string
}

func newMockBuildingRepo() *mockBuildingRepo {
	return &mockBuildingRepo{
		buildings:      make(map[string]*productionModels.CompanyProductionBuilding),
		nextBuildingID: "building-1",
	}
}

func (m *mockBuildingRepo) CreateCompanyProductionBuilding(_ context.Context, companyID, productionBuildingID string, constructionEndsAt time.Time) (*productionModels.CompanyProductionBuilding, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	b := &productionModels.CompanyProductionBuilding{
		ID:                   m.nextBuildingID,
		CompanyID:            companyID,
		ProductionBuildingID: productionBuildingID,
		Level:                1,
		ConstructionEndsAt:   &constructionEndsAt,
		CreatedAt:            time.Now().UTC(),
	}
	m.buildings[b.ID] = b
	return b, nil
}
func (m *mockBuildingRepo) GetCompanyProductionBuildingByID(_ context.Context, buildingID string) (*productionModels.CompanyProductionBuilding, error) {
	return m.buildings[buildingID], nil
}
func (m *mockBuildingRepo) GetCompanyProductionBuildingsByCompanyID(_ context.Context, companyID string) ([]*productionModels.CompanyProductionBuilding, error) {
	var result []*productionModels.CompanyProductionBuilding
	for _, b := range m.buildings {
		if b.CompanyID == companyID {
			result = append(result, b)
		}
	}
	return result, nil
}
func (m *mockBuildingRepo) UpdateBuildingLevel(_ context.Context, buildingID string, newLevel int64, constructionEndsAt time.Time) error {
	if m.updateLevelErr != nil {
		return m.updateLevelErr
	}
	b := m.buildings[buildingID]
	if b != nil {
		b.Level = newLevel
		b.ConstructionEndsAt = &constructionEndsAt
	}
	return nil
}

// ============================================================================
// Mock ProductionRunRepository
// ============================================================================

type mockProductionRunRepo struct {
	runs                map[string]*productionModels.ProductionRun
	activeRunByBuilding map[string]*productionModels.ProductionRun
	createErr           error
	nextRunID           string
}

func newMockProductionRunRepo() *mockProductionRunRepo {
	return &mockProductionRunRepo{
		runs:                make(map[string]*productionModels.ProductionRun),
		activeRunByBuilding: make(map[string]*productionModels.ProductionRun),
		nextRunID:           "run-1",
	}
}

func (m *mockProductionRunRepo) CreateProductionRun(_ context.Context, companyBuildingID, processID string, cycles int64, startedAt, endsAt time.Time) (*productionModels.ProductionRun, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	r := &productionModels.ProductionRun{
		ID:                m.nextRunID,
		CompanyBuildingID: companyBuildingID,
		ProcessID:         processID,
		ProductionCycles:  cycles,
		StartedAt:         startedAt,
		EndsAt:            endsAt,
	}
	m.runs[r.ID] = r
	m.activeRunByBuilding[companyBuildingID] = r
	return r, nil
}
func (m *mockProductionRunRepo) GetActiveRunByBuildingID(_ context.Context, buildingID string) (*productionModels.ProductionRun, error) {
	return m.activeRunByBuilding[buildingID], nil
}
func (m *mockProductionRunRepo) GetProductionRunByID(_ context.Context, runID string) (*productionModels.ProductionRun, error) {
	return m.runs[runID], nil
}
func (m *mockProductionRunRepo) MarkRunCollected(_ context.Context, runID string) error {
	r := m.runs[runID]
	if r != nil {
		r.IsCollected = true
		// Remove from active runs
		for buildingID, run := range m.activeRunByBuilding {
			if run.ID == runID {
				delete(m.activeRunByBuilding, buildingID)
			}
		}
	}
	return nil
}

// ============================================================================
// Test helpers
// ============================================================================

const (
	testCompanyID        = "co-1"
	testMasterBuildingID = "master-bakery"
	testBuildingDBID     = "db-bakery-1"
	testProcessMasterID  = "process-bake-bread"
	testProcessDBID      = "db-process-1"
	testInputResourceID  = "db-flour-1"
	testOutputResourceID = "db-bread-1"
)

func newTestProductionCache() *cache.GamedataCache {
	gc := cache.NewGamedataCache()
	gc.SetProductionBuildings([]cache.ProductionBuilding{
		{
			ID:                testBuildingDBID,
			MasterID:          testMasterBuildingID,
			Name:              "Bakery",
			ConstructionCost:  100,
			ConstructionTimeS: 60,
			Processes: []cache.ProductionProcess{
				{
					ID:         testProcessDBID,
					MasterID:   testProcessMasterID,
					Name:       "Bake Bread",
					CycleTimeS: 30,
					Resources: []cache.ProductionProcessResource{
						{ResourceID: testInputResourceID, IsOutput: false, Quantity: 2},
						{ResourceID: testOutputResourceID, IsOutput: true, Quantity: 5},
					},
				},
			},
		},
	})
	return gc
}

func newTestCompany(id string, money int64) *companyModels.Company {
	return &companyModels.Company{
		ID:        id,
		UserID:    "user-1",
		Name:      "Test Co",
		Money:     money,
		CreatedAt: time.Now().UTC(),
	}
}

func newIdleBuilding(id, companyID, buildingDBID string, level int64) *productionModels.CompanyProductionBuilding {
	return &productionModels.CompanyProductionBuilding{
		ID:                   id,
		CompanyID:            companyID,
		ProductionBuildingID: buildingDBID,
		Level:                level,
		CreatedAt:            time.Now().UTC(),
	}
}

func newConstructingBuilding(id, companyID, buildingDBID string) *productionModels.CompanyProductionBuilding {
	endsAt := time.Now().Add(1 * time.Hour)
	return &productionModels.CompanyProductionBuilding{
		ID:                   id,
		CompanyID:            companyID,
		ProductionBuildingID: buildingDBID,
		Level:                1,
		ConstructionEndsAt:   &endsAt,
		CreatedAt:            time.Now().UTC(),
	}
}

// ============================================================================
// BuildProductionBuilding tests
// ============================================================================

func TestProductionService_BuildProductionBuilding_Success(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockBuildingRepo()
	runRepo := newMockProductionRunRepo()
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	result, err := svc.BuildProductionBuilding(context.Background(), testCompanyID, testMasterBuildingID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if companyRepo.companies[testCompanyID].Money != 400 {
		t.Errorf("expected money=400, got %d", companyRepo.companies[testCompanyID].Money)
	}
	if result.Status != "constructing" {
		t.Errorf("expected status=constructing, got %s", result.Status)
	}
}

func TestProductionService_BuildProductionBuilding_BuildingNotFound(t *testing.T) {
	svc := NewProductionService(newMockCompanyRepo(), newMockInventoryRepo(), newMockBuildingRepo(), newMockProductionRunRepo(), cache.NewGamedataCache(), nil)
	_, err := svc.BuildProductionBuilding(context.Background(), testCompanyID, "nonexistent-master")
	if err != ErrBuildingNotFound {
		t.Errorf("expected ErrBuildingNotFound, got %v", err)
	}
}

func TestProductionService_BuildProductionBuilding_InsufficientFunds(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 10) // needs 100
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), newMockBuildingRepo(), newMockProductionRunRepo(), gc, nil)
	_, err := svc.BuildProductionBuilding(context.Background(), testCompanyID, testMasterBuildingID)
	if err != companyModels.ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

// ============================================================================
// UpgradeBuilding tests
// ============================================================================

func TestProductionService_UpgradeBuilding_Success(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockBuildingRepo()
	idleBuilding := newIdleBuilding("building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["building-1"] = idleBuilding
	runRepo := newMockProductionRunRepo()
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	result, err := svc.UpgradeBuilding(context.Background(), testCompanyID, "building-1", 1)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Status != "constructing" {
		t.Errorf("expected status=constructing after upgrade, got %s", result.Status)
	}
	if companyRepo.companies[testCompanyID].Money != 400 {
		t.Errorf("expected money=400, got %d", companyRepo.companies[testCompanyID].Money)
	}
}

func TestProductionService_UpgradeBuilding_NotIdle_UnderConstruction(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockBuildingRepo()
	constructing := newConstructingBuilding("building-1", testCompanyID, testBuildingDBID)
	buildingRepo.buildings["building-1"] = constructing
	runRepo := newMockProductionRunRepo()
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.UpgradeBuilding(context.Background(), testCompanyID, "building-1", 1)
	if err != ErrBuildingNotIdle {
		t.Errorf("expected ErrBuildingNotIdle, got %v", err)
	}
}

func TestProductionService_UpgradeBuilding_NotIdle_Producing(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockBuildingRepo()
	idleBuilding := newIdleBuilding("building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["building-1"] = idleBuilding
	runRepo := newMockProductionRunRepo()
	// Set an active run so building appears to be producing
	activeRun := &productionModels.ProductionRun{
		ID:                "run-active",
		CompanyBuildingID: "building-1",
		ProcessID:         testProcessDBID,
		ProductionCycles:  1,
		StartedAt:         time.Now().Add(-1 * time.Minute),
		EndsAt:            time.Now().Add(1 * time.Hour),
	}
	runRepo.activeRunByBuilding["building-1"] = activeRun
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.UpgradeBuilding(context.Background(), testCompanyID, "building-1", 1)
	if err != ErrBuildingNotIdle {
		t.Errorf("expected ErrBuildingNotIdle, got %v", err)
	}
}

// ============================================================================
// StartProduction tests
// ============================================================================

func TestProductionService_StartProduction_Success(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	inventoryRepo := newMockInventoryRepo()
	// Add enough input resources (need 2 per cycle × 1 cycle × level 1 = 2)
	inventoryRepo.items[testCompanyID] = map[string]*companyModels.CompanyInventoryItem{
		testInputResourceID: {ID: "inv-1", CompanyID: testCompanyID, ResourceID: testInputResourceID, Quantity: 10},
	}
	buildingRepo := newMockBuildingRepo()
	idleBuilding := newIdleBuilding("building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["building-1"] = idleBuilding
	runRepo := newMockProductionRunRepo()
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, inventoryRepo, buildingRepo, runRepo, gc, nil)
	result, err := svc.StartProduction(context.Background(), testCompanyID, "building-1", testProcessMasterID, 1)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Status != "producing" {
		t.Errorf("expected status=producing, got %s", result.Status)
	}
	// Input resources consumed: 2 units
	if inventoryRepo.items[testCompanyID][testInputResourceID].Quantity != 8 {
		t.Errorf("expected remaining inventory=8, got %d", inventoryRepo.items[testCompanyID][testInputResourceID].Quantity)
	}
}

func TestProductionService_StartProduction_NotIdle(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockBuildingRepo()
	constructing := newConstructingBuilding("building-1", testCompanyID, testBuildingDBID)
	buildingRepo.buildings["building-1"] = constructing
	runRepo := newMockProductionRunRepo()
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.StartProduction(context.Background(), testCompanyID, "building-1", testProcessMasterID, 1)
	if err != ErrBuildingNotIdle {
		t.Errorf("expected ErrBuildingNotIdle, got %v", err)
	}
}

func TestProductionService_StartProduction_ProcessNotFound(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockBuildingRepo()
	idleBuilding := newIdleBuilding("building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["building-1"] = idleBuilding
	runRepo := newMockProductionRunRepo()
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.StartProduction(context.Background(), testCompanyID, "building-1", "nonexistent-process", 1)
	if err != ErrProcessNotFound {
		t.Errorf("expected ErrProcessNotFound, got %v", err)
	}
}

func TestProductionService_StartProduction_InsufficientInventory(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	inventoryRepo := newMockInventoryRepo()
	// Only 1 unit of input, but need 2
	inventoryRepo.items[testCompanyID] = map[string]*companyModels.CompanyInventoryItem{
		testInputResourceID: {ID: "inv-1", CompanyID: testCompanyID, ResourceID: testInputResourceID, Quantity: 1},
	}
	buildingRepo := newMockBuildingRepo()
	idleBuilding := newIdleBuilding("building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["building-1"] = idleBuilding
	runRepo := newMockProductionRunRepo()
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, inventoryRepo, buildingRepo, runRepo, gc, nil)
	_, err := svc.StartProduction(context.Background(), testCompanyID, "building-1", testProcessMasterID, 1)
	if err != companyModels.ErrInsufficientInventory {
		t.Errorf("expected ErrInsufficientInventory, got %v", err)
	}
}

// ============================================================================
// CollectProduction tests
// ============================================================================

func TestProductionService_CollectProduction_Success(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 0)
	inventoryRepo := newMockInventoryRepo()
	buildingRepo := newMockBuildingRepo()
	idleBuilding := newIdleBuilding("building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["building-1"] = idleBuilding
	runRepo := newMockProductionRunRepo()
	// Completed run (endsAt in the past)
	completedRun := &productionModels.ProductionRun{
		ID:                "run-done",
		CompanyBuildingID: "building-1",
		ProcessID:         testProcessDBID,
		ProductionCycles:  2,
		StartedAt:         time.Now().Add(-2 * time.Hour),
		EndsAt:            time.Now().Add(-1 * time.Hour),
	}
	runRepo.runs["run-done"] = completedRun
	runRepo.activeRunByBuilding["building-1"] = completedRun
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, inventoryRepo, buildingRepo, runRepo, gc, nil)
	result, err := svc.CollectProduction(context.Background(), testCompanyID, "building-1")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Status != "idle" {
		t.Errorf("expected status=idle after collect, got %s", result.Status)
	}
	// Output: 5 units per cycle × 2 cycles × level 1 = 10
	if inventoryRepo.items[testCompanyID][testOutputResourceID] == nil {
		t.Fatal("expected output inventory item to exist")
	}
	if inventoryRepo.items[testCompanyID][testOutputResourceID].Quantity != 10 {
		t.Errorf("expected output quantity=10, got %d", inventoryRepo.items[testCompanyID][testOutputResourceID].Quantity)
	}
}

func TestProductionService_CollectProduction_NotComplete(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 0)
	buildingRepo := newMockBuildingRepo()
	idleBuilding := newIdleBuilding("building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["building-1"] = idleBuilding
	runRepo := newMockProductionRunRepo()
	// Run not yet completed (endsAt in the future)
	futureRun := &productionModels.ProductionRun{
		ID:                "run-future",
		CompanyBuildingID: "building-1",
		ProcessID:         testProcessDBID,
		ProductionCycles:  1,
		StartedAt:         time.Now().Add(-1 * time.Minute),
		EndsAt:            time.Now().Add(1 * time.Hour),
	}
	runRepo.runs["run-future"] = futureRun
	runRepo.activeRunByBuilding["building-1"] = futureRun
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.CollectProduction(context.Background(), testCompanyID, "building-1")
	if err != ErrProductionNotComplete {
		t.Errorf("expected ErrProductionNotComplete, got %v", err)
	}
}

func TestProductionService_CollectProduction_NotProducing(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 0)
	buildingRepo := newMockBuildingRepo()
	idleBuilding := newIdleBuilding("building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["building-1"] = idleBuilding
	runRepo := newMockProductionRunRepo() // no active run
	gc := newTestProductionCache()

	svc := NewProductionService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.CollectProduction(context.Background(), testCompanyID, "building-1")
	if err != ErrBuildingNotProducing {
		t.Errorf("expected ErrBuildingNotProducing, got %v", err)
	}
}
