package service

import (
	"context"
	"testing"
	"time"

	companyModels "github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	saleModels "github.com/joanob/yourownboss/internal/sale/models"
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
func (m *mockCompanyRepo) GetCompanyByUserID(_ context.Context, _ string) (*companyModels.Company, error) {
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
func (m *mockInventoryRepo) UpdateInventoryQuantity(_ context.Context, _ string, _ int64) (*companyModels.CompanyInventoryItem, error) {
	return nil, nil
}
func (m *mockInventoryRepo) UpsertInventoryItem(_ context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
	return m.AddToInventory(context.Background(), companyID, resourceID, quantity)
}

// ============================================================================
// Mock CompanySaleBuildingRepository
// ============================================================================

type mockSaleBuildingRepo struct {
	buildings      map[string]*saleModels.CompanySaleBuilding
	createErr      error
	nextBuildingID string
}

func newMockSaleBuildingRepo() *mockSaleBuildingRepo {
	return &mockSaleBuildingRepo{
		buildings:      make(map[string]*saleModels.CompanySaleBuilding),
		nextBuildingID: "sale-building-1",
	}
}

func (m *mockSaleBuildingRepo) CreateCompanySaleBuilding(_ context.Context, companyID, saleBuildingID string, constructionEndsAt time.Time) (*saleModels.CompanySaleBuilding, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	b := &saleModels.CompanySaleBuilding{
		ID:                 m.nextBuildingID,
		CompanyID:          companyID,
		SaleBuildingID:     saleBuildingID,
		Level:              1,
		ConstructionEndsAt: &constructionEndsAt,
		CreatedAt:          time.Now().UTC(),
	}
	m.buildings[b.ID] = b
	return b, nil
}
func (m *mockSaleBuildingRepo) GetCompanySaleBuildingByID(_ context.Context, id string) (*saleModels.CompanySaleBuilding, error) {
	return m.buildings[id], nil
}
func (m *mockSaleBuildingRepo) GetCompanySaleBuildingsByCompanyID(_ context.Context, companyID string) ([]*saleModels.CompanySaleBuilding, error) {
	var result []*saleModels.CompanySaleBuilding
	for _, b := range m.buildings {
		if b.CompanyID == companyID {
			result = append(result, b)
		}
	}
	return result, nil
}
func (m *mockSaleBuildingRepo) UpdateBuildingLevel(_ context.Context, id string, level int64, constructionEndsAt time.Time) error {
	b := m.buildings[id]
	if b != nil {
		b.Level = level
		b.ConstructionEndsAt = &constructionEndsAt
	}
	return nil
}

// ============================================================================
// Mock SaleRunRepository
// ============================================================================

type mockSaleRunRepo struct {
	runs                map[string]*saleModels.SaleRun
	activeRunByBuilding map[string]*saleModels.SaleRun
	createErr           error
	nextRunID           string
}

func newMockSaleRunRepo() *mockSaleRunRepo {
	return &mockSaleRunRepo{
		runs:                make(map[string]*saleModels.SaleRun),
		activeRunByBuilding: make(map[string]*saleModels.SaleRun),
		nextRunID:           "sale-run-1",
	}
}

func (m *mockSaleRunRepo) CreateSaleRun(_ context.Context, companySaleBuildingID, resourceID string, unitsToSell int64, startedAt, endsAt time.Time) (*saleModels.SaleRun, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	r := &saleModels.SaleRun{
		ID:                    m.nextRunID,
		CompanySaleBuildingID: companySaleBuildingID,
		ResourceID:            resourceID,
		UnitsToSell:           unitsToSell,
		StartedAt:             startedAt,
		EndsAt:                endsAt,
	}
	m.runs[r.ID] = r
	m.activeRunByBuilding[companySaleBuildingID] = r
	return r, nil
}
func (m *mockSaleRunRepo) GetSaleRunByID(_ context.Context, id string) (*saleModels.SaleRun, error) {
	return m.runs[id], nil
}
func (m *mockSaleRunRepo) GetActiveRunByBuildingID(_ context.Context, companySaleBuildingID string) (*saleModels.SaleRun, error) {
	return m.activeRunByBuilding[companySaleBuildingID], nil
}
func (m *mockSaleRunRepo) MarkRunCollected(_ context.Context, id string, collectedAt time.Time) error {
	r := m.runs[id]
	if r != nil {
		r.IsCollected = true
		r.CollectedAt = &collectedAt
		for buildingID, run := range m.activeRunByBuilding {
			if run.ID == id {
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
	testMasterBuildingID = "master-market"
	testBuildingDBID     = "db-market-1"
	testResourceID       = "db-bread-1"
)

func newTestSaleCache() *cache.GamedataCache {
	gc := cache.NewGamedataCache()
	gc.SetSaleBuildings([]cache.SaleBuilding{
		{
			ID:                testBuildingDBID,
			MasterID:          testMasterBuildingID,
			Name:              "Market",
			ConstructionCost:  200,
			ConstructionTimeS: 120,
			Resources: []cache.SaleResource{
				{
					ResourceID:         testResourceID,
					PricePerUnit:       15,
					UnitsSoldPerSecond: 2,
				},
			},
		},
	})
	return gc
}

func newTestCompany(id string, money int64) *companyModels.Company {
	return &companyModels.Company{ID: id, UserID: "user-1", Name: "Test Co", Money: money, CreatedAt: time.Now().UTC()}
}

func newIdleSaleBuilding(id, companyID, buildingDBID string, level int64) *saleModels.CompanySaleBuilding {
	return &saleModels.CompanySaleBuilding{
		ID:             id,
		CompanyID:      companyID,
		SaleBuildingID: buildingDBID,
		Level:          level,
		CreatedAt:      time.Now().UTC(),
	}
}

func newConstructingSaleBuilding(id, companyID, buildingDBID string) *saleModels.CompanySaleBuilding {
	endsAt := time.Now().Add(1 * time.Hour)
	return &saleModels.CompanySaleBuilding{
		ID:                 id,
		CompanyID:          companyID,
		SaleBuildingID:     buildingDBID,
		Level:              1,
		ConstructionEndsAt: &endsAt,
		CreatedAt:          time.Now().UTC(),
	}
}

// ============================================================================
// BuildSaleBuilding tests
// ============================================================================

func TestSaleService_BuildSaleBuilding_Success(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockSaleBuildingRepo()
	runRepo := newMockSaleRunRepo()
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	result, err := svc.BuildSaleBuilding(context.Background(), testCompanyID, testMasterBuildingID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != "constructing" {
		t.Errorf("expected status=constructing, got %s", result.Status)
	}
	if companyRepo.companies[testCompanyID].Money != 300 {
		t.Errorf("expected money=300, got %d", companyRepo.companies[testCompanyID].Money)
	}
}

func TestSaleService_BuildSaleBuilding_BuildingNotFound(t *testing.T) {
	svc := NewSaleService(newMockCompanyRepo(), newMockInventoryRepo(), newMockSaleBuildingRepo(), newMockSaleRunRepo(), cache.NewGamedataCache(), nil)
	_, err := svc.BuildSaleBuilding(context.Background(), testCompanyID, "nonexistent-master")
	if err != ErrBuildingNotFound {
		t.Errorf("expected ErrBuildingNotFound, got %v", err)
	}
}

func TestSaleService_BuildSaleBuilding_InsufficientFunds(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 50) // needs 200
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), newMockSaleBuildingRepo(), newMockSaleRunRepo(), gc, nil)
	_, err := svc.BuildSaleBuilding(context.Background(), testCompanyID, testMasterBuildingID)
	if err != companyModels.ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

// ============================================================================
// UpgradeBuilding tests
// ============================================================================

func TestSaleService_UpgradeBuilding_Success(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockSaleBuildingRepo()
	idleBuilding := newIdleSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["sale-building-1"] = idleBuilding
	runRepo := newMockSaleRunRepo()
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	result, err := svc.UpgradeBuilding(context.Background(), testCompanyID, "sale-building-1", 1)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Status != "constructing" {
		t.Errorf("expected status=constructing after upgrade, got %s", result.Status)
	}
	if companyRepo.companies[testCompanyID].Money != 300 {
		t.Errorf("expected money=300, got %d", companyRepo.companies[testCompanyID].Money)
	}
}

func TestSaleService_UpgradeBuilding_NotIdle_UnderConstruction(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockSaleBuildingRepo()
	constructing := newConstructingSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID)
	buildingRepo.buildings["sale-building-1"] = constructing
	runRepo := newMockSaleRunRepo()
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.UpgradeBuilding(context.Background(), testCompanyID, "sale-building-1", 1)
	if err != ErrBuildingNotIdle {
		t.Errorf("expected ErrBuildingNotIdle, got %v", err)
	}
}

func TestSaleService_UpgradeBuilding_NotIdle_Selling(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockSaleBuildingRepo()
	idleBuilding := newIdleSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["sale-building-1"] = idleBuilding
	runRepo := newMockSaleRunRepo()
	activeRun := &saleModels.SaleRun{
		ID:                    "run-active",
		CompanySaleBuildingID: "sale-building-1",
		ResourceID:            testResourceID,
		UnitsToSell:           10,
		StartedAt:             time.Now().Add(-1 * time.Minute),
		EndsAt:                time.Now().Add(1 * time.Hour),
	}
	runRepo.activeRunByBuilding["sale-building-1"] = activeRun
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.UpgradeBuilding(context.Background(), testCompanyID, "sale-building-1", 1)
	if err != ErrBuildingNotIdle {
		t.Errorf("expected ErrBuildingNotIdle, got %v", err)
	}
}

// ============================================================================
// StartSale tests
// ============================================================================

func TestSaleService_StartSale_Success(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	inventoryRepo := newMockInventoryRepo()
	inventoryRepo.items[testCompanyID] = map[string]*companyModels.CompanyInventoryItem{
		testResourceID: {ID: "inv-1", CompanyID: testCompanyID, ResourceID: testResourceID, Quantity: 20},
	}
	buildingRepo := newMockSaleBuildingRepo()
	idleBuilding := newIdleSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["sale-building-1"] = idleBuilding
	runRepo := newMockSaleRunRepo()
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, inventoryRepo, buildingRepo, runRepo, gc, nil)
	result, err := svc.StartSale(context.Background(), testCompanyID, "sale-building-1", testResourceID, 10)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Status != "selling" {
		t.Errorf("expected status=selling, got %s", result.Status)
	}
	// Inventory should be reduced by 10
	if inventoryRepo.items[testCompanyID][testResourceID].Quantity != 10 {
		t.Errorf("expected remaining inventory=10, got %d", inventoryRepo.items[testCompanyID][testResourceID].Quantity)
	}
}

func TestSaleService_StartSale_NotIdle(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockSaleBuildingRepo()
	constructing := newConstructingSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID)
	buildingRepo.buildings["sale-building-1"] = constructing
	runRepo := newMockSaleRunRepo()
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.StartSale(context.Background(), testCompanyID, "sale-building-1", testResourceID, 10)
	if err != ErrBuildingNotIdle {
		t.Errorf("expected ErrBuildingNotIdle, got %v", err)
	}
}

func TestSaleService_StartSale_ResourceNotFound(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	buildingRepo := newMockSaleBuildingRepo()
	idleBuilding := newIdleSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["sale-building-1"] = idleBuilding
	runRepo := newMockSaleRunRepo()
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.StartSale(context.Background(), testCompanyID, "sale-building-1", "nonexistent-resource", 10)
	if err != ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestSaleService_StartSale_InsufficientInventory(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 500)
	inventoryRepo := newMockInventoryRepo()
	// Only 3 units, trying to sell 10
	inventoryRepo.items[testCompanyID] = map[string]*companyModels.CompanyInventoryItem{
		testResourceID: {ID: "inv-1", CompanyID: testCompanyID, ResourceID: testResourceID, Quantity: 3},
	}
	buildingRepo := newMockSaleBuildingRepo()
	idleBuilding := newIdleSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["sale-building-1"] = idleBuilding
	runRepo := newMockSaleRunRepo()
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, inventoryRepo, buildingRepo, runRepo, gc, nil)
	_, err := svc.StartSale(context.Background(), testCompanyID, "sale-building-1", testResourceID, 10)
	if err != companyModels.ErrInsufficientInventory {
		t.Errorf("expected ErrInsufficientInventory, got %v", err)
	}
}

