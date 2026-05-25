package service

import (
	"context"
	"errors"
	"testing"
	"time"

	companyModels "github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// ============================================================================
// Mock implementations
// ============================================================================

type mockCompanyRepo struct {
	GetCompanyByIDFunc     func(ctx context.Context, companyID string) (*companyModels.Company, error)
	UpdateCompanyMoneyFunc func(ctx context.Context, companyID string, newMoney int64) (*companyModels.Company, error)
}

func (m *mockCompanyRepo) CreateCompany(ctx context.Context, userID, name string, initialMoney int64) (*companyModels.Company, error) {
	return nil, nil
}
func (m *mockCompanyRepo) GetCompanyByID(ctx context.Context, companyID string) (*companyModels.Company, error) {
	return m.GetCompanyByIDFunc(ctx, companyID)
}
func (m *mockCompanyRepo) GetCompanyByUserID(ctx context.Context, userID string) (*companyModels.Company, error) {
	return nil, nil
}
func (m *mockCompanyRepo) UpdateCompanyName(ctx context.Context, companyID, newName string) (*companyModels.Company, error) {
	return nil, nil
}
func (m *mockCompanyRepo) UpdateCompanyMoney(ctx context.Context, companyID string, newMoney int64) (*companyModels.Company, error) {
	return m.UpdateCompanyMoneyFunc(ctx, companyID, newMoney)
}
func (m *mockCompanyRepo) CheckCompanyExists(ctx context.Context, companyID string) (bool, error) {
	return false, nil
}
func (m *mockCompanyRepo) SoftDeleteCompany(ctx context.Context, companyID string) error {
	return nil
}

type mockInventoryRepo struct {
	GetInventoryFunc        func(ctx context.Context, companyID string) (*companyModels.CompanyInventory, error)
	GetInventoryItemFunc    func(ctx context.Context, companyID, resourceID string) (*companyModels.CompanyInventoryItem, error)
	AddToInventoryFunc      func(ctx context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error)
	RemoveFromInventoryFunc func(ctx context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error)
}

func (m *mockInventoryRepo) GetInventory(ctx context.Context, companyID string) (*companyModels.CompanyInventory, error) {
	return m.GetInventoryFunc(ctx, companyID)
}
func (m *mockInventoryRepo) GetInventoryItem(ctx context.Context, companyID, resourceID string) (*companyModels.CompanyInventoryItem, error) {
	return m.GetInventoryItemFunc(ctx, companyID, resourceID)
}
func (m *mockInventoryRepo) AddToInventory(ctx context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
	return m.AddToInventoryFunc(ctx, companyID, resourceID, quantity)
}
func (m *mockInventoryRepo) RemoveFromInventory(ctx context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
	return m.RemoveFromInventoryFunc(ctx, companyID, resourceID, quantity)
}
func (m *mockInventoryRepo) UpdateInventoryQuantity(ctx context.Context, itemID string, newQuantity int64) (*companyModels.CompanyInventoryItem, error) {
	return nil, nil
}
func (m *mockInventoryRepo) UpsertInventoryItem(ctx context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
	return nil, nil
}

// ============================================================================
// Helpers
// ============================================================================

func newTestCache() *cache.GamedataCache {
	c := cache.NewGamedataCache()
	c.SetResources([]cache.Resource{
		{ID: "res-uuid-water", MasterID: "water", Name: "Water", MarketPrice: 10, MarketSaleQty: 5},
	})
	return c
}

func newTestCompany(id string, money int64) *companyModels.Company {
	return &companyModels.Company{
		ID:        id,
		UserID:    "user-1",
		Name:      "Test Co",
		Money:     money,
		CreatedAt: time.Now(),
	}
}

func newEmptyInventory(companyID string) *companyModels.CompanyInventory {
	return companyModels.NewCompanyInventory(companyID)
}

// ============================================================================
// BuyResource tests
// ============================================================================

func TestBuyResource_Success(t *testing.T) {
	gameCache := newTestCache()

	companyRepo := &mockCompanyRepo{
		GetCompanyByIDFunc: func(ctx context.Context, companyID string) (*companyModels.Company, error) {
			return newTestCompany(companyID, 100), nil
		},
		UpdateCompanyMoneyFunc: func(ctx context.Context, companyID string, newMoney int64) (*companyModels.Company, error) {
			return newTestCompany(companyID, newMoney), nil
		},
	}
	inventoryRepo := &mockInventoryRepo{
		AddToInventoryFunc: func(ctx context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
			return &companyModels.CompanyInventoryItem{
				ID:         "inv-1",
				CompanyID:  companyID,
				ResourceID: resourceID,
				Quantity:   quantity,
			}, nil
		},
		GetInventoryFunc: func(ctx context.Context, companyID string) (*companyModels.CompanyInventory, error) {
			return newEmptyInventory(companyID), nil
		},
	}

	svc := NewMarketService(companyRepo, inventoryRepo, gameCache, nil)

	// water: MarketPrice=10, MarketSaleQty=5; buy 10 → totalCost = 10*(10/5) = 20
	result, err := svc.BuyResource(context.Background(), "co-1", "water", 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Company.Money != 80 {
		t.Errorf("Expected money=80, got: %d", result.Company.Money)
	}
}

func TestBuyResource_ResourceNotFound(t *testing.T) {
	gameCache := newTestCache()

	svc := NewMarketService(&mockCompanyRepo{}, &mockInventoryRepo{}, gameCache, nil)

	_, err := svc.BuyResource(context.Background(), "co-1", "nonexistent", 5)
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("Expected ErrResourceNotFound, got: %v", err)
	}
}

func TestBuyResource_InvalidQuantityMultiple(t *testing.T) {
	gameCache := newTestCache()

	svc := NewMarketService(&mockCompanyRepo{}, &mockInventoryRepo{}, gameCache, nil)

	// water has MarketSaleQty=5; 7 is not a multiple of 5
	_, err := svc.BuyResource(context.Background(), "co-1", "water", 7)
	if !errors.Is(err, ErrInvalidQuantityMultiple) {
		t.Errorf("Expected ErrInvalidQuantityMultiple, got: %v", err)
	}
}

func TestBuyResource_InsufficientFunds(t *testing.T) {
	gameCache := newTestCache()

	companyRepo := &mockCompanyRepo{
		GetCompanyByIDFunc: func(ctx context.Context, companyID string) (*companyModels.Company, error) {
			// company has only 5 money; buying 10 water costs 20
			return newTestCompany(companyID, 5), nil
		},
		UpdateCompanyMoneyFunc: func(ctx context.Context, companyID string, newMoney int64) (*companyModels.Company, error) {
			return newTestCompany(companyID, newMoney), nil
		},
	}

	svc := NewMarketService(companyRepo, &mockInventoryRepo{}, gameCache, nil)

	_, err := svc.BuyResource(context.Background(), "co-1", "water", 10)
	if !errors.Is(err, companyModels.ErrInsufficientFunds) {
		t.Errorf("Expected ErrInsufficientFunds, got: %v", err)
	}
}

// ============================================================================
// SellResource tests
// ============================================================================

func TestSellResource_Success(t *testing.T) {
	gameCache := newTestCache()

	companyRepo := &mockCompanyRepo{
		GetCompanyByIDFunc: func(ctx context.Context, companyID string) (*companyModels.Company, error) {
			return newTestCompany(companyID, 50), nil
		},
		UpdateCompanyMoneyFunc: func(ctx context.Context, companyID string, newMoney int64) (*companyModels.Company, error) {
			return newTestCompany(companyID, newMoney), nil
		},
	}
	inventoryRepo := &mockInventoryRepo{
		GetInventoryItemFunc: func(ctx context.Context, companyID, resourceID string) (*companyModels.CompanyInventoryItem, error) {
			return &companyModels.CompanyInventoryItem{
				ID:         "inv-1",
				CompanyID:  companyID,
				ResourceID: resourceID,
				Quantity:   20,
			}, nil
		},
		RemoveFromInventoryFunc: func(ctx context.Context, companyID, resourceID string, quantity int64) (*companyModels.CompanyInventoryItem, error) {
			return &companyModels.CompanyInventoryItem{
				ID:         "inv-1",
				CompanyID:  companyID,
				ResourceID: resourceID,
				Quantity:   20 - quantity,
			}, nil
		},
		GetInventoryFunc: func(ctx context.Context, companyID string) (*companyModels.CompanyInventory, error) {
			return newEmptyInventory(companyID), nil
		},
	}

	svc := NewMarketService(companyRepo, inventoryRepo, gameCache, nil)

	// water: MarketPrice=10, MarketSaleQty=5; sell 10 → revenue = 10*(10/5) = 20
	result, err := svc.SellResource(context.Background(), "co-1", "water", 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Company.Money != 70 {
		t.Errorf("Expected money=70, got: %d", result.Company.Money)
	}
}

func TestSellResource_ResourceNotFound(t *testing.T) {
	gameCache := newTestCache()

	svc := NewMarketService(&mockCompanyRepo{}, &mockInventoryRepo{}, gameCache, nil)

	_, err := svc.SellResource(context.Background(), "co-1", "nonexistent", 5)
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("Expected ErrResourceNotFound, got: %v", err)
	}
}

func TestSellResource_InvalidQuantityMultiple(t *testing.T) {
	gameCache := newTestCache()

	svc := NewMarketService(&mockCompanyRepo{}, &mockInventoryRepo{}, gameCache, nil)

	// water has MarketSaleQty=5; 3 is not a multiple of 5
	_, err := svc.SellResource(context.Background(), "co-1", "water", 3)
	if !errors.Is(err, ErrInvalidQuantityMultiple) {
		t.Errorf("Expected ErrInvalidQuantityMultiple, got: %v", err)
	}
}

func TestSellResource_InsufficientInventory(t *testing.T) {
	gameCache := newTestCache()

	inventoryRepo := &mockInventoryRepo{
		GetInventoryItemFunc: func(ctx context.Context, companyID, resourceID string) (*companyModels.CompanyInventoryItem, error) {
			// Only 3 units available, trying to sell 10
			return &companyModels.CompanyInventoryItem{
				ID:         "inv-1",
				CompanyID:  companyID,
				ResourceID: resourceID,
				Quantity:   3,
			}, nil
		},
	}

	svc := NewMarketService(&mockCompanyRepo{}, inventoryRepo, gameCache, nil)

	_, err := svc.SellResource(context.Background(), "co-1", "water", 10)
	if !errors.Is(err, companyModels.ErrInsufficientInventory) {
		t.Errorf("Expected ErrInsufficientInventory, got: %v", err)
	}
}

func TestSellResource_NoInventoryItem(t *testing.T) {
	gameCache := newTestCache()

	inventoryRepo := &mockInventoryRepo{
		GetInventoryItemFunc: func(ctx context.Context, companyID, resourceID string) (*companyModels.CompanyInventoryItem, error) {
			return nil, nil // no inventory item
		},
	}

	svc := NewMarketService(&mockCompanyRepo{}, inventoryRepo, gameCache, nil)

	_, err := svc.SellResource(context.Background(), "co-1", "water", 5)
	if !errors.Is(err, companyModels.ErrInsufficientInventory) {
		t.Errorf("Expected ErrInsufficientInventory, got: %v", err)
	}
}
