package service

import (
	"context"
	"testing"
	"time"

	"github.com/joanob/yourownboss/internal/company/models"
)

// MockCompanyRepository mocks the CompanyRepository interface
type MockCompanyRepository struct {
	CreateCompanyFunc      func(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error)
	GetCompanyByIDFunc     func(ctx context.Context, companyID string) (*models.Company, error)
	GetCompanyByUserIDFunc func(ctx context.Context, userID string) (*models.Company, error)
	UpdateCompanyNameFunc  func(ctx context.Context, companyID, newName string) (*models.Company, error)
	UpdateCompanyMoneyFunc func(ctx context.Context, companyID string, newMoney int64) (*models.Company, error)
	CheckCompanyExistsFunc func(ctx context.Context, companyID string) (bool, error)
	SoftDeleteCompanyFunc  func(ctx context.Context, companyID string) error
}

func (m *MockCompanyRepository) CreateCompany(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error) {
	return m.CreateCompanyFunc(ctx, userID, name, initialMoney)
}

func (m *MockCompanyRepository) GetCompanyByID(ctx context.Context, companyID string) (*models.Company, error) {
	return m.GetCompanyByIDFunc(ctx, companyID)
}

func (m *MockCompanyRepository) GetCompanyByUserID(ctx context.Context, userID string) (*models.Company, error) {
	return m.GetCompanyByUserIDFunc(ctx, userID)
}

func (m *MockCompanyRepository) UpdateCompanyName(ctx context.Context, companyID, newName string) (*models.Company, error) {
	return m.UpdateCompanyNameFunc(ctx, companyID, newName)
}

func (m *MockCompanyRepository) UpdateCompanyMoney(ctx context.Context, companyID string, newMoney int64) (*models.Company, error) {
	return m.UpdateCompanyMoneyFunc(ctx, companyID, newMoney)
}

func (m *MockCompanyRepository) CheckCompanyExists(ctx context.Context, companyID string) (bool, error) {
	return m.CheckCompanyExistsFunc(ctx, companyID)
}

func (m *MockCompanyRepository) SoftDeleteCompany(ctx context.Context, companyID string) error {
	return m.SoftDeleteCompanyFunc(ctx, companyID)
}

