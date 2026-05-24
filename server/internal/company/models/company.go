package models

import "time"

// ============================================================================
// DBO - Database Object (mapeo directo de tabla companies)
// ============================================================================

// CompanyDBO representa el mapeo directo de la tabla companies en la BD
type CompanyDBO struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	Name      string     `db:"name"`
	Money     int64      `db:"money"`
	CreatedAt time.Time  `db:"created_at"`
	IsDeleted int        `db:"is_deleted"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// ============================================================================
// Model - Tipo de dominio enriquecido con reglas de negocio
// ============================================================================

// Company representa una empresa del jugador con su lógica de negocio
type Company struct {
	ID        string
	UserID    string
	Name      string
	Money     int64 // dinero de la empresa (siempre >= 0)
	CreatedAt time.Time
	IsDeleted bool
	DeletedAt *time.Time
}

// NewCompanyFromDBO convierte un DBO a Model
func NewCompanyFromDBO(dbo *CompanyDBO) *Company {
	return &Company{
		ID:        dbo.ID,
		UserID:    dbo.UserID,
		Name:      dbo.Name,
		Money:     dbo.Money,
		CreatedAt: dbo.CreatedAt,
		IsDeleted: dbo.IsDeleted == 1,
		DeletedAt: dbo.DeletedAt,
	}
}

// ToDTO convierte el Model a DTO para la API
func (c *Company) ToDTO() *CompanyDTO {
	return &CompanyDTO{
		ID:        c.ID,
		UserID:    c.UserID,
		Name:      c.Name,
		Money:     c.Money,
		CreatedAt: c.CreatedAt,
	}
}

// CanAfford valida si la empresa tiene suficiente dinero
func (c *Company) CanAfford(amount int64) bool {
	return c.Money >= amount
}

// CanRemoveMoney resta dinero si es posible, retorna error si no hay suficiente
func (c *Company) RemoveMoney(amount int64) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	if !c.CanAfford(amount) {
		return ErrInsufficientFunds
	}
	c.Money -= amount
	return nil
}

// AddMoney suma dinero a la empresa
func (c *Company) AddMoney(amount int64) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	c.Money += amount
	return nil
}

// ============================================================================
// DTO - Data Transfer Object para la API
// ============================================================================

// CompanyDTO representa los datos de una empresa en la API
type CompanyDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Money     int64     `json:"money"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// CreateCompanyRequest
// ============================================================================

// CreateCompanyRequest representa el payload de POST /api/v1/company
type CreateCompanyRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

// ============================================================================
// UpdateCompanyRequest
// ============================================================================

// UpdateCompanyRequest representa el payload de PUT /api/v1/company
type UpdateCompanyRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}
