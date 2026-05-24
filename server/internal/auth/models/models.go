package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionTokenClaims estructura para el JWT de sesión corta (1 minuto)
// Se valida sin acceso a BD (solo firma y expiry)
type SessionTokenClaims struct {
	UserID    string  `json:"user_id"`
	SessionID string  `json:"session_id"`
	CompanyID *string `json:"company_id"` // nullable, null si usuario no tiene empresa
	ExpiresAt int64   `json:"exp"`        // Unix timestamp
	IssuedAt  int64   `json:"iat"`        // Unix timestamp
}

// Implementar interfaz jwt.Claims para SessionTokenClaims
func (c SessionTokenClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

func (c SessionTokenClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

func (c SessionTokenClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c SessionTokenClaims) GetIssuer() (string, error) {
	return "", nil
}

func (c SessionTokenClaims) GetSubject() (string, error) {
	return c.UserID, nil
}

func (c SessionTokenClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}

// RefreshTokenClaims estructura para el JWT de refresco (300 días)
// Se almacena como hash en BD y se valida buscándolo en cache o BD
type RefreshTokenClaims struct {
	UserID             string `json:"user_id"`
	SessionID          string `json:"session_id"`
	VerificationString string `json:"verification_string"` // hash para validar en cache/BD
	ExpiresAt          int64  `json:"exp"`
	IssuedAt           int64  `json:"iat"`
}

// Implementar interfaz jwt.Claims para RefreshTokenClaims
func (c RefreshTokenClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

func (c RefreshTokenClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

func (c RefreshTokenClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c RefreshTokenClaims) GetIssuer() (string, error) {
	return "", nil
}

func (c RefreshTokenClaims) GetSubject() (string, error) {
	return c.UserID, nil
}

func (c RefreshTokenClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}

// SessionData estructura para almacenar sesiones en cache
// Contiene información de validación de refresh token
type SessionData struct {
	SessionID          string
	UserID             string
	CompanyID          *string    // nullable
	VerificationString string     // para validar refresh token
	TokenHash          string     // hash del refresh token para comparación
	ExpiresAt          time.Time  // cuándo expira la sesión
	RevokedAt          *time.Time // cuándo fue revocada (null si está activa)
	CreatedAt          time.Time
}

// IsValid valida si la sesión está activa
func (s *SessionData) IsValid() bool {
	// Sesión válida si no está revocada y no ha expirado
	return s.RevokedAt == nil && time.Now().Before(s.ExpiresAt)
}

// IsRevoked verifica si la sesión fue revocada
func (s *SessionData) IsRevoked() bool {
	return s.RevokedAt != nil
}

// IsExpired verifica si la sesión ha expirado
func (s *SessionData) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
