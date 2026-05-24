package service

import (
	"testing"

	"github.com/joanob/yourownboss/internal/pkg/cache"
	prodModels "github.com/joanob/yourownboss/internal/production/models"
	resModels "github.com/joanob/yourownboss/internal/resources/models"
	saleModels "github.com/joanob/yourownboss/internal/sale/models"
)

// TestValidateGamedataImport_ValidData tests validation with valid data
func TestValidateGamedataImport_ValidData(t *testing.T) {
	service := &GamedataService{
		gamedataCache: cache.NewGamedataCache(),
	}

	data := &GamedataImportRequest{
		Resources: []*resModels.Resource{
			{ID: "res-1", MasterID: "master-1", Name: "Water", MarketPrice: 10, MarketSaleQty: 1},
		},
		ProductionBuildings: []*prodModels.ProductionBuilding{
			{ID: "pb-1", MasterID: "master-pb-1", Name: "Factory", ConstructionCost: 1000, ConstructionTimeSec: 60},
		},
		SaleBuildings: []*saleModels.SaleBuilding{
			{ID: "sb-1", MasterID: "master-sb-1", Name: "Shop", ConstructionCost: 500, ConstructionTimeSec: 30},
		},
	}

	err := service.validateGamedataImport(data)
	if err != nil {
		t.Errorf("Expected no error for valid data, got: %v", err)
	}
}

// TestValidateGamedataImport_NilRequest tests validation with nil request
func TestValidateGamedataImport_NilRequest(t *testing.T) {
	service := &GamedataService{
		gamedataCache: cache.NewGamedataCache(),
	}

	err := service.validateGamedataImport(nil)
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}
}

// TestValidateGamedataImport_DuplicateResourceMasterID tests duplicate detection
func TestValidateGamedataImport_DuplicateResourceMasterID(t *testing.T) {
	service := &GamedataService{
		gamedataCache: cache.NewGamedataCache(),
	}

	data := &GamedataImportRequest{
		Resources: []*resModels.Resource{
			{ID: "res-1", MasterID: "master-1", Name: "Water", MarketPrice: 10, MarketSaleQty: 1},
			{ID: "res-2", MasterID: "master-1", Name: "Different", MarketPrice: 20, MarketSaleQty: 2},
		},
		ProductionBuildings: []*prodModels.ProductionBuilding{},
		SaleBuildings:       []*saleModels.SaleBuilding{},
	}

	err := service.validateGamedataImport(data)
	if err == nil {
		t.Error("Expected error for duplicate resource master_id, got nil")
	}
}

// TestValidateGamedataImport_MissingRequiredFields tests validation of required fields
func TestValidateGamedataImport_MissingRequiredFields(t *testing.T) {
	service := &GamedataService{
		gamedataCache: cache.NewGamedataCache(),
	}

	data := &GamedataImportRequest{
		Resources: []*resModels.Resource{
			{ID: "res-1", MasterID: "", Name: "Water", MarketPrice: 10, MarketSaleQty: 1},
		},
		ProductionBuildings: []*prodModels.ProductionBuilding{},
		SaleBuildings:       []*saleModels.SaleBuilding{},
	}

	err := service.validateGamedataImport(data)
	if err == nil {
		t.Error("Expected error for missing required fields, got nil")
	}
}

// TestValidateGamedataImport_DuplicateProductionBuildingMasterID tests duplicate building detection
func TestValidateGamedataImport_DuplicateProductionBuildingMasterID(t *testing.T) {
	service := &GamedataService{
		gamedataCache: cache.NewGamedataCache(),
	}

	data := &GamedataImportRequest{
		Resources: []*resModels.Resource{},
		ProductionBuildings: []*prodModels.ProductionBuilding{
			{ID: "pb-1", MasterID: "master-pb-1", Name: "Factory", ConstructionCost: 1000, ConstructionTimeSec: 60},
			{ID: "pb-2", MasterID: "master-pb-1", Name: "Factory2", ConstructionCost: 2000, ConstructionTimeSec: 120},
		},
		SaleBuildings: []*saleModels.SaleBuilding{},
	}

	err := service.validateGamedataImport(data)
	if err == nil {
		t.Error("Expected error for duplicate production building master_id, got nil")
	}
}

// TestValidateGamedataImport_DuplicateSaleBuildingMasterID tests duplicate sale building detection
func TestValidateGamedataImport_DuplicateSaleBuildingMasterID(t *testing.T) {
	service := &GamedataService{
		gamedataCache: cache.NewGamedataCache(),
	}

	data := &GamedataImportRequest{
		Resources:           []*resModels.Resource{},
		ProductionBuildings: []*prodModels.ProductionBuilding{},
		SaleBuildings: []*saleModels.SaleBuilding{
			{ID: "sb-1", MasterID: "master-sb-1", Name: "Shop", ConstructionCost: 500, ConstructionTimeSec: 30},
			{ID: "sb-2", MasterID: "master-sb-1", Name: "Shop2", ConstructionCost: 1000, ConstructionTimeSec: 60},
		},
	}

	err := service.validateGamedataImport(data)
	if err == nil {
		t.Error("Expected error for duplicate sale building master_id, got nil")
	}
}

// TestGamedataCacheIntegration tests gamedata cache operations
func TestGamedataCacheIntegration(t *testing.T) {
	testResources := []cache.Resource{
		{ID: "res-1", MasterID: "master-1", Name: "Water", MarketPrice: 10, MarketSaleQty: 1},
		{ID: "res-2", MasterID: "master-2", Name: "Tomato", MarketPrice: 20, MarketSaleQty: 3},
	}

	gameCache := cache.NewGamedataCache()
	gameCache.SetResources(testResources)

	cachedResources := gameCache.GetAllResources()
	if len(cachedResources) != 2 {
		t.Errorf("Expected 2 cached resources, got %d", len(cachedResources))
	}

	resource, found := gameCache.GetResource("res-1")
	if !found || resource.Name != "Water" {
		t.Errorf("Expected to find Water resource, got: %v (found: %v)", resource, found)
	}

	notFound, found := gameCache.GetResource("non-existent")
	if found {
		t.Errorf("Expected resource not found, got: %v", notFound)
	}
}
