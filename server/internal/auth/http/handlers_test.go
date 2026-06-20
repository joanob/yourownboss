package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joanob/yourownboss/internal/auth/service"
	userservice "github.com/joanob/yourownboss/internal/users/service"
)

// ============================================================================
// Mock UserService
// ============================================================================

type mockUserService struct {
	registerErr      error
	getUserErr       error
	updateUserErr    error
	deleteUserErr    error
	registerResult   *userservice.User
	getUserResult    *userservice.User
	updateUserResult *userservice.User
}

func (m *mockUserService) Register(ctx context.Context, username, email, password, timezone string) (*userservice.User, error) {
	if m.registerErr != nil {
		return nil, m.registerErr
	}
	if m.registerResult != nil {
		return m.registerResult, nil
	}
	return &userservice.User{
		ID:       "user-123",
		Username: username,
		Email:    email,
		Role:     "P",
	}, nil
}

func (m *mockUserService) GetUser(ctx context.Context, userID string) (*userservice.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	if m.getUserResult != nil {
		return m.getUserResult, nil
	}
	return &userservice.User{
		ID:       userID,
		Username: "alice",
		Email:    "alice@example.com",
		Role:     "P",
	}, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, userID string, updates *userservice.UserUpdateRequest) (*userservice.User, error) {
	if m.updateUserErr != nil {
		return nil, m.updateUserErr
	}
	if m.updateUserResult != nil {
		return m.updateUserResult, nil
	}
	return &userservice.User{
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
// Mock AuthService
// ============================================================================

type mockAuthService struct {
	loginErr    error
	logoutErr   error
	loginResult *service.LoginResponse
}

func (m *mockAuthService) Login(ctx context.Context, username, password string) (*service.LoginResponse, error) {
	if m.loginErr != nil {
		return nil, m.loginErr
	}
	if m.loginResult != nil {
		return m.loginResult, nil
	}
	return &service.LoginResponse{
		User: &service.User{
			ID:       "user-123",
			Username: username,
			Email:    fmt.Sprintf("%s@example.com", username),
			Role:     "P",
		},
		SessionToken: "mock_session_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Minute).Unix(),
	}, nil
}

func (m *mockAuthService) Logout(ctx context.Context, sessionID string) error {
	return m.logoutErr
}

// ============================================================================
// Tests: RegisterHandler
// ============================================================================

func TestRegisterHandler_Success(t *testing.T) {
	mockUserService := &mockUserService{}
	validator := validator.New()
	handler := NewRegisterHandler(mockUserService, validator, nil)

	req := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "SecurePassword123!",
		Timezone: "UTC",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestRegisterHandler_InvalidInput(t *testing.T) {
	mockUserService := &mockUserService{}
	validator := validator.New()
	handler := NewRegisterHandler(mockUserService, validator, nil)

	// Invalid JSON
	httpReq := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRegisterHandler_UsernameExists(t *testing.T) {
	mockUserService := &mockUserService{
		registerErr: fmt.Errorf("username already exists"),
	}
	validator := validator.New()
	handler := NewRegisterHandler(mockUserService, validator, nil)

	req := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "SecurePassword123!",
		Timezone: "UTC",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}
}

// ============================================================================
// Tests: LoginHandler
// ============================================================================

func TestLoginHandler_Success(t *testing.T) {
	mockAuthService := &mockAuthService{}
	validator := validator.New()
	handler := NewLoginHandler(mockAuthService, validator)

	req := LoginRequest{
		Username: "alice",
		Password: "SecurePassword123!",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response GenericResponse
	json.NewDecoder(w.Body).Decode(&response)
	if response.Data == nil {
		t.Error("Expected data in response")
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	mockAuthService := &mockAuthService{
		loginErr: fmt.Errorf("invalid_credentials"),
	}
	validator := validator.New()
	handler := NewLoginHandler(mockAuthService, validator)

	req := LoginRequest{
		Username: "alice",
		Password: "wrongpassword",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// ============================================================================
// Tests: LogoutHandler
// ============================================================================

func TestLogoutHandler_Success(t *testing.T) {
	mockAuthService := &mockAuthService{}
	handler := NewLogoutHandler(mockAuthService)

	httpReq := httptest.NewRequest("POST", "/logout", nil)
	httpReq.AddCookie(&http.Cookie{
		Name:     "session_token",
		Value:    "mock_session_token",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// Tests: LogoutHandler - Session Not Found
// ============================================================================

func TestLogoutHandler_SessionNotFound(t *testing.T) {
	mockAuthService := &mockAuthService{
		logoutErr: fmt.Errorf("session_not_found"),
	}
	handler := NewLogoutHandler(mockAuthService)

	httpReq := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, httpReq)

	// Should handle the error appropriately (could be 400, 401, or 404 depending on handler)
	if w.Code == http.StatusOK {
		t.Errorf("Expected error status, got 200")
	}
}
