package models

import (
	"time"
)

// ============================================================================
// DBO (Database Object) - Estructura mapeada a tabla user_sessions
// ============================================================================

// UserSessionDBO estructura exactamente mapeada a tabla user_sessions
// Generada por sqlc, NO exponer directamente a la API
type UserSessionDBO struct {
	ID                 string     `db:"id"`
	UserID             string     `db:"user_id"`
	SessionID          string     `db:"session_id"`
	VerificationString string     `db:"verification_string"`
	TokenHash          string     `db:"token_hash"`
	ExpiresAt          time.Time  `db:"expires_at"`
	RevokedAt          *time.Time `db:"revoked_at"` // nullable
	CreatedAt          time.Time  `db:"created_at"`
	IsDeleted          int        `db:"is_deleted"`
	DeletedAt          *time.Time `db:"deleted_at"` // nullable
}

// ============================================================================
// Model - Estructura interna
// ============================================================================

// UserSession modelo interno para sesiones de usuario
type UserSession struct {
	ID                 string
	UserID             string
	SessionID          string
	VerificationString string
	TokenHash          string
	ExpiresAt          time.Time
	RevokedAt          *time.Time
	CreatedAt          time.Time
	IsDeleted          bool
	DeletedAt          *time.Time
}

// FromDBO convierte un UserSessionDBO a UserSession
func (s *UserSession) FromDBO(dbo *UserSessionDBO) {
	s.ID = dbo.ID
	s.UserID = dbo.UserID
	s.SessionID = dbo.SessionID
	s.VerificationString = dbo.VerificationString
	s.TokenHash = dbo.TokenHash
	s.ExpiresAt = dbo.ExpiresAt
	s.RevokedAt = dbo.RevokedAt
	s.CreatedAt = dbo.CreatedAt
	s.IsDeleted = dbo.IsDeleted == 1
	s.DeletedAt = dbo.DeletedAt
}

// ToDBO convierte un UserSession a UserSessionDBO
func (s *UserSession) ToDBO() *UserSessionDBO {
	isDeleted := 0
	if s.IsDeleted {
		isDeleted = 1
	}
	return &UserSessionDBO{
		ID:                 s.ID,
		UserID:             s.UserID,
		SessionID:          s.SessionID,
		VerificationString: s.VerificationString,
		TokenHash:          s.TokenHash,
		ExpiresAt:          s.ExpiresAt,
		RevokedAt:          s.RevokedAt,
		CreatedAt:          s.CreatedAt,
		IsDeleted:          isDeleted,
		DeletedAt:          s.DeletedAt,
	}
}

// IsValid valida si la sesión está activa
func (s *UserSession) IsValid() bool {
	// Sesión válida si no está revocada y no ha expirado
	return s.RevokedAt == nil && time.Now().Before(s.ExpiresAt) && !s.IsDeleted
}

// IsRevoked verifica si la sesión fue revocada
func (s *UserSession) IsRevoked() bool {
	return s.RevokedAt != nil
}

// IsExpired verifica si la sesión ha expirado
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// ============================================================================
// DTO - Para respuestas API (si es necesario exponer sesiones)
// ============================================================================

// UserSessionDTO estructura para respuestas API sobre sesiones
type UserSessionDTO struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"` // ISO 8601
	ExpiresAt string `json:"expires_at"` // ISO 8601
	IsActive  bool   `json:"is_active"`
}

// ToDTO convierte un UserSession a UserSessionDTO
func (s *UserSession) ToDTO() *UserSessionDTO {
	return &UserSessionDTO{
		SessionID: s.SessionID,
		UserID:    s.UserID,
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
		ExpiresAt: s.ExpiresAt.Format(time.RFC3339),
		IsActive:  s.IsValid(),
	}
}
