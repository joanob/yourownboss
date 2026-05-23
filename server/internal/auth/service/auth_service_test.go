package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/joanob/yourownboss/internal/auth"
	"github.com/joanob/yourownboss/internal/db/gen"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// ============================================================================
// Mock UserRepository for AuthService
// ============================================================================

type mockAuthUserRepository struct {
	users            map[string]*gen.User
	getByUsernameErr error
}

func newMockAuthUserRepository() *mockAuthUserRepository {
	return &mockAuthUserRepository{
		users: make(map[string]*gen.User),
	}
}

func (m *mockAuthUserRepository) CreateUser(ctx context.Context, params *gen.CreateUserParams) (*gen.User, error) {
	return nil, nil
}

func (m *mockAuthUserRepository) GetByUsername(ctx context.Context, username string) (*gen.User, error) {
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

func (m *mockAuthUserRepository) GetByID(ctx context.Context, userID string) (*gen.User, error) {
	return m.users[userID], nil
}

func (m *mockAuthUserRepository) GetByEmail(ctx context.Context, email string) (*gen.User, error) {
	return nil, nil
}

func (m *mockAuthUserRepository) UpdateUser(ctx context.Context, params *gen.UpdateUserParams) (*gen.User, error) {
	return nil, nil
}

func (m *mockAuthUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (m *mockAuthUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (m *mockAuthUserRepository) SoftDeleteUser(ctx context.Context, userID string) error {
	return nil
}

// ============================================================================
// Mock UserSessionRepository
// ============================================================================

type mockUserSessionRepository struct {
	sessions          map[string]*gen.UserSession
	createSessionErr  error
	getBySessionIDErr error
	revokeSessionErr  error
}

func newMockUserSessionRepository() *mockUserSessionRepository {
	return &mockUserSessionRepository{
		sessions: make(map[string]*gen.UserSession),
	}
}

func (m *mockUserSessionRepository) CreateSession(ctx context.Context, params *gen.CreateSessionParams) (*gen.UserSession, error) {
	if m.createSessionErr != nil {
		return nil, m.createSessionErr
	}

	session := &gen.UserSession{
		ID:                 params.ID,
		UserID:             params.UserID,
		SessionID:          params.SessionID,
		VerificationString: params.VerificationString,
		TokenHash:          params.TokenHash,
		ExpiresAt:          params.ExpiresAt,
		RevokedAt:          nil,
		CreatedAt:          time.Now().UTC(),
	}

	m.sessions[params.SessionID] = session
	return session, nil
}

func (m *mockUserSessionRepository) GetBySessionID(ctx context.Context, sessionID string) (*gen.UserSession, error) {
	if m.getBySessionIDErr != nil {
		return nil, m.getBySessionIDErr
	}

	return m.sessions[sessionID], nil
}

func (m *mockUserSessionRepository) GetByID(ctx context.Context, sessionRecordID string) (*gen.UserSession, error) {
	for _, session := range m.sessions {
		if session.ID == sessionRecordID {
			return session, nil
		}
	}
	return nil, nil
}

func (m *mockUserSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*gen.UserSession, error) {
	for _, session := range m.sessions {
		if session.TokenHash == tokenHash {
			return session, nil
		}
	}
	return nil, nil
}

func (m *mockUserSessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*gen.UserSession, error) {
	var sessions []*gen.UserSession
	for _, session := range m.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *mockUserSessionRepository) RevokeSession(ctx context.Context, sessionID string) error {
	if m.revokeSessionErr != nil {
		return m.revokeSessionErr
	}

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	now := time.Now().UTC()
	session.RevokedAt = &now
	m.sessions[sessionID] = session

	return nil
}

func (m *mockUserSessionRepository) CountActiveSessions(ctx context.Context, userID string) (int64, error) {
	count := int64(0)
	for _, session := range m.sessions {
		if session.UserID == userID && session.RevokedAt == nil && session.ExpiresAt.After(time.Now()) {
			count++
		}
	}
	return count, nil
}

func (m *mockUserSessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	toDelete := []string{}
	for sessionID, session := range m.sessions {
		if session.ExpiresAt.Before(time.Now()) {
			toDelete = append(toDelete, sessionID)
		}
	}
	for _, sessionID := range toDelete {
		delete(m.sessions, sessionID)
	}
	return nil
}

// ============================================================================
// Mock PasswordManager for AuthService
// ============================================================================

type mockAuthPasswordManager struct {
	verifyResult bool
	verifyErr    error
}

func newMockAuthPasswordManager() *mockAuthPasswordManager {
	return &mockAuthPasswordManager{
		verifyResult: true,
	}
}

func (m *mockAuthPasswordManager) HashPassword(password string) (string, error) {
	return fmt.Sprintf("$argon2id$mock$%s", password), nil
}

func (m *mockAuthPasswordManager) VerifyPassword(password, hash string) bool {
	return m.verifyResult
}

// ============================================================================
// Mock JWTManager
// ============================================================================

type mockJWTManager struct {
	generateSessionTokenErr error
	generateRefreshTokenErr error
}

func newMockJWTManager() *mockJWTManager {
	return &mockJWTManager{}
}

func (m *mockJWTManager) GenerateSessionToken(userID, sessionID string, companyID *string) (string, time.Time, error) {
	if m.generateSessionTokenErr != nil {
		return "", time.Time{}, m.generateSessionTokenErr
	}

	return "mock_session_token", time.Now().Add(1 * time.Minute), nil
}

func (m *mockJWTManager) GenerateRefreshToken(userID, sessionID string) (string, string, time.Time, error) {
	if m.generateRefreshTokenErr != nil {
		return "", "", time.Time{}, m.generateRefreshTokenErr
	}

	return "mock_refresh_token", "mock_verification_string", time.Now().Add(300 * 24 * time.Hour), nil
}

func (m *mockJWTManager) ValidateSessionToken(token string) (*auth.SessionTokenClaims, error) {
	return nil, nil
}

func (m *mockJWTManager) ValidateRefreshToken(token string) (*auth.RefreshTokenClaims, error) {
	return nil, nil
}

// ============================================================================
// Mock SessionCache
// ============================================================================

type mockSessionCache struct {
	data map[string]cache.SessionData
}

func newMockSessionCache() *mockSessionCache {
	return &mockSessionCache{
		data: make(map[string]cache.SessionData),
	}
}

func (m *mockSessionCache) Store(sessionID string, data cache.SessionData) {
	m.data[sessionID] = data
}

func (m *mockSessionCache) Get(sessionID string) (cache.SessionData, bool) {
	data, exists := m.data[sessionID]
	return data, exists
}

func (m *mockSessionCache) Revoke(sessionID string) {
	if data, exists := m.data[sessionID]; exists {
		now := time.Now().UTC()
		data.RevokedAt = &now
		m.data[sessionID] = data
	}
}

// ============================================================================
// Tests
// ============================================================================

func TestAuthService_Login_Success(t *testing.T) {
	// Setup
	mockUserRepo := newMockAuthUserRepository()
	mockSessionRepo := newMockUserSessionRepository()
	mockPM := newMockAuthPasswordManager()
	mockJWT := newMockJWTManager()
	mockSessionCache := newMockSessionCache()

	user := &gen.User{
		ID:           "user-123",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "$argon2id$mock$password123",
		Role:         "P",
	}
	mockUserRepo.users["user-123"] = user

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPM, mockJWT, mockSessionCache)

	// Test: Login with correct credentials
	response, err := service.Login(context.Background(), "alice", "password123")

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected login response")
	}

	if response.User.Username != "alice" {
		t.Errorf("Expected username alice, got %s", response.User.Username)
	}

	if response.SessionToken == "" {
		t.Error("Expected session token")
	}

	if response.RefreshToken == "" {
		t.Error("Expected refresh token")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Setup
	mockUserRepo := newMockAuthUserRepository()
	mockSessionRepo := newMockUserSessionRepository()
	mockPM := newMockAuthPasswordManager()
	mockJWT := newMockJWTManager()
	mockSessionCache := newMockSessionCache()

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPM, mockJWT, mockSessionCache)

	// Test: Login with non-existent user
	_, err := service.Login(context.Background(), "nonexistent", "password")

	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	if err.Error() != "invalid_credentials" {
		t.Errorf("Expected invalid_credentials, got %v", err)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Setup
	mockUserRepo := newMockAuthUserRepository()
	mockSessionRepo := newMockUserSessionRepository()
	mockPM := newMockAuthPasswordManager()
	mockPM.verifyResult = false
	mockJWT := newMockJWTManager()
	mockSessionCache := newMockSessionCache()

	user := &gen.User{
		ID:           "user-123",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "$argon2id$mock$password123",
		Role:         "P",
	}
	mockUserRepo.users["user-123"] = user

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPM, mockJWT, mockSessionCache)

	// Test: Login with wrong password
	_, err := service.Login(context.Background(), "alice", "wrongpassword")

	if err == nil {
		t.Error("Expected error for wrong password")
	}

	if err.Error() != "invalid_credentials" {
		t.Errorf("Expected invalid_credentials, got %v", err)
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	// Setup
	mockUserRepo := newMockAuthUserRepository()
	mockSessionRepo := newMockUserSessionRepository()
	mockPM := newMockAuthPasswordManager()
	mockJWT := newMockJWTManager()
	mockSessionCache := newMockSessionCache()

	// Create a session first
	session := &gen.UserSession{
		ID:        "session-record-123",
		UserID:    "user-123",
		SessionID: "session-123",
		CreatedAt: time.Now().UTC(),
	}
	mockSessionRepo.sessions["session-123"] = session

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPM, mockJWT, mockSessionCache)

	// Test: Logout
	err := service.Logout(context.Background(), "session-123")

	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Verify session is revoked in repository
	revokedSession := mockSessionRepo.sessions["session-123"]
	if revokedSession.RevokedAt == nil {
		t.Error("Expected session to be revoked")
	}
}

func TestAuthService_Logout_SessionNotFound(t *testing.T) {
	// Setup
	mockUserRepo := newMockAuthUserRepository()
	mockSessionRepo := newMockUserSessionRepository()
	mockPM := newMockAuthPasswordManager()
	mockJWT := newMockJWTManager()
	mockSessionCache := newMockSessionCache()

	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPM, mockJWT, mockSessionCache)

	// Test: Logout non-existent session
	err := service.Logout(context.Background(), "non-existent-session")

	if err == nil {
		t.Error("Expected error for non-existent session")
	}
}
