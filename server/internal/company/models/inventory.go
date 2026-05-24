package models

import "time"

// ============================================================================
// DBO - Database Object (mapeo directo de tabla company_inventory)
// ============================================================================

// CompanyInventoryItemDBO representa el mapeo directo de la tabla company_inventory en la BD
type CompanyInventoryItemDBO struct {
	ID         string     `db:"id"`
	CompanyID  string     `db:"company_id"`
	ResourceID string     `db:"resource_id"`
	Quantity   int64      `db:"quantity"`
	IsDeleted  int        `db:"is_deleted"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

// ============================================================================
// Model - Tipo de dominio enriquecido con reglas de negocio
// ============================================================================

// CompanyInventoryItem representa un recurso en el inventario de una empresa
type CompanyInventoryItem struct {
	ID         string
	CompanyID  string
	ResourceID string
	Quantity   int64 // cantidad disponible (siempre >= 0)
	IsDeleted  bool
	DeletedAt  *time.Time
}

// NewCompanyInventoryItemFromDBO convierte un DBO a Model
func NewCompanyInventoryItemFromDBO(dbo *CompanyInventoryItemDBO) *CompanyInventoryItem {
	return &CompanyInventoryItem{
		ID:         dbo.ID,
		CompanyID:  dbo.CompanyID,
		ResourceID: dbo.ResourceID,
		Quantity:   dbo.Quantity,
		IsDeleted:  dbo.IsDeleted == 1,
		DeletedAt:  dbo.DeletedAt,
	}
}

// ToDTO convierte el Model a DTO para la API
func (i *CompanyInventoryItem) ToDTO() *CompanyInventoryItemDTO {
	return &CompanyInventoryItemDTO{
		ID:         i.ID,
		CompanyID:  i.CompanyID,
		ResourceID: i.ResourceID,
		Quantity:   i.Quantity,
	}
}

// CanRemoveQuantity valida si hay suficiente cantidad
func (i *CompanyInventoryItem) CanRemoveQuantity(amount int64) bool {
	return i.Quantity >= amount
}

// RemoveQuantity resta cantidad si es posible, retorna error si no hay suficiente
func (i *CompanyInventoryItem) RemoveQuantity(amount int64) error {
	if amount < 0 {
		return ErrInvalidQuantity
	}
	if !i.CanRemoveQuantity(amount) {
		return ErrInsufficientInventory
	}
	i.Quantity -= amount
	return nil
}

// AddQuantity suma cantidad al inventario
func (i *CompanyInventoryItem) AddQuantity(amount int64) error {
	if amount < 0 {
		return ErrInvalidQuantity
	}
	i.Quantity += amount
	return nil
}

// ============================================================================
// CompanyInventory - Contenedor del inventario completo
// ============================================================================

// CompanyInventory representa el inventario completo de una empresa
type CompanyInventory struct {
	CompanyID string
	Items     map[string]*CompanyInventoryItem // key: resource_id, value: inventory item
}

// NewCompanyInventory crea un nuevo inventario vacío
func NewCompanyInventory(companyID string) *CompanyInventory {
	return &CompanyInventory{
		CompanyID: companyID,
		Items:     make(map[string]*CompanyInventoryItem),
	}
}

// AddItem agrega o actualiza un item en el inventario
func (inv *CompanyInventory) AddItem(item *CompanyInventoryItem) {
	inv.Items[item.ResourceID] = item
}

// GetItem obtiene un item específico del inventario
func (inv *CompanyInventory) GetItem(resourceID string) (*CompanyInventoryItem, bool) {
	item, exists := inv.Items[resourceID]
	return item, exists
}

// ToDTO convierte el inventario a DTO para la API
func (inv *CompanyInventory) ToDTO() []*CompanyInventoryItemDTO {
	dto := make([]*CompanyInventoryItemDTO, 0, len(inv.Items))
	for _, item := range inv.Items {
		dto = append(dto, item.ToDTO())
	}
	return dto
}

// ============================================================================
// DTO - Data Transfer Object para la API
// ============================================================================

// CompanyInventoryItemDTO representa un recurso del inventario en la API
type CompanyInventoryItemDTO struct {
	ID         string `json:"id"`
	CompanyID  string `json:"company_id"`
	ResourceID string `json:"resource_id"`
	Quantity   int64  `json:"quantity"`
}

// ============================================================================
// Requests
// ============================================================================

// AddResourceRequest representa el payload para agregar recursos al inventario
type AddResourceRequest struct {
	ResourceID string `json:"resource_id" validate:"required,uuid4"`
	Quantity   int64  `json:"quantity" validate:"required,gt=0"`
}

// RemoveResourceRequest representa el payload para remover recursos del inventario
type RemoveResourceRequest struct {
	ResourceID string `json:"resource_id" validate:"required,uuid4"`
	Quantity   int64  `json:"quantity" validate:"required,gt=0"`
}
