package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// ============================================================================
// Mock UserRepository
// ============================================================================

type mockUserRepository struct {
	users               map[string]*dbqueries.User
	usernameExists      map[string]bool
	emailExists         map[string]bool
	createUserErr       error
	getByIDErr          error
	getByUsernameErr    error
	getByEmailErr       error
	updateUserErr       error
	softDeleteUserErr   error
	existsByUsernameErr error
	existsByEmailErr    error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:          make(map[string]*dbqueries.User),
		usernameExists: make(map[string]bool),
		emailExists:    make(map[string]bool),
	}
}

func (m *mockUserRepository) CreateUser(ctx context.Context, params *dbqueries.CreateUserParams) (*dbqueries.User, error) {
	if m.createUserErr != nil {
		return nil, m.createUserErr
	}

	user := &dbqueries.User{
		ID:           params.ID,
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		Role:         params.Role,
		Timezone:     params.Timezone,
		CreatedAt:    time.Now().UTC(),
	}

	m.users[params.ID] = user
	m.usernameExists[params.Username] = true
	m.emailExists[params.Email] = true

	return user, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*dbqueries.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.users[id], nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*dbqueries.User, error) {
	if m.getByUsernameErr != nil {
		return nil, m.getByUsernameErr
	}

	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*dbqueries.User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}

	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, nil
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, params *dbqueries.UpdateUserParams) (*dbqueries.User, error) {
	if m.updateUserErr != nil {
		return nil, m.updateUserErr
	}

	user, exists := m.users[params.ID]
	if !exists {
		return nil, nil
	}

	if params.Timezone != nil {
		user.Timezone = params.Timezone
	}
	if params.LastTimezoneModificationAt != nil {
		user.LastTimezoneModificationAt = params.LastTimezoneModificationAt
	}

	m.users[params.ID] = user
	return user, nil
}

func (m *mockUserRepository) SoftDeleteUser(ctx context.Context, id string) error {
	if m.softDeleteUserErr != nil {
		return m.softDeleteUserErr
	}

	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.existsByUsernameErr != nil {
		return false, m.existsByUsernameErr
	}

	return m.usernameExists[username], nil
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.existsByEmailErr != nil {
		return false, m.existsByEmailErr
	}

	return m.emailExists[email], nil
}

// ============================================================================
// Mock PasswordManager
// ============================================================================

type mockPasswordManager struct {
	hashErr      error
	verifyResult bool
	hashFunc     func(string) (string, error)
}

func newMockPasswordManager() *mockPasswordManager {
	return &mockPasswordManager{}
}

func (m *mockPasswordManager) HashPassword(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}

	if m.hashErr != nil {
		return "", m.hashErr
	}

	// Simple mock hash
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=4$mock_salt$mock_hash_%s", password), nil
}

func (m *mockPasswordManager) VerifyPassword(password, hash string) bool {
	return m.verifyResult
}

// ============================================================================
// Tests
// ============================================================================

func TestUserService_Register_Success(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Register new user
	user, err := service.Register(context.Background(), "alice", "alice@example.com", "password123", "Europe/Madrid")

	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user, got nil")
	}

	if user.Username != "alice" {
		t.Errorf("Expected username alice, got %s", user.Username)
	}

	if user.Email != "alice@example.com" {
		t.Errorf("Expected email alice@example.com, got %s", user.Email)
	}

	if user.Timezone == nil || *user.Timezone != "Europe/Madrid" {
		t.Errorf("Expected timezone Europe/Madrid, got %v", user.Timezone)
	}
}

func TestUserService_Register_UsernameExists(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockRepo.usernameExists["alice"] = true
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Register with existing username
	_, err := service.Register(context.Background(), "alice", "alice@example.com", "password123", "Europe/Madrid")

	if err == nil {
		t.Error("Expected error for existing username")
	}

	if err.Error() != "username_already_exists" {
		t.Errorf("Expected username_already_exists, got %v", err)
	}
}

func TestUserService_Register_EmailExists(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockRepo.emailExists["alice@example.com"] = true
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Register with existing email
	_, err := service.Register(context.Background(), "alice", "alice@example.com", "password123", "Europe/Madrid")

	if err == nil {
		t.Error("Expected error for existing email")
	}

	if err.Error() != "email_already_exists" {
		t.Errorf("Expected email_already_exists, got %v", err)
	}
}