// ============================================================================
// CollectSale tests
// ============================================================================

func TestSaleService_CollectSale_Success(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 100)
	buildingRepo := newMockSaleBuildingRepo()
	idleBuilding := newIdleSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["sale-building-1"] = idleBuilding
	runRepo := newMockSaleRunRepo()
	// Completed run (endsAt in the past): sold 10 units at price 15 each = 150 revenue
	completedRun := &saleModels.SaleRun{
		ID:                    "sale-run-done",
		CompanySaleBuildingID: "sale-building-1",
		ResourceID:            testResourceID,
		UnitsToSell:           10,
		StartedAt:             time.Now().Add(-2 * time.Hour),
		EndsAt:                time.Now().Add(-1 * time.Hour),
	}
	runRepo.runs["sale-run-done"] = completedRun
	runRepo.activeRunByBuilding["sale-building-1"] = completedRun
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	result, err := svc.CollectSale(context.Background(), testCompanyID, "sale-building-1")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Status != "idle" {
		t.Errorf("expected status=idle after collect, got %s", result.Status)
	}
	// Revenue = 10 * 15 = 150 → final money = 100 + 150 = 250
	if companyRepo.companies[testCompanyID].Money != 250 {
		t.Errorf("expected money=250, got %d", companyRepo.companies[testCompanyID].Money)
	}
}

