package crypto

import (
	"os"
	"testing"
	"time"
)

func TestJWTManager_GenerateSessionToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_SESSION_EXPIRY", "60")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_SESSION_EXPIRY")
	}()

	jwtManager, err := NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	userID := "user-123"
	sessionID := "session-456"
	companyID := "company-789"

	// Test: Generate session token successfully
	token, expiresAt, err := jwtManager.GenerateSessionToken(userID, sessionID, &companyID)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	if expiresAt.IsZero() {
		t.Error("Expected non-zero expiry time")
	}

	if expiresAt.Before(time.Now()) {
		t.Error("Expected future expiry time")
	}
}

func TestJWTManager_GenerateSessionToken_NilCompanyID(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_SESSION_EXPIRY", "60")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_SESSION_EXPIRY")
	}()

	jwtManager, err := NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	userID := "user-123"
	sessionID := "session-456"

	// Test: Generate session token with nil company ID
	token, _, err := jwtManager.GenerateSessionToken(userID, sessionID, nil)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Validate the token
	claims, err := jwtManager.ValidateSessionToken(token)
	if err != nil {
		t.Fatalf("ValidateSessionToken failed: %v", err)
	}

	if claims.CompanyID != nil {
		t.Error("Expected nil company ID")
	}
}

func TestJWTManager_ValidateSessionToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_SESSION_EXPIRY", "60")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_SESSION_EXPIRY")
	}()

	jwtManager, err := NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	userID := "user-123"
	sessionID := "session-456"
	companyID := "company-789"

	// Generate token
	token, _, err := jwtManager.GenerateSessionToken(userID, sessionID, &companyID)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	// Test: Validate session token
	claims, err := jwtManager.ValidateSessionToken(token)
	if err != nil {
		t.Fatalf("ValidateSessionToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, claims.SessionID)
	}

	if claims.CompanyID == nil || *claims.CompanyID != companyID {
		t.Errorf("Expected company ID %s, got %v", companyID, claims.CompanyID)
	}
}

func TestJWTManager_ValidateSessionToken_Invalid(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	defer os.Unsetenv("JWT_SECRET")

	jwtManager, err := NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	// Test: Validate invalid token
	_, err = jwtManager.ValidateSessionToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestJWTManager_GenerateRefreshToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_REFRESH_EXPIRY", "25920000")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_REFRESH_EXPIRY")
	}()

	jwtManager, err := NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	userID := "user-123"
	sessionID := "session-456"

	// Test: Generate refresh token
	token, verificationString, expiresAt, err := jwtManager.GenerateRefreshToken(userID, sessionID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	if verificationString == "" {
		t.Error("Expected non-empty verification string")
	}

	if expiresAt.IsZero() {
		t.Error("Expected non-zero expiry time")
	}

	if expiresAt.Before(time.Now()) {
		t.Error("Expected future expiry time")
	}
}

func TestJWTManager_ValidateRefreshToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_REFRESH_EXPIRY", "25920000")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_REFRESH_EXPIRY")
	}()

	jwtManager, err := NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	userID := "user-123"
	sessionID := "session-456"

	// Generate refresh token
	token, verificationString, _, err := jwtManager.GenerateRefreshToken(userID, sessionID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Test: Validate refresh token
	claims, err := jwtManager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, claims.SessionID)
	}

	if claims.VerificationString != verificationString {
		t.Errorf("Expected verification string %s, got %s", verificationString, claims.VerificationString)
	}
}

func TestJWTManager_ValidateRefreshToken_Expired(t *testing.T) {
	// Setup with very short expiry
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_REFRESH_EXPIRY", "1") // 1 second
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_REFRESH_EXPIRY")
	}()

	jwtManager, err := NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	userID := "user-123"
	sessionID := "session-456"

	// Generate refresh token
	token, _, _, err := jwtManager.GenerateRefreshToken(userID, sessionID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Test: Validate expired token
	_, err = jwtManager.ValidateRefreshToken(token)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestJWTManager_TokensAreUnique(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_SESSION_EXPIRY", "60")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_SESSION_EXPIRY")
	}()

	jwtManager, err := NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	// Generate two tokens
	token1, _, err := jwtManager.GenerateSessionToken("user-1", "session-1", nil)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	token2, _, err := jwtManager.GenerateSessionToken("user-2", "session-2", nil)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	// Test: Tokens should be different
	if token1 == token2 {
		t.Error("Expected tokens to be unique")
	}
}
