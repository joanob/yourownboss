package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/joanob/yourownboss/internal/users/service"
)

// ============================================================================
// Mock UserService
// ============================================================================

type mockUserService struct {
	registerErr      error
	getUserErr       error
	updateUserErr    error
	deleteUserErr    error
	registerResult   *service.User
	getUserResult    *service.User
	updateUserResult *service.User
}

func (m *mockUserService) Register(ctx context.Context, username, email, password, timezone string) (*service.User, error) {
	if m.registerErr != nil {
		return nil, m.registerErr
	}
	if m.registerResult != nil {
		return m.registerResult, nil
	}
	return &service.User{
		ID:       "user-123",
		Username: username,
		Email:    email,
		Role:     "P",
	}, nil
}

func (m *mockUserService) GetUser(ctx context.Context, userID string) (*service.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	if m.getUserResult != nil {
		return m.getUserResult, nil
	}
	return &service.User{
		ID:       userID,
		Username: "alice",
		Email:    "alice@example.com",
		Role:     "P",
	}, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, userID string, updates *service.UserUpdateRequest) (*service.User, error) {
	if m.updateUserErr != nil {
		return nil, m.updateUserErr
	}
	if m.updateUserResult != nil {
		return m.updateUserResult, nil
	}
	return &service.User{
		ID:       userID,
		Username: "alice",
		Email:    "alice@example.com",
		Role:     "P",
	}, nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, userID string) error {
	return m.deleteUserErr
}

// ============================================================================
// Tests: GetMeHandler
// ============================================================================

func TestGetMeHandler_Success(t *testing.T) {
	mockUserService := &mockUserService{}
	handler := NewGetMeHandler(mockUserService)

	httpReq := httptest.NewRequest("GET", "/me", nil)
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetMeHandler_Unauthorized(t *testing.T) {
	mockUserService := &mockUserService{}
	handler := NewGetMeHandler(mockUserService)

	httpReq := httptest.NewRequest("GET", "/me", nil)
	// No user_id in context
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestGetMeHandler_NotFound(t *testing.T) {
	mockUserService := &mockUserService{
		getUserErr: fmt.Errorf("user not found"),
	}
	handler := NewGetMeHandler(mockUserService)

	httpReq := httptest.NewRequest("GET", "/me", nil)
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), "user_id", "nonexistent"))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code == http.StatusOK {
		t.Errorf("Expected error status, got 200")
	}
}

// ============================================================================
// Tests: UpdateMeHandler
// ============================================================================

func TestUpdateMeHandler_Success(t *testing.T) {
	mockUserService := &mockUserService{}
	validator := validator.New()
	handler := NewUpdateMeHandler(mockUserService, validator)

	req := UpdateMeRequest{
		Timezone: "America/New_York",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/me", bytes.NewReader(body))
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUpdateMeHandler_Unauthorized(t *testing.T) {
	mockUserService := &mockUserService{}
	validator := validator.New()
	handler := NewUpdateMeHandler(mockUserService, validator)

	req := UpdateMeRequest{
		Timezone: "America/New_York",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/me", bytes.NewReader(body))
	// No user_id in context
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestUpdateMeHandler_InvalidInput(t *testing.T) {
	mockUserService := &mockUserService{}
	validator := validator.New()
	handler := NewUpdateMeHandler(mockUserService, validator)

	// Invalid JSON
	httpReq := httptest.NewRequest("PUT", "/me", bytes.NewReader([]byte("invalid json")))
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), "user_id", "user-123"))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