func TestSaleService_CollectSale_NotComplete(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 100)
	buildingRepo := newMockSaleBuildingRepo()
	idleBuilding := newIdleSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["sale-building-1"] = idleBuilding
	runRepo := newMockSaleRunRepo()
	// Run not yet complete (endsAt in the future)
	futureRun := &saleModels.SaleRun{
		ID:                    "sale-run-future",
		CompanySaleBuildingID: "sale-building-1",
		ResourceID:            testResourceID,
		UnitsToSell:           10,
		StartedAt:             time.Now().Add(-1 * time.Minute),
		EndsAt:                time.Now().Add(1 * time.Hour),
	}
	runRepo.runs["sale-run-future"] = futureRun
	runRepo.activeRunByBuilding["sale-building-1"] = futureRun
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.CollectSale(context.Background(), testCompanyID, "sale-building-1")
	if err != ErrSaleNotComplete {
		t.Errorf("expected ErrSaleNotComplete, got %v", err)
	}
}

func TestSaleService_CollectSale_NotSelling(t *testing.T) {
	companyRepo := newMockCompanyRepo()
	companyRepo.companies[testCompanyID] = newTestCompany(testCompanyID, 100)
	buildingRepo := newMockSaleBuildingRepo()
	idleBuilding := newIdleSaleBuilding("sale-building-1", testCompanyID, testBuildingDBID, 1)
	buildingRepo.buildings["sale-building-1"] = idleBuilding
	runRepo := newMockSaleRunRepo() // no active run
	gc := newTestSaleCache()

	svc := NewSaleService(companyRepo, newMockInventoryRepo(), buildingRepo, runRepo, gc, nil)
	_, err := svc.CollectSale(context.Background(), testCompanyID, "sale-building-1")
	if err != ErrBuildingNotSelling {
		t.Errorf("expected ErrBuildingNotSelling, got %v", err)
	}
}