func TestUserService_Register_RepositoryError(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockRepo.createUserErr = fmt.Errorf("database error")
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Register with repository error
	_, err := service.Register(context.Background(), "alice", "alice@example.com", "password123", "Europe/Madrid")

	if err == nil {
		t.Error("Expected error")
	}
}

func TestUserService_Register_PasswordHashError(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockPM := newMockPasswordManager()
	mockPM.hashErr = fmt.Errorf("hash error")
	service := NewUserService(mockRepo, mockPM)

	// Test: Register with password hash error
	_, err := service.Register(context.Background(), "alice", "alice@example.com", "password123", "Europe/Madrid")

	if err == nil {
		t.Error("Expected error for hash failure")
	}
}

func TestUserService_GetUser_Found(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	user := &dbqueries.User{
		ID:       "user-123",
		Username: "alice",
		Email:    "alice@example.com",
		Role:     "P",
		Timezone: stringPtr("Europe/Madrid"),
	}
	mockRepo.users["user-123"] = user
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Get existing user
	retrievedUser, err := service.GetUser(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if retrievedUser == nil {
		t.Fatal("Expected user, got nil")
	}

	if retrievedUser.Username != "alice" {
		t.Errorf("Expected username alice, got %s", retrievedUser.Username)
	}
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Get non-existent user
	retrievedUser, err := service.GetUser(context.Background(), "non-existent")

	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if retrievedUser != nil {
		t.Error("Expected nil for non-existent user")
	}
}

func TestUserService_GetUser_RepositoryError(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockRepo.getByIDErr = fmt.Errorf("database error")
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Get user with repository error
	_, err := service.GetUser(context.Background(), "user-123")

	if err == nil {
		t.Error("Expected error")
	}
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	newTimezone := "America/NewYork"
	user := &dbqueries.User{
		ID:       "user-123",
		Username: "alice",
		Email:    "alice@example.com",
		Role:     "P",
		Timezone: stringPtr("Europe/Madrid"),
	}
	mockRepo.users["user-123"] = user
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Update timezone (no previous modification)
	updatedUser, err := service.UpdateUser(context.Background(), "user-123", &UserUpdateRequest{
		Timezone: &newTimezone,
	})

	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	if updatedUser == nil {
		t.Fatal("Expected user, got nil")
	}

	if updatedUser.Timezone == nil || *updatedUser.Timezone != newTimezone {
		t.Errorf("Expected timezone America/NewYork, got %v", updatedUser.Timezone)
	}
}

func TestUserService_UpdateUser_TimezoneRestriction(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	newTimezone := "America/NewYork"
	lastMod := time.Now().Add(-10 * time.Hour) // Only 10 hours ago
	user := &dbqueries.User{
		ID:                         "user-123",
		Username:                   "alice",
		Email:                      "alice@example.com",
		Role:                       "P",
		Timezone:                   stringPtr("Europe/Madrid"),
		LastTimezoneModificationAt: &lastMod,
	}
	mockRepo.users["user-123"] = user
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Try to update timezone too soon
	_, err := service.UpdateUser(context.Background(), "user-123", &UserUpdateRequest{
		Timezone: &newTimezone,
	})

	if err == nil {
		t.Error("Expected error for timezone change too soon")
	}

	if err.Error() != "timezone_recently_changed" {
		t.Errorf("Expected timezone_recently_changed, got %v", err)
	}
}

func TestUserService_UpdateUser_UserNotFound(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	newTimezone := "America/NewYork"

	// Test: Update non-existent user
	_, err := service.UpdateUser(context.Background(), "non-existent", &UserUpdateRequest{
		Timezone: &newTimezone,
	})

	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	if err.Error() != "user_not_found" {
		t.Errorf("Expected user_not_found, got %v", err)
	}
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	user := &dbqueries.User{
		ID:       "user-123",
		Username: "alice",
		Email:    "alice@example.com",
		Role:     "P",
	}
	mockRepo.users["user-123"] = user
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Delete user
	err := service.DeleteUser(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify user is deleted
	retrievedUser, _ := service.GetUser(context.Background(), "user-123")
	if retrievedUser != nil {
		t.Error("Expected user to be deleted")
	}
}

func TestUserService_DeleteUser_RepositoryError(t *testing.T) {
	// Setup
	mockRepo := newMockUserRepository()
	mockRepo.softDeleteUserErr = fmt.Errorf("database error")
	mockPM := newMockPasswordManager()
	service := NewUserService(mockRepo, mockPM)

	// Test: Delete with repository error
	err := service.DeleteUser(context.Background(), "user-123")

	if err == nil {
		t.Error("Expected error")
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
