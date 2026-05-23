package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joanob/yourownboss/internal/auth"
	"github.com/joanob/yourownboss/internal/db/gen"
	"github.com/joanob/yourownboss/internal/users/repository"
	"github.com/rs/zerolog/log"
)

// AuthService defines the authentication business logic.
type AuthService interface {
	// Login authenticates a user with username and password.
	// Returns session token, refresh token, and user info or error.
	// Errors: invalid_credentials, user_not_found, internal errors
	Login(ctx context.Context, username, password string) (*LoginResponse, error)

	// Logout revokes the user's current session.
	// Returns error if operation fails.
	Logout(ctx context.Context, sessionID string) error
}

// LoginResponse contains the response data after successful login.
type LoginResponse struct {
	User         *User  `json:"user"`
	SessionToken string `json:"session_token"`      // JWT session token (short-lived, 1 min)
	RefreshToken string `json:"refresh_token"`      // JWT refresh token (long-lived, 300 days)
	ExpiresAt    int64  `json:"session_expires_at"` // Unix timestamp when session expires
}

// User represents a user returned from the service layer.
type User struct {
	ID                         string     `json:"id"`
	Username                   string     `json:"username"`
	Email                      string     `json:"email"`
	Role                       string     `json:"role"`
	Timezone                   *string    `json:"timezone"`
	LastTimezoneModificationAt *time.Time `json:"last_timezone_modification_at"`
	CreatedAt                  time.Time  `json:"created_at"`
}

// authService implements AuthService.
type authService struct {
	userRepo        repository.UserRepository
	sessionRepo     auth.UserSessionRepository
	passwordManager *auth.PasswordManager
	jwtManager      *auth.JWTManager
	sessionCache    *auth.SessionCache
}

// NewAuthService creates a new auth service with dependency injection.
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo auth.UserSessionRepository,
	passwordManager *auth.PasswordManager,
	jwtManager *auth.JWTManager,
	sessionCache *auth.SessionCache,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		passwordManager: passwordManager,
		jwtManager:      jwtManager,
		sessionCache:    sessionCache,
	}
}

// Login authenticates a user and creates a session.
func (s *authService) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	log.Info().Str("username", username).Msg("Login attempt")

	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to get user by username")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		log.Warn().Str("username", username).Msg("Login failed: user not found")
		return nil, fmt.Errorf("invalid_credentials")
	}

	// Verify password
	if !s.passwordManager.VerifyPassword(password, user.PasswordHash) {
		log.Warn().Str("username", username).Msg("Login failed: invalid password")
		return nil, fmt.Errorf("invalid_credentials")
	}

	// Generate session IDs
	sessionID := uuid.New().String()
	verificationString := uuid.New().String()
	tokenHash := hashVerificationString(verificationString)

	// Create session in database
	now := time.Now().UTC()
	expiresAt := now.Add(300 * 24 * time.Hour) // 300 days from Fase 1.7 spec

	createSessionParams := &gen.CreateSessionParams{
		ID:                 uuid.New().String(),
		UserID:             user.ID,
		SessionID:          sessionID,
		VerificationString: verificationString,
		TokenHash:          tokenHash,
		ExpiresAt:          expiresAt,
	}

	_, err = s.sessionRepo.CreateSession(ctx, createSessionParams)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("Failed to create session in database")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Store session in cache
	cacheData := auth.SessionData{
		UserID:             user.ID,
		CompanyID:          nil, // User doesn't have a company yet after registration
		VerificationString: verificationString,
		ExpiresAt:          expiresAt,
		RevokedAt:          nil,
		CreatedAt:          now,
	}
	s.sessionCache.Store(sessionID, cacheData)

	// Generate tokens
	sessionToken, sessionExpiresAt, err := s.jwtManager.GenerateSessionToken(user.ID, sessionID, nil)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("Failed to generate session token")
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	refreshToken, _, _, err := s.jwtManager.GenerateRefreshToken(user.ID, sessionID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	log.Info().
		Str("user_id", user.ID).
		Str("username", username).
		Str("session_id", sessionID).
		Msg("User successfully logged in")

	return &LoginResponse{
		User:         dbUserToModel(user),
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    sessionExpiresAt.Unix(),
	}, nil
}

// Logout revokes the user's session.
func (s *authService) Logout(ctx context.Context, sessionID string) error {
	log.Info().Str("session_id", sessionID).Msg("Logout attempt")

	// Get session from database to verify it exists
	session, err := s.sessionRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session")
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		log.Warn().Str("session_id", sessionID).Msg("Session not found for logout")
		return fmt.Errorf("session_not_found")
	}

	// Revoke session in database
	err = s.sessionRepo.RevokeSession(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID).Msg("Failed to revoke session in database")
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Revoke session in cache
	s.sessionCache.Revoke(sessionID)

	log.Info().
		Str("session_id", sessionID).
		Str("user_id", session.UserID).
		Msg("User successfully logged out")

	return nil
}

// hashVerificationString creates SHA256 hash of verification string (for token_hash storage).
func hashVerificationString(verificationString string) string {
	hash := sha256.Sum256([]byte(verificationString))
	return hex.EncodeToString(hash[:])
}

// dbUserToModel converts a DBO user to a Model user.
func dbUserToModel(dbo *gen.User) *User {
	return &User{
		ID:                         dbo.ID,
		Username:                   dbo.Username,
		Email:                      dbo.Email,
		Role:                       dbo.Role,
		Timezone:                   dbo.Timezone,
		LastTimezoneModificationAt: dbo.LastTimezoneModificationAt,
		CreatedAt:                  dbo.CreatedAt,
	}
}
