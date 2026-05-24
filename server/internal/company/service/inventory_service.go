package service

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/company/repository"
)

// InventoryService defines operations for inventory management
type InventoryService interface {
	GetInventory(ctx context.Context, companyID string) (*models.CompanyInventory, error)
	AddResource(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error)
	RemoveResource(ctx context.Context, companyID, resourceID string, quantity int64, validateBefore bool) (*models.CompanyInventoryItem, error)
	TransferResource(ctx context.Context, fromCompanyID, toCompanyID, resourceID string, quantity int64) error
}

// inventoryService implements InventoryService
type inventoryService struct {
	inventoryRepo *repository.InventoryRepository
	companyRepo   *repository.CompanyRepository
}

// NewInventoryService creates a new inventory service
func NewInventoryService(inventoryRepo *repository.InventoryRepository, companyRepo *repository.CompanyRepository) InventoryService {
	return &inventoryService{
		inventoryRepo: inventoryRepo,
		companyRepo:   companyRepo,
	}
}

// GetInventory retrieves all inventory items for a company
func (s *inventoryService) GetInventory(ctx context.Context, companyID string) (*models.CompanyInventory, error) {
	logger := log.With().Str("company_id", companyID).Logger()

	// Verify company exists
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to verify company exists")
		return nil, err
	}

	if company == nil {
		logger.Warn().Msg("Company not found")
		return nil, models.ErrInsufficientFunds
	}

	// Get inventory
	inventory, err := s.inventoryRepo.GetInventory(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get inventory")
		return nil, err
	}

	logger.Debug().Int("item_count", len(*inventory)).Msg("Inventory retrieved successfully")
	return inventory, nil
}

// AddResource adds resources to a company's inventory
func (s *inventoryService) AddResource(ctx context.Context, companyID, resourceID string, quantity int64) (*models.CompanyInventoryItem, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("resource_id", resourceID).
		Int64("quantity", quantity).
		Logger()

	// Validate inputs
	if quantity <= 0 {
		logger.Warn().Int64("quantity", quantity).Msg("Quantity must be positive")
		return nil, models.ErrInvalidQuantity
	}

	// Verify company exists
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to verify company exists")
		return nil, err
	}

	if company == nil {
		logger.Warn().Msg("Company not found")
		return nil, models.ErrInsufficientFunds
	}

	// Add to inventory (UPSERT)
	item, err := s.inventoryRepo.AddToInventory(ctx, companyID, resourceID, quantity)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to add resource to inventory")
		return nil, err
	}

	logger.Info().Int64("new_quantity", item.Quantity).Msg("Resource added to inventory successfully")
	return item, nil
}

// RemoveResource removes resources from a company's inventory
// If validateBefore is true, validates availability before removal
func (s *inventoryService) RemoveResource(ctx context.Context, companyID, resourceID string, quantity int64, validateBefore bool) (*models.CompanyInventoryItem, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("resource_id", resourceID).
		Int64("quantity", quantity).
		Bool("validate_before", validateBefore).
		Logger()

	// Validate inputs
	if quantity <= 0 {
		logger.Warn().Int64("quantity", quantity).Msg("Quantity must be positive")
		return nil, models.ErrInvalidQuantity
	}

	// Verify company exists
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to verify company exists")
		return nil, err
	}

	if company == nil {
		logger.Warn().Msg("Company not found")
		return nil, models.ErrInsufficientFunds
	}

	// If validateBefore is true, check inventory first
	if validateBefore {
		item, err := s.inventoryRepo.GetInventoryItem(ctx, companyID, resourceID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get inventory item")
			return nil, err
		}

		if item == nil || item.Quantity < quantity {
			logger.Warn().
				Int64("available", func() int64 {
					if item == nil {
						return 0
					}
					return item.Quantity
				}()).
				Msg("Insufficient inventory")
			return nil, models.ErrInsufficientInventory
		}
	}

	// Remove from inventory (RemoveFromInventory handles validation internally)
	item, err := s.inventoryRepo.RemoveFromInventory(ctx, companyID, resourceID, quantity)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to remove resource from inventory")
		return nil, err
	}

	logger.Info().Int64("new_quantity", item.Quantity).Msg("Resource removed from inventory successfully")
	return item, nil
}

// TransferResource transfers resources between two companies
// This is a compound operation that should ideally use a transaction
// For now, we perform both operations separately with error handling
func (s *inventoryService) TransferResource(ctx context.Context, fromCompanyID, toCompanyID, resourceID string, quantity int64) error {
	logger := log.With().
		Str("from_company_id", fromCompanyID).
		Str("to_company_id", toCompanyID).
		Str("resource_id", resourceID).
		Int64("quantity", quantity).
		Logger()

	// Validate inputs
	if quantity <= 0 {
		logger.Warn().Int64("quantity", quantity).Msg("Quantity must be positive")
		return models.ErrInvalidQuantity
	}

	// Verify both companies exist
	fromCompany, err := s.companyRepo.GetCompanyByID(ctx, fromCompanyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to verify source company")
		return err
	}

	if fromCompany == nil {
		logger.Warn().Msg("Source company not found")
		return models.ErrInsufficientFunds
	}

	toCompany, err := s.companyRepo.GetCompanyByID(ctx, toCompanyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to verify destination company")
		return err
	}

	if toCompany == nil {
		logger.Warn().Msg("Destination company not found")
		return models.ErrInsufficientFunds
	}

	// Remove from source company
	_, err = s.inventoryRepo.RemoveFromInventory(ctx, fromCompanyID, resourceID, quantity)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to remove resource from source company")
		return err
	}

	// Add to destination company
	_, err = s.inventoryRepo.AddToInventory(ctx, toCompanyID, resourceID, quantity)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to add resource to destination company")
		// TODO: In production, this should rollback the removal from source company
		// This requires database transaction support
		return err
	}

	logger.Info().Msg("Resource transferred successfully")
	return nil
}
