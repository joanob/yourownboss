package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// InventoryRepositoryInterface defines the interface for inventory repository operations
type InventoryRepositoryInterface interface {
	GetInventory(ctx context.Context, companyID string) (*models.CompanyInventory, error)
	GetInventoryItem(ctx context.Context, companyID, resourceID string) (*models.CompanyInventoryItem, error)
	AddToInventory(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
	RemoveFromInventory(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
	UpdateInventoryQuantity(ctx context.Context, itemID string, newQuantity int64) (*models.CompanyInventoryItem, error)
	UpsertInventoryItem(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
}

// InventoryRepository handles all inventory-related database operations
type InventoryRepository struct {
	queries *dbqueries.Queries
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(queries *dbqueries.Queries) *InventoryRepository {
	return &InventoryRepository{
		queries: queries,
	}
}

// GetInventory retrieves all inventory items for a company
func (r *InventoryRepository) GetInventory(ctx context.Context, companyID string) (*models.CompanyInventory, error) {
	logger := log.With().Str("company_id", companyID).Logger()

	items, err := r.queries.GetInventoryByCompanyID(ctx, companyID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error().Err(err).Msg("Failed to get inventory")
		return nil, err
	}

	inventory := models.NewCompanyInventory(companyID)
	for _, item := range items {
		inventoryItem := &models.CompanyInventoryItem{
			ID:         item.ID,
			CompanyID:  item.CompanyID,
			ResourceID: item.ResourceID,
			Quantity:   item.Quantity,
			IsDeleted:  item.IsDeleted == 1,
			DeletedAt:  item.DeletedAt,
		}
		inventory.AddItem(inventoryItem)
	}

	return inventory, nil
}

// GetInventoryItem retrieves a specific inventory item
func (r *InventoryRepository) GetInventoryItem(ctx context.Context, companyID, resourceID string) (*models.CompanyInventoryItem, error) {
	dbo, err := r.queries.GetInventoryItem(ctx, dbqueries.GetInventoryItemParams{
		CompanyID:  companyID,
		ResourceID: resourceID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &models.CompanyInventoryItem{
		ID:         dbo.ID,
		CompanyID:  dbo.CompanyID,
		ResourceID: dbo.ResourceID,
		Quantity:   dbo.Quantity,
		IsDeleted:  dbo.IsDeleted == 1,
		DeletedAt:  dbo.DeletedAt,
	}, nil
}

// AddToInventory adds resources to the inventory using UPSERT
// If the item doesn't exist, it creates it. If it exists, it adds to the quantity.
func (r *InventoryRepository) AddToInventory(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("resource_id", resourceID).
		Int64("quantity", quantity).
		Logger()

	if quantity <= 0 {
		logger.Error().Int64("quantity", quantity).Msg("Quantity must be positive")
		return nil, models.ErrInvalidQuantity
	}

	itemID := uuid.New().String()

	dbo, err := r.queries.AddToInventory(ctx, dbqueries.AddToInventoryParams{
		ID:         itemID,
		CompanyID:  companyID,
		ResourceID: resourceID,
		Quantity:   quantity,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to add to inventory")
		return nil, err
	}

	logger.Info().Msg("Resources added to inventory successfully")

	return &models.CompanyInventoryItem{
		ID:         dbo.ID,
		CompanyID:  dbo.CompanyID,
		ResourceID: dbo.ResourceID,
		Quantity:   dbo.Quantity,
		IsDeleted:  dbo.IsDeleted == 1,
		DeletedAt:  dbo.DeletedAt,
	}, nil
}

// RemoveFromInventory removes resources from the inventory
// Validates that sufficient quantity exists
// Note: SQLite doesn't support FOR UPDATE for pessimistic locking,
// so we validate at the application level
func (r *InventoryRepository) RemoveFromInventory(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("resource_id", resourceID).
		Int64("quantity", quantity).
		Logger()

	if quantity <= 0 {
		logger.Error().Int64("quantity", quantity).Msg("Quantity must be positive")
		return nil, models.ErrInvalidQuantity
	}

	// Get item to check current quantity
	item, err := r.GetInventoryItem(ctx, companyID, resourceID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get inventory item")
		return nil, err
	}
	if item == nil {
		logger.Warn().Msg("Inventory item not found")
		return nil, models.ErrInsufficientInventory
	}

	// Check if we have enough quantity
	if item.Quantity < quantity {
		logger.Warn().
			Int64("available", item.Quantity).
			Int64("requested", quantity).
			Msg("Insufficient inventory")
		return nil, models.ErrInsufficientInventory
	}

	// Calculate new quantity
	newQuantity := item.Quantity - quantity

	// Update the inventory
	dbo, err := r.queries.RemoveFromInventory(ctx, dbqueries.RemoveFromInventoryParams{
		Quantity:   newQuantity,
		CompanyID:  companyID,
		ResourceID: resourceID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to remove from inventory")
		return nil, err
	}

	logger.Info().Int64("new_quantity", newQuantity).Msg("Resources removed from inventory successfully")

	return &models.CompanyInventoryItem{
		ID:         dbo.ID,
		CompanyID:  dbo.CompanyID,
		ResourceID: dbo.ResourceID,
		Quantity:   dbo.Quantity,
		IsDeleted:  dbo.IsDeleted == 1,
		DeletedAt:  dbo.DeletedAt,
	}, nil
}

// UpdateInventoryQuantity updates the quantity of an inventory item
// Used when we need to set an exact quantity rather than add/remove
func (r *InventoryRepository) UpdateInventoryQuantity(ctx context.Context, itemID string, newQuantity int64) (*models.CompanyInventoryItem, error) {
	logger := log.With().
		Str("item_id", itemID).
		Int64("new_quantity", newQuantity).
		Logger()

	if newQuantity < 0 {
		logger.Error().Int64("quantity", newQuantity).Msg("Quantity cannot be negative")
		return nil, models.ErrInvalidQuantity
	}

	dbo, err := r.queries.UpdateInventoryQuantity(ctx, dbqueries.UpdateInventoryQuantityParams{
		Quantity: newQuantity,
		ID:       itemID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update inventory quantity")
		return nil, err
	}

	logger.Info().Msg("Inventory quantity updated successfully")

	return &models.CompanyInventoryItem{
		ID:         dbo.ID,
		CompanyID:  dbo.CompanyID,
		ResourceID: dbo.ResourceID,
		Quantity:   dbo.Quantity,
		IsDeleted:  dbo.IsDeleted == 1,
		DeletedAt:  dbo.DeletedAt,
	}, nil
}

// UpsertInventoryItem inserts or updates an inventory item with a specific quantity
// Unlike AddToInventory, this sets the exact quantity rather than adding to it
func (r *InventoryRepository) UpsertInventoryItem(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("resource_id", resourceID).
		Int64("quantity", quantity).
		Logger()

	if quantity < 0 {
		logger.Error().Int64("quantity", quantity).Msg("Quantity cannot be negative")
		return nil, models.ErrInvalidQuantity
	}

	itemID := uuid.New().String()

	dbo, err := r.queries.UpsertInventoryItem(ctx, dbqueries.UpsertInventoryItemParams{
		ID:         itemID,
		CompanyID:  companyID,
		ResourceID: resourceID,
		Quantity:   quantity,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upsert inventory item")
		return nil, err
	}

	logger.Info().Msg("Inventory item upserted successfully")

	return &models.CompanyInventoryItem{
		ID:         dbo.ID,
		CompanyID:  dbo.CompanyID,
		ResourceID: dbo.ResourceID,
		Quantity:   dbo.Quantity,
		IsDeleted:  dbo.IsDeleted == 1,
		DeletedAt:  dbo.DeletedAt,
	}, nil
}
