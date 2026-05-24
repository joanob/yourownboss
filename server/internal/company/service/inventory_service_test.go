package service

import (
	"context"
	"testing"

	"github.com/joanob/yourownboss/internal/company/models"
)

// Test GetInventory - Success
func TestGetInventorySuccess(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"
	inventory := models.NewCompanyInventory(companyID)

	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByIDFunc: func(ctx context.Context, cid string) (*models.Company, error) {
			return &models.Company{
				ID:     cid,
				UserID: "user-123",
				Name:   "Test Company",
				Money:  1000,
			}, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{
		GetInventoryFunc: func(ctx context.Context, cid string) (*models.CompanyInventory, error) {
			return inventory, nil
		},
	}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	result, err := svc.GetInventory(ctx, companyID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Errorf("Expected inventory, got nil")
	}
}

// Test GetInventory - Company Not Found
func TestGetInventoryCompanyNotFound(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"

	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByIDFunc: func(ctx context.Context, cid string) (*models.Company, error) {
			return nil, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	_, err := svc.GetInventory(ctx, companyID)

	if err == nil {
		t.Errorf("Expected error for company not found, got nil")
	}
}

// Test AddResource - Success
func TestAddResourceSuccess(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"
	resourceID := "resource-123"
	quantity := int64(100)

	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByIDFunc: func(ctx context.Context, cid string) (*models.Company, error) {
			return &models.Company{
				ID:     cid,
				UserID: "user-123",
				Name:   "Test Company",
				Money:  1000,
			}, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{
		AddToInventoryFunc: func(ctx context.Context, cid, rid string, qty int64) (*models.CompanyInventoryItem, error) {
			return &models.CompanyInventoryItem{
				ID:         "item-123",
				CompanyID:  cid,
				ResourceID: rid,
				Quantity:   qty,
			}, nil
		},
	}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	item, err := svc.AddResource(ctx, companyID, resourceID, quantity)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if item == nil {
		t.Errorf("Expected item, got nil")
	}

	if item.Quantity != quantity {
		t.Errorf("Expected quantity %d, got %d", quantity, item.Quantity)
	}
}

// Test AddResource - Invalid Quantity
func TestAddResourceInvalidQuantity(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"
	resourceID := "resource-123"
	quantity := int64(-100)

	mockCompanyRepo := &MockCompanyRepository{}
	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	_, err := svc.AddResource(ctx, companyID, resourceID, quantity)

	if err == nil {
		t.Errorf("Expected error for invalid quantity, got nil")
	}
}

// Test RemoveResource - Success
func TestRemoveResourceSuccess(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"
	resourceID := "resource-123"
	quantity := int64(50)

	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByIDFunc: func(ctx context.Context, cid string) (*models.Company, error) {
			return &models.Company{
				ID:     cid,
				UserID: "user-123",
				Name:   "Test Company",
				Money:  1000,
			}, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{
		RemoveFromInventoryFunc: func(ctx context.Context, cid, rid string, qty int64) (*models.CompanyInventoryItem, error) {
			return &models.CompanyInventoryItem{
				ID:         "item-123",
				CompanyID:  cid,
				ResourceID: rid,
				Quantity:   50, // 100 - 50
			}, nil
		},
	}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	item, err := svc.RemoveResource(ctx, companyID, resourceID, quantity, false)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if item == nil {
		t.Errorf("Expected item, got nil")
	}
}

// Test RemoveResource - Invalid Quantity
func TestRemoveResourceInvalidQuantity(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"
	resourceID := "resource-123"
	quantity := int64(-50)

	mockCompanyRepo := &MockCompanyRepository{}
	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	_, err := svc.RemoveResource(ctx, companyID, resourceID, quantity, false)

	if err == nil {
		t.Errorf("Expected error for invalid quantity, got nil")
	}
}

// Test TransferResource - Success
func TestTransferResourceSuccess(t *testing.T) {
	ctx := context.Background()
	fromCompanyID := "company-123"
	toCompanyID := "company-456"
	resourceID := "resource-123"
	quantity := int64(50)

	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByIDFunc: func(ctx context.Context, cid string) (*models.Company, error) {
			return &models.Company{
				ID:     cid,
				UserID: "user-123",
				Name:   "Test Company",
				Money:  1000,
			}, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{
		RemoveFromInventoryFunc: func(ctx context.Context, cid, rid string, qty int64) (*models.CompanyInventoryItem, error) {
			return &models.CompanyInventoryItem{
				ID:         "item-123",
				CompanyID:  cid,
				ResourceID: rid,
				Quantity:   50,
			}, nil
		},
		AddToInventoryFunc: func(ctx context.Context, cid, rid string, qty int64) (*models.CompanyInventoryItem, error) {
			return &models.CompanyInventoryItem{
				ID:         "item-456",
				CompanyID:  cid,
				ResourceID: rid,
				Quantity:   qty,
			}, nil
		},
	}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	err := svc.TransferResource(ctx, fromCompanyID, toCompanyID, resourceID, quantity)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// Test TransferResource - Invalid Quantity
func TestTransferResourceInvalidQuantity(t *testing.T) {
	ctx := context.Background()
	fromCompanyID := "company-123"
	toCompanyID := "company-456"
	resourceID := "resource-123"
	quantity := int64(-50)

	mockCompanyRepo := &MockCompanyRepository{}
	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	err := svc.TransferResource(ctx, fromCompanyID, toCompanyID, resourceID, quantity)

	if err == nil {
		t.Errorf("Expected error for invalid quantity, got nil")
	}
}

// Test TransferResource - Source Company Not Found
func TestTransferResourceSourceCompanyNotFound(t *testing.T) {
	ctx := context.Background()
	fromCompanyID := "company-123"
	toCompanyID := "company-456"
	resourceID := "resource-123"
	quantity := int64(50)

	callCount := 0
	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByIDFunc: func(ctx context.Context, cid string) (*models.Company, error) {
			callCount++
			if callCount == 1 {
				return nil, nil
			}
			return &models.Company{
				ID:     cid,
				UserID: "user-123",
				Name:   "Test Company",
				Money:  1000,
			}, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewInventoryService(mockInventoryRepo, mockCompanyRepo)

	err := svc.TransferResource(ctx, fromCompanyID, toCompanyID, resourceID, quantity)

	if err == nil {
		t.Errorf("Expected error for source company not found, got nil")
	}
}
