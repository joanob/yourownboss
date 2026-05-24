package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// MockUserRepository implements repository.UserRepository for testing
type MockUserRepository struct {
	GetByIDFunc          func(context.Context, string) (*dbqueries.User, error)
	GetByUsernameFunc    func(context.Context, string) (*dbqueries.User, error)
	GetByEmailFunc       func(context.Context, string) (*dbqueries.User, error)
	CreateUserFunc       func(context.Context, *dbqueries.CreateUserParams) (*dbqueries.User, error)
	UpdateUserFunc       func(context.Context, *dbqueries.UpdateUserParams) (*dbqueries.User, error)
	SoftDeleteUserFunc   func(context.Context, string) error
	ExistsByUsernameFunc func(context.Context, string) (bool, error)
	ExistsByEmailFunc    func(context.Context, string) (bool, error)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*dbqueries.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*dbqueries.User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(ctx, username)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*dbqueries.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepository) CreateUser(ctx context.Context, params *dbqueries.CreateUserParams) (*dbqueries.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, params)
	}
	return nil, nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, params *dbqueries.UpdateUserParams) (*dbqueries.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, params)
	}
	return nil, nil
}

func (m *MockUserRepository) SoftDeleteUser(ctx context.Context, id string) error {
	if m.SoftDeleteUserFunc != nil {
		return m.SoftDeleteUserFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.ExistsByUsernameFunc != nil {
		return m.ExistsByUsernameFunc(ctx, username)
	}
	return false, nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.ExistsByEmailFunc != nil {
		return m.ExistsByEmailFunc(ctx, email)
	}
	return false, nil
}

// Test handlers

func TestRequireAdmin_WithAdminUser(t *testing.T) {
	adminUser := &dbqueries.User{
		ID:       "admin-id",
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "A", // Admin role
	}

	mockRepo := &MockUserRepository{
		GetByIDFunc: func(_ context.Context, id string) (*dbqueries.User, error) {
			if id == "admin-id" {
				return adminUser, nil
			}
			return nil, nil
		},
	}

	middleware := RequireAdmin(mockRepo)

	// Create a next handler that we expect to be called
	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(nextHandler)

	// Create request with user_id in context
	req := httptest.NewRequest("GET", "/api/v1/gamedata", nil)
	ctx := context.WithValue(req.Context(), "user_id", "admin-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected next handler to be called for admin user")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireAdmin_WithNonAdminUser(t *testing.T) {
	playerUser := &dbqueries.User{
		ID:       "player-id",
		Username: "player",
		Email:    "player@example.com",
		Role:     "P", // Player role (not admin)
	}

	mockRepo := &MockUserRepository{
		GetByIDFunc: func(_ context.Context, id string) (*dbqueries.User, error) {
			if id == "player-id" {
				return playerUser, nil
			}
			return nil, nil
		},
	}

	middleware := RequireAdmin(mockRepo)

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(nextHandler)

	req := httptest.NewRequest("GET", "/api/v1/gamedata", nil)
	ctx := context.WithValue(req.Context(), "user_id", "player-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Expected next handler to NOT be called for non-admin user")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response GenericResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error == nil {
		t.Error("Expected error response for non-admin user")
		return
	}

	// Type assert the error to get Code field
	errMap, ok := response.Error.(map[string]interface{})
	if !ok {
		t.Fatalf("Error response is not a map, got type: %T", response.Error)
	}

	code, exists := errMap["code"]
	if !exists || code != "INSUFFICIENT_PERMISSIONS" {
		t.Errorf("Expected INSUFFICIENT_PERMISSIONS error, got: %v", code)
	}
}

func TestRequireAdmin_WithoutUserID(t *testing.T) {
	mockRepo := &MockUserRepository{}

	middleware := RequireAdmin(mockRepo)

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(nextHandler)

	// Create request WITHOUT user_id in context
	req := httptest.NewRequest("GET", "/api/v1/gamedata", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Expected next handler to NOT be called when user_id is missing")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response GenericResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error == nil {
		t.Error("Expected error response when user_id is missing")
		return
	}

	// Type assert the error to get Code field
	errMap, ok := response.Error.(map[string]interface{})
	if !ok {
		t.Fatalf("Error response is not a map, got type: %T", response.Error)
	}

	code, exists := errMap["code"]
	if !exists || code != "UNAUTHORIZED" {
		t.Errorf("Expected UNAUTHORIZED error, got: %v", code)
	}
}

func TestRequireAdmin_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(_ context.Context, id string) (*dbqueries.User, error) {
			return nil, nil // User not found
		},
	}

	middleware := RequireAdmin(mockRepo)

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(nextHandler)

	req := httptest.NewRequest("GET", "/api/v1/gamedata", nil)
	ctx := context.WithValue(req.Context(), "user_id", "non-existent-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Expected next handler to NOT be called when user is not found")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response GenericResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error == nil {
		t.Error("Expected error response when user is not found")
		return
	}

	// Type assert the error to get Code field
	errMap, ok := response.Error.(map[string]interface{})
	if !ok {
		t.Fatalf("Error response is not a map, got type: %T", response.Error)
	}

	code, exists := errMap["code"]
	if !exists || code != "INSUFFICIENT_PERMISSIONS" {
		t.Errorf("Expected INSUFFICIENT_PERMISSIONS error, got: %v", code)
	}
}
