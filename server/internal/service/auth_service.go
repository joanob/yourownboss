package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"yourownboss/internal/auth"
	"yourownboss/internal/db"
	"yourownboss/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserAlreadyExists  = errors.New("username already exists")
	ErrWeakPassword       = errors.New("password must be at least 4 characters")
)

// AuthService handles authentication business logic
type AuthService interface {
	Register(ctx context.Context, username, password string) (*AuthResult, error)
	Login(ctx context.Context, username, password string) (*AuthResult, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserByID(ctx context.Context, userID int64) (*db.User, error)
}

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository) AuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

// AuthResult contains the result of authentication operations
type AuthResult struct {
	User         *db.User
	AccessToken  string
	RefreshToken string
}

func (s *authService) Register(ctx context.Context, username, password string) (*AuthResult, error) {
	// Validate password strength
	if len(password) < 4 {
		return nil, ErrWeakPassword
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.userRepo.Create(ctx, username, string(hash))
	if err != nil {
		if err == db.ErrUserAlreadyExists {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	// Generate tokens
	return s.generateTokens(ctx, user)
}

func (s *authService) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	// Get user
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if err == db.ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	return s.generateTokens(ctx, user)
}

func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// Validate refresh token
	userID, err := s.tokenRepo.Validate(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	// Generate new access token
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.tokenRepo.Revoke(ctx, refreshToken)
}

func (s *authService) GetUserByID(ctx context.Context, userID int64) (*db.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *authService) generateTokens(ctx context.Context, user *db.User) (*AuthResult, error) {
	// Generate token pair
	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// Save refresh token to database
	if err := s.tokenRepo.Save(ctx, user.ID, tokens.RefreshToken, auth.GetRefreshTokenExpiry()); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
