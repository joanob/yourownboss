package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joanob/yourownboss/internal/db/gen"
	"github.com/joanob/yourownboss/internal/users/repository"
	"github.com/rs/zerolog/log"
)

// PasswordManager interface for dependency injection (allows mocking in tests)
type PasswordManager interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

// UserService defines the business logic for user operations.
type UserService interface {
	// Register creates a new user with the provided username, email, and password.
	// Returns the created user or an error if user already exists, validation fails, or DB operation fails.
	Register(ctx context.Context, username, email, password, timezone string) (*User, error)

	// GetUser retrieves a user by ID.
	// Returns the user or nil if not found, or an error if DB operation fails.
	GetUser(ctx context.Context, userID string) (*User, error)

	// UpdateUser updates user information (timezone, preferences).
	// For timezone: validates 30-day restriction before last timezone change.
	// Returns updated user or error if validation/update fails.
	UpdateUser(ctx context.Context, userID string, updates *UserUpdateRequest) (*User, error)

	// DeleteUser soft-deletes a user (and cascades to related data).
	// Returns error if DB operation fails.
	DeleteUser(ctx context.Context, userID string) error
}

// User represents a user with all business logic applied.
type User struct {
	ID                         string     `json:"id"`
	Username                   string     `json:"username"`
	Email                      string     `json:"email"`
	Role                       string     `json:"role"`
	Timezone                   *string    `json:"timezone"`
	LastTimezoneModificationAt *time.Time `json:"last_timezone_modification_at"`
	CreatedAt                  time.Time  `json:"created_at"`
}

// UserUpdateRequest contains optional user updates.
type UserUpdateRequest struct {
	Timezone *string
}

// userService implements UserService.
type userService struct {
	userRepo        repository.UserRepository
	passwordManager PasswordManager
}

// NewUserService creates a new user service with dependency injection.
func NewUserService(
	userRepo repository.UserRepository,
	passwordManager PasswordManager,
) UserService {
	return &userService{
		userRepo:        userRepo,
		passwordManager: passwordManager,
	}
}

// Register creates a new user after validation.
// Validates username/email uniqueness, hashes password, persists to DB.
func (s *userService) Register(ctx context.Context, username, email, password, timezone string) (*User, error) {
	log.Info().
		Str("username", username).
		Str("email", email).
		Str("timezone", timezone).
		Msg("Registering new user")

	// Validate username uniqueness
	existsByUsername, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to check username uniqueness")
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existsByUsername {
		log.Warn().Str("username", username).Msg("Username already exists")
		return nil, fmt.Errorf("username_already_exists")
	}

	// Validate email uniqueness
	existsByEmail, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("Failed to check email uniqueness")
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existsByEmail {
		log.Warn().Str("email", email).Msg("Email already exists")
		return nil, fmt.Errorf("email_already_exists")
	}

	// Hash password with argon2id
	hashedPassword, err := s.passwordManager.HashPassword(password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	userID := uuid.New().String()

	timezonePtr := timezone
	createdUser, err := s.userRepo.CreateUser(ctx, &gen.CreateUserParams{
		ID:           userID,
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "P", // Default role is Player (not Admin)
		Timezone:     &timezonePtr,
	})
	if err != nil {
		log.Error().Err(err).Str("userID", userID).Msg("Failed to create user in database")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Info().
		Str("userID", userID).
		Str("username", username).
		Msg("User successfully registered")

	return dbUserToModel(createdUser), nil
}

// GetUser retrieves a user by ID.
// Returns nil if user not found, error if DB fails.
func (s *userService) GetUser(ctx context.Context, userID string) (*User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userID).Msg("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		log.Debug().Str("userID", userID).Msg("User not found")
		return nil, nil
	}

	return dbUserToModel(user), nil
}

// UpdateUser updates user information (timezone, etc.).
// Validates 30-day restriction for timezone changes.
// Returns updated user or error.
func (s *userService) UpdateUser(ctx context.Context, userID string, updates *UserUpdateRequest) (*User, error) {
	log.Info().
		Str("userID", userID).
		Interface("updates", updates).
		Msg("Updating user")

	// Get current user
	currentUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userID).Msg("Failed to get current user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if currentUser == nil {
		log.Warn().Str("userID", userID).Msg("User not found for update")
		return nil, fmt.Errorf("user_not_found")
	}

	// If timezone update requested, validate 30-day restriction
	if updates.Timezone != nil && *updates.Timezone != "" {
		if currentUser.Timezone != nil && *currentUser.Timezone != "" {
			// User has a timezone set, check if 30 days have passed
			if currentUser.LastTimezoneModificationAt != nil {
				daysSinceChange := time.Since(*currentUser.LastTimezoneModificationAt).Hours() / 24
				if daysSinceChange < 30 {
					log.Warn().
						Str("userID", userID).
						Float64("daysSinceChange", daysSinceChange).
						Msg("Timezone change attempted too soon (less than 30 days)")
					return nil, fmt.Errorf("timezone_recently_changed")
				}
			}
		}
	}
	// Prepare update parameters
	now := time.Now().UTC()
	updateParams := gen.UpdateUserParams{
		ID:                         userID,
		Timezone:                   updates.Timezone,
		LastTimezoneModificationAt: &now,
	}

	// Update user
	updatedUser, err := s.userRepo.UpdateUser(ctx, &updateParams)
	if err != nil {
		log.Error().Err(err).Str("userID", userID).Msg("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	if updatedUser == nil {
		log.Warn().Str("userID", userID).Msg("User not found after update")
		return nil, fmt.Errorf("user_not_found")
	}

	log.Info().Str("userID", userID).Msg("User successfully updated")

	return dbUserToModel(updatedUser), nil
}

// DeleteUser soft-deletes a user and cascades to related data.
// Returns error if operation fails.
func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	log.Info().Str("userID", userID).Msg("Deleting user")

	err := s.userRepo.SoftDeleteUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userID).Msg("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Info().Str("userID", userID).Msg("User successfully deleted")

	return nil
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
