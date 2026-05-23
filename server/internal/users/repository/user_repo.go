package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/joanob/yourownboss/internal/db/gen"
	"github.com/rs/zerolog/log"
)

// UserRepository define la interfaz para operaciones de usuario en la BD
type UserRepository interface {
	CreateUser(ctx context.Context, id, username, email, passwordHash, role string, timezone *string) error
	GetByID(ctx context.Context, id string) (*gen.User, error)
	GetByUsername(ctx context.Context, username string) (*gen.User, error)
	GetByEmail(ctx context.Context, email string) (*gen.User, error)
	UpdateUser(ctx context.Context, id string, timezone *string, lastTimezoneModAt *time.Time) error
	SoftDeleteUser(ctx context.Context, id string) error
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// userRepository implementa UserRepository
type userRepository struct {
	queries *gen.Queries
}

// NewUserRepository crea un nuevo repositorio de usuarios
func NewUserRepository(queries *gen.Queries) UserRepository {
	return &userRepository{
		queries: queries,
	}
}

// CreateUser inserta un nuevo usuario en la BD
func (r *userRepository) CreateUser(ctx context.Context, id, username, email, passwordHash, role string, timezone *string) error {
	logger := log.With().
		Str("user_id", id).
		Str("username", username).
		Str("email", email).
		Logger()

	err := r.queries.CreateUser(ctx, gen.CreateUserParams{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Timezone:     timezone,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Error creando usuario en BD")
		return fmt.Errorf("error creando usuario: %w", err)
	}

	logger.Info().Msg("Usuario creado en BD")
	return nil
}

// GetByID obtiene un usuario por su ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*gen.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("user_id", id).Msg("Usuario no encontrado por ID")
			return nil, nil // Usuario no existe
		}
		log.Error().Err(err).Str("user_id", id).Msg("Error obteniendo usuario por ID")
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}

	return &user, nil
}

// GetByUsername obtiene un usuario por su username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*gen.User, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("username", username).Msg("Usuario no encontrado por username")
			return nil, nil // Usuario no existe
		}
		log.Error().Err(err).Str("username", username).Msg("Error obteniendo usuario por username")
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}

	return &user, nil
}

// GetByEmail obtiene un usuario por su email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*gen.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("email", email).Msg("Usuario no encontrado por email")
			return nil, nil // Usuario no existe
		}
		log.Error().Err(err).Str("email", email).Msg("Error obteniendo usuario por email")
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}

	return &user, nil
}

// UpdateUser actualiza los datos del usuario (timezone)
func (r *userRepository) UpdateUser(ctx context.Context, id string, timezone *string, lastTimezoneModAt *time.Time) error {
	logger := log.With().
		Str("user_id", id).
		Logger()

	err := r.queries.UpdateUser(ctx, gen.UpdateUserParams{
		Timezone:                   timezone,
		LastTimezoneModificationAt: lastTimezoneModAt,
		ID:                         id,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Error actualizando usuario")
		return fmt.Errorf("error actualizando usuario: %w", err)
	}

	logger.Info().Msg("Usuario actualizado")
	return nil
}

// SoftDeleteUser marca un usuario como borrado (soft delete)
func (r *userRepository) SoftDeleteUser(ctx context.Context, id string) error {
	logger := log.With().
		Str("user_id", id).
		Logger()

	err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Error borrando usuario (soft delete)")
		return fmt.Errorf("error borrando usuario: %w", err)
	}

	logger.Info().Msg("Usuario borrado (soft delete)")
	return nil
}

// ExistsByUsername verifica si un username ya existe
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	count, err := r.queries.CountUserByUsername(ctx, username)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Error verificando existencia de username")
		return false, fmt.Errorf("error verificando username: %w", err)
	}

	return count > 0, nil
}

// ExistsByEmail verifica si un email ya existe
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.queries.CountUserByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("Error verificando existencia de email")
		return false, fmt.Errorf("error verificando email: %w", err)
	}

	return count > 0, nil
}
