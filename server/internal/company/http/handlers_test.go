package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joanob/yourownboss/internal/company/models"
)

// MockCompanyService mocks the CompanyService interface
type MockCompanyService struct {
	CreateCompanyFunc     func(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error)
	GetCompanyFunc        func(ctx context.Context, userID string) (*models.Company, error)
	UpdateCompanyNameFunc func(ctx context.Context, companyID, newName string) (*models.Company, error)
	DeleteCompanyFunc     func(ctx context.Context, companyID string) error
}

func (m *MockCompanyService) CreateCompany(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error) {
	return m.CreateCompanyFunc(ctx, userID, name, initialMoney)
}

func (m *MockCompanyService) GetCompany(ctx context.Context, userID string) (*models.Company, error) {
	return m.GetCompanyFunc(ctx, userID)
}

func (m *MockCompanyService) UpdateCompanyName(ctx context.Context, companyID, newName string) (*models.Company, error) {
	return m.UpdateCompanyNameFunc(ctx, companyID, newName)
}

func (m *MockCompanyService) DeleteCompany(ctx context.Context, companyID string) error {
	return m.DeleteCompanyFunc(ctx, companyID)
}

// MockInventoryService mocks the InventoryService interface
type MockInventoryService struct {
	GetInventoryFunc     func(ctx context.Context, companyID string) (*models.CompanyInventory, error)
	AddResourceFunc      func(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
	RemoveResourceFunc   func(ctx context.Context, companyID, resourceID string, quantity int64, validateBefore bool) (*models.CompanyInventoryItem, error)
	TransferResourceFunc func(ctx context.Context, fromCompanyID, toCompanyID, resourceID string, quantity int64) error
}

func (m *MockInventoryService) GetInventory(ctx context.Context, companyID string) (*models.CompanyInventory, error) {
	return m.GetInventoryFunc(ctx, companyID)
}

func (m *MockInventoryService) AddResource(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	return m.AddResourceFunc(ctx, companyID, resourceID, quantity)
}

func (m *MockInventoryService) RemoveResource(ctx context.Context, companyID, resourceID string, quantity int64, validateBefore bool) (*models.CompanyInventoryItem, error) {
	return m.RemoveResourceFunc(ctx, companyID, resourceID, quantity, validateBefore)
}

func (m *MockInventoryService) TransferResource(ctx context.Context, fromCompanyID, toCompanyID, resourceID string, quantity int64) error {
	return m.TransferResourceFunc(ctx, fromCompanyID, toCompanyID, resourceID, quantity)
}

// Test GetCompanyHandler - Success
func TestGetCompanyHandlerSuccess(t *testing.T) {
	expectedCompany := &models.Company{
		ID:        "company-123",
		UserID:    "user-123",
		Name:      "Test Company",
		Money:     1000,
		CreatedAt: time.Now(),
	}

	mockService := &MockCompanyService{
		GetCompanyFunc: func(ctx context.Context, userID string) (*models.Company, error) {
			return expectedCompany, nil
		},
	}

	handler := GetCompanyHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/company", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse[CompanyDTO]
	json.NewDecoder(w.Body).Decode(&response)

	if response.Data.ID != expectedCompany.ID {
		t.Errorf("Expected company ID %s, got %s", expectedCompany.ID, response.Data.ID)
	}
}

// Test GetCompanyHandler - Unauthorized
func TestGetCompanyHandlerUnauthorized(t *testing.T) {
	mockService := &MockCompanyService{}

	handler := GetCompanyHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/company", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Test CreateCompanyHandler - Success
func TestCreateCompanyHandlerSuccess(t *testing.T) {
	requestBody := CreateCompanyRequest{
		Name:         "New Company",
		InitialMoney: 1000,
	}

	expectedCompany := &models.Company{
		ID:        "company-123",
		UserID:    "user-123",
		Name:      requestBody.Name,
		Money:     requestBody.InitialMoney,
		CreatedAt: time.Now(),
	}

	mockService := &MockCompanyService{
		CreateCompanyFunc: func(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error) {
			return expectedCompany, nil
		},
	}

	handler := CreateCompanyHandler(mockService)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/company", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

// Test CreateCompanyHandler - Validation Error
func TestCreateCompanyHandlerValidationError(t *testing.T) {
	requestBody := CreateCompanyRequest{
		Name:         "",
		InitialMoney: 1000,
	}

	mockService := &MockCompanyService{}

	handler := CreateCompanyHandler(mockService)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/company", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// Test CreateCompanyHandler - Company Already Exists
func TestCreateCompanyHandlerAlreadyExists(t *testing.T) {
	requestBody := CreateCompanyRequest{
		Name:         "New Company",
		InitialMoney: 1000,
	}

	mockService := &MockCompanyService{
		CreateCompanyFunc: func(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error) {
			return nil, models.ErrCompanyAlreadyExists
		},
	}

	handler := CreateCompanyHandler(mockService)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/company", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}
}

// Test UpdateCompanyHandler - Success
func TestUpdateCompanyHandlerSuccess(t *testing.T) {
	requestBody := UpdateCompanyRequest{
		Name: "Updated Company",
	}

	expectedCompany := &models.Company{
		ID:        "company-123",
		UserID:    "user-123",
		Name:      requestBody.Name,
		Money:     1000,
		CreatedAt: time.Now(),
	}

	mockService := &MockCompanyService{
		GetCompanyFunc: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{
				ID:        "company-123",
				UserID:    userID,
				Name:      "Old Name",
				Money:     1000,
				CreatedAt: time.Now(),
			}, nil
		},
		UpdateCompanyNameFunc: func(ctx context.Context, companyID, newName string) (*models.Company, error) {
			return expectedCompany, nil
		},
	}

	handler := UpdateCompanyHandler(mockService)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/api/v1/company", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// Test DeleteCompanyHandler - Success
func TestDeleteCompanyHandlerSuccess(t *testing.T) {
	mockService := &MockCompanyService{
		GetCompanyFunc: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{
				ID:        "company-123",
				UserID:    userID,
				Name:      "Test Company",
				Money:     1000,
				CreatedAt: time.Now(),
			}, nil
		},
		DeleteCompanyFunc: func(ctx context.Context, companyID string) error {
			return nil
		},
	}

	handler := DeleteCompanyHandler(mockService)

	req := httptest.NewRequest("DELETE", "/api/v1/company", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

// Test GetInventoryHandler - Success
func TestGetInventoryHandlerSuccess(t *testing.T) {
	inventory := models.NewCompanyInventory("company-123")

	mockService := &MockInventoryService{
		GetInventoryFunc: func(ctx context.Context, companyID string) (*models.CompanyInventory, error) {
			return inventory, nil
		},
	}

	handler := GetInventoryHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/company/inventory", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	req = req.WithContext(context.WithValue(req.Context(), "company_id", "company-123"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// Test GetInventoryHandler - Unauthorized
func TestGetInventoryHandlerUnauthorized(t *testing.T) {
	mockService := &MockInventoryService{}

	handler := GetInventoryHandler(mockService)

	req := httptest.NewRequest("GET", "/api/v1/company/inventory", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