// MockInventoryRepository mocks the InventoryRepository interface
type MockInventoryRepository struct {
	GetInventoryFunc            func(ctx context.Context, companyID string) (*models.CompanyInventory, error)
	GetInventoryItemFunc        func(ctx context.Context, companyID, resourceID string) (*models.CompanyInventoryItem, error)
	AddToInventoryFunc          func(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
	RemoveFromInventoryFunc     func(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
	UpdateInventoryQuantityFunc func(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
	UpsertInventoryItemFunc     func(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
}

func (m *MockInventoryRepository) GetInventory(ctx context.Context, companyID string) (*models.CompanyInventory, error) {
	return m.GetInventoryFunc(ctx, companyID)
}

func (m *MockInventoryRepository) GetInventoryItem(ctx context.Context, companyID, resourceID string) (*models.CompanyInventoryItem, error) {
	return m.GetInventoryItemFunc(ctx, companyID, resourceID)
}

func (m *MockInventoryRepository) AddToInventory(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	return m.AddToInventoryFunc(ctx, companyID, resourceID, quantity)
}

func (m *MockInventoryRepository) RemoveFromInventory(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	return m.RemoveFromInventoryFunc(ctx, companyID, resourceID, quantity)
}

func (m *MockInventoryRepository) UpdateInventoryQuantity(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	return m.UpdateInventoryQuantityFunc(ctx, companyID, resourceID, quantity)
}

func (m *MockInventoryRepository) UpsertInventoryItem(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	return m.UpsertInventoryItemFunc(ctx, companyID, resourceID, quantity)
}

// Test CreateCompany - Success
func TestCreateCompanySuccess(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	companyName := "Test Company"
	initialMoney := int64(1000)

	mockCompanyRepo := &MockCompanyRepository{
		CreateCompanyFunc: func(ctx context.Context, uid, name string, money int64) (*models.Company, error) {
			return &models.Company{
				ID:        "company-123",
				UserID:    uid,
				Name:      name,
				Money:     money,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewCompanyService(mockCompanyRepo, mockInventoryRepo)

	company, err := svc.CreateCompany(ctx, userID, companyName, initialMoney)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if company == nil {
		t.Errorf("Expected company, got nil")
	}

	if company.Name != companyName {
		t.Errorf("Expected company name %s, got %s", companyName, company.Name)
	}

	if company.Money != initialMoney {
		t.Errorf("Expected money %d, got %d", initialMoney, company.Money)
	}
}

// Test CreateCompany - Empty Name
func TestCreateCompanyEmptyName(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	companyName := ""
	initialMoney := int64(1000)

	mockCompanyRepo := &MockCompanyRepository{}
	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewCompanyService(mockCompanyRepo, mockInventoryRepo)

	_, err := svc.CreateCompany(ctx, userID, companyName, initialMoney)

	if err == nil {
		t.Errorf("Expected error for empty name, got nil")
	}
}

// Test CreateCompany - Negative Money
func TestCreateCompanyNegativeMoney(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	companyName := "Test Company"
	initialMoney := int64(-100)

	mockCompanyRepo := &MockCompanyRepository{}
	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewCompanyService(mockCompanyRepo, mockInventoryRepo)

	_, err := svc.CreateCompany(ctx, userID, companyName, initialMoney)

	if err == nil {
		t.Errorf("Expected error for negative money, got nil")
	}
}

// Test GetCompany - Success
func TestGetCompanySuccess(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	expectedCompany := &models.Company{
		ID:        "company-123",
		UserID:    userID,
		Name:      "Test Company",
		Money:     1000,
		CreatedAt: time.Now(),
	}

	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByUserIDFunc: func(ctx context.Context, uid string) (*models.Company, error) {
			return expectedCompany, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewCompanyService(mockCompanyRepo, mockInventoryRepo)

	company, err := svc.GetCompany(ctx, userID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if company == nil {
		t.Errorf("Expected company, got nil")
	}

	if company.ID != expectedCompany.ID {
		t.Errorf("Expected company ID %s, got %s", expectedCompany.ID, company.ID)
	}
}

// Test GetCompany - Not Found
func TestGetCompanyNotFound(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByUserIDFunc: func(ctx context.Context, uid string) (*models.Company, error) {
			return nil, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewCompanyService(mockCompanyRepo, mockInventoryRepo)

	company, err := svc.GetCompany(ctx, userID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if company != nil {
		t.Errorf("Expected nil, got company")
	}
}

// Test UpdateCompanyName - Success
func TestUpdateCompanyNameSuccess(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"
	newName := "New Company Name"

	mockCompanyRepo := &MockCompanyRepository{
		GetCompanyByIDFunc: func(ctx context.Context, cid string) (*models.Company, error) {
			return &models.Company{
				ID:        cid,
				UserID:    "user-123",
				Name:      "Old Name",
				Money:     1000,
				CreatedAt: time.Now(),
			}, nil
		},
		UpdateCompanyNameFunc: func(ctx context.Context, cid, name string) (*models.Company, error) {
			return &models.Company{
				ID:        cid,
				UserID:    "user-123",
				Name:      name,
				Money:     1000,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewCompanyService(mockCompanyRepo, mockInventoryRepo)

	company, err := svc.UpdateCompanyName(ctx, companyID, newName)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if company.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, company.Name)
	}
}

// Test DeleteCompany - Success
func TestDeleteCompanySuccess(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"

	mockCompanyRepo := &MockCompanyRepository{
		CheckCompanyExistsFunc: func(ctx context.Context, cid string) (bool, error) {
			return true, nil
		},
		SoftDeleteCompanyFunc: func(ctx context.Context, cid string) error {
			return nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewCompanyService(mockCompanyRepo, mockInventoryRepo)

	err := svc.DeleteCompany(ctx, companyID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// Test DeleteCompany - Not Found
func TestDeleteCompanyNotFound(t *testing.T) {
	ctx := context.Background()
	companyID := "company-123"

	mockCompanyRepo := &MockCompanyRepository{
		CheckCompanyExistsFunc: func(ctx context.Context, cid string) (bool, error) {
			return false, nil
		},
	}

	mockInventoryRepo := &MockInventoryRepository{}

	svc := NewCompanyService(mockCompanyRepo, mockInventoryRepo)

	err := svc.DeleteCompany(ctx, companyID)

	if err == nil {
		t.Errorf("Expected error for company not found, got nil")
	}
}
