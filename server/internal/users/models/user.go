package models

import (
	"time"
)

// Role representa el rol de un usuario
type Role string

const (
	RolePlayer Role = "P" // Player (usuario normal)
	RoleAdmin  Role = "A" // Admin (administrador)
)

// IsValid valida si el rol es válido
func (r Role) IsValid() bool {
	return r == RolePlayer || r == RoleAdmin
}

// IsAdmin devuelve true si el rol es admin
func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}

// ============================================================================
// DBO (Database Object) - Estructura mapeada directamente a tabla BD
// ============================================================================

// UserDBO estructura exactamente mapeada a tabla users
// Generada por sqlc, NO exponer directamente a la API
type UserDBO struct {
	ID                         string     `db:"id"`
	Username                   string     `db:"username"`
	Email                      string     `db:"email"`
	PasswordHash               string     `db:"password_hash"`
	Role                       Role       `db:"role"`
	Timezone                   *string    `db:"timezone"`                      // nullable
	LastTimezoneModificationAt *time.Time `db:"last_timezone_modification_at"` // nullable
	CreatedAt                  time.Time  `db:"created_at"`
	IsDeleted                  int        `db:"is_deleted"`
	DeletedAt                  *time.Time `db:"deleted_at"` // nullable
}

// ============================================================================
// Model - Estructura interna con reglas de negocio
// ============================================================================

// User modelo interno con validaciones y métodos de negocio
type User struct {
	ID                         string
	Username                   string
	Email                      string
	PasswordHash               string
	Role                       Role
	Timezone                   *string
	LastTimezoneModificationAt *time.Time
	CreatedAt                  time.Time
	IsDeleted                  bool
	DeletedAt                  *time.Time
}

// FromDBO convierte un UserDBO a User
func (u *User) FromDBO(dbo *UserDBO) {
	u.ID = dbo.ID
	u.Username = dbo.Username
	u.Email = dbo.Email
	u.PasswordHash = dbo.PasswordHash
	u.Role = dbo.Role
	u.Timezone = dbo.Timezone
	u.LastTimezoneModificationAt = dbo.LastTimezoneModificationAt
	u.CreatedAt = dbo.CreatedAt
	u.IsDeleted = dbo.IsDeleted == 1
	u.DeletedAt = dbo.DeletedAt
}

// ToDBO convierte un User a UserDBO
func (u *User) ToDBO() *UserDBO {
	isDeleted := 0
	if u.IsDeleted {
		isDeleted = 1
	}
	return &UserDBO{
		ID:                         u.ID,
		Username:                   u.Username,
		Email:                      u.Email,
		PasswordHash:               u.PasswordHash,
		Role:                       u.Role,
		Timezone:                   u.Timezone,
		LastTimezoneModificationAt: u.LastTimezoneModificationAt,
		CreatedAt:                  u.CreatedAt,
		IsDeleted:                  isDeleted,
		DeletedAt:                  u.DeletedAt,
	}
}

// CanChangeTimezone verifica si el usuario puede cambiar timezone
// Puede cambiar si: nunca cambió OR pasaron más de 30 días desde el último cambio
func (u *User) CanChangeTimezone() bool {
	if u.LastTimezoneModificationAt == nil {
		return true // Nunca cambió, puede cambiar
	}
	// Puede cambiar si pasaron 30 días
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return u.LastTimezoneModificationAt.Before(thirtyDaysAgo)
}

// TimezoneChangeBlockedUntil devuelve la fecha hasta la cual está bloqueado el cambio de timezone
// Devuelve nil si puede cambiar
func (u *User) TimezoneChangeBlockedUntil() *time.Time {
	if u.CanChangeTimezone() {
		return nil
	}
	blockedUntil := u.LastTimezoneModificationAt.AddDate(0, 0, 30)
	return &blockedUntil
}

// ============================================================================
// DTO (Data Transfer Object) - Para respuestas API
// ============================================================================

// UserDTO estructura para respuestas API (GET /api/v1/users/me)
type UserDTO struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	Timezone  *string `json:"timezone"`
	CreatedAt string  `json:"created_at"` // ISO 8601
}

// ToDTO convierte un User a UserDTO
func (u *User) ToDTO() *UserDTO {
	return &UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      string(u.Role),
		Timezone:  u.Timezone,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

// ============================================================================
// Request DTOs - Para requests HTTP
// ============================================================================

// RegisterRequest estructura para POST /api/v1/auth/register
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Timezone string `json:"timezone" validate:"required,timezone"` // ej: "Europe/Madrid"
}

// LoginRequest estructura para POST /api/v1/auth/login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserRequest estructura para PUT /api/v1/users/me
type UpdateUserRequest struct {
	Timezone *string `json:"timezone" validate:"omitempty,timezone"` // opcional
}

// ============================================================================
// Response DTOs - Para respuestas HTTP
// ============================================================================

// AuthResponse estructura para respuestas de login/register
// Las cookies se envían en headers, no en body
type AuthResponse struct {
	User             *UserDTO `json:"user"`
	SessionExpiresAt string   `json:"session_expires_at"` // ISO 8601
	RefreshExpiresAt string   `json:"refresh_expires_at"` // ISO 8601
}

// ============================================================================
// Session response
// ============================================================================

// SessionResponse estructura para respuestas de sesión
type SessionResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
