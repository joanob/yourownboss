package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/joanob/yourownboss/internal/auth"
	"github.com/joanob/yourownboss/internal/db/dbqueries"
	"github.com/rs/zerolog/log"
)

// userSessionRepository implementa auth.UserSessionRepository
type userSessionRepository struct {
	queries *dbqueries.Queries
}

// NewUserSessionRepository crea un nuevo repositorio de sesiones de usuario
func NewUserSessionRepository(queries *dbqueries.Queries) auth.UserSessionRepository {
	return &userSessionRepository{
		queries: queries,
	}
}

// CreateSession inserta una nueva sesión en la BD
func (r *userSessionRepository) CreateSession(ctx context.Context, params *dbqueries.CreateSessionParams) (*dbqueries.UserSession, error) {
	logger := log.With().
		Str("session_id", params.SessionID).
		Str("user_id", params.UserID).
		Logger()

	err := r.queries.CreateSession(ctx, *params)

	if err != nil {
		logger.Error().Err(err).Msg("Error creando sesión en BD")
		return nil, fmt.Errorf("error creando sesión: %w", err)
	}

	logger.Info().Msg("Sesión creada en BD")

	// Retrieve the created session
	session, err := r.queries.GetSessionBySessionID(ctx, params.SessionID)
	if err != nil {
		logger.Error().Err(err).Msg("Error obteniendo sesión recién creada")
		return nil, fmt.Errorf("error obteniendo sesión: %w", err)
	}

	return &session, nil
}

// GetByID obtiene una sesión por su ID
func (r *userSessionRepository) GetByID(ctx context.Context, id string) (*dbqueries.UserSession, error) {
	session, err := r.queries.GetSessionByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("session_id", id).Msg("Sesión no encontrada por ID")
			return nil, nil // Sesión no existe
		}
		log.Error().Err(err).Str("session_id", id).Msg("Error obteniendo sesión por ID")
		return nil, fmt.Errorf("error obteniendo sesión: %w", err)
	}

	return &session, nil
}

// GetBySessionID obtiene una sesión por su session_id
func (r *userSessionRepository) GetBySessionID(ctx context.Context, sessionID string) (*dbqueries.UserSession, error) {
	session, err := r.queries.GetSessionBySessionID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("session_id", sessionID).Msg("Sesión no encontrada por session_id")
			return nil, nil // Sesión no existe
		}
		log.Error().Err(err).Str("session_id", sessionID).Msg("Error obteniendo sesión por session_id")
		return nil, fmt.Errorf("error obteniendo sesión: %w", err)
	}

	return &session, nil
}

// GetByTokenHash obtiene una sesión por su token_hash
func (r *userSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*dbqueries.UserSession, error) {
	session, err := r.queries.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Msg("Sesión no encontrada por token_hash")
			return nil, nil // Sesión no existe
		}
		log.Error().Err(err).Msg("Error obteniendo sesión por token_hash")
		return nil, fmt.Errorf("error obteniendo sesión: %w", err)
	}

	return &session, nil
}

// RevokeSession marca una sesión como revocada
func (r *userSessionRepository) RevokeSession(ctx context.Context, sessionID string) error {
	logger := log.With().
		Str("session_id", sessionID).
		Logger()

	err := r.queries.RevokeSession(ctx, sessionID)
	if err != nil {
		logger.Error().Err(err).Msg("Error revocando sesión")
		return fmt.Errorf("error revocando sesión: %w", err)
	}

	logger.Info().Msg("Sesión revocada")
	return nil
}

// DeleteExpiredSessions marca como borradas todas las sesiones expiradas o revocadas
func (r *userSessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	err := r.queries.DeleteExpiredSessions(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error eliminando sesiones expiradas")
		return fmt.Errorf("error eliminando sesiones expiradas: %w", err)
	}

	log.Debug().Msg("Sesiones expiradas eliminadas")
	return nil
}

// CountActiveSessions cuenta el número de sesiones activas de un usuario
func (r *userSessionRepository) CountActiveSessions(ctx context.Context, userID string) (int64, error) {
	count, err := r.queries.CountActiveSessionsByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Error contando sesiones activas")
		return 0, fmt.Errorf("error contando sesiones: %w", err)
	}

	return count, nil
}

// GetUserSessions obtiene todas las sesiones de un usuario
func (r *userSessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*dbqueries.UserSession, error) {
	sessions, err := r.queries.GetUserSessionsByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Error obteniendo sesiones del usuario")
		return nil, fmt.Errorf("error obteniendo sesiones: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*dbqueries.UserSession, len(sessions))
	for i := range sessions {
		result[i] = &sessions[i]
	}

	return result, nil
}
