package service

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	auditRepository "github.com/joanob/yourownboss/internal/audit/repository"
	companyModels "github.com/joanob/yourownboss/internal/company/models"
	companyrepository "github.com/joanob/yourownboss/internal/company/repository"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

var (
	ErrResourceNotFound        = errors.New("resource_not_found")
	ErrInvalidQuantityMultiple = errors.New("invalid_quantity_multiple")
)

// MarketTransactionResult holds the result of a buy or sell operation
type MarketTransactionResult struct {
	Company   *companyModels.Company
	Inventory *companyModels.CompanyInventory
}

// MarketService handles market buy/sell operations
type MarketService struct {
	companyRepo   companyrepository.CompanyRepositoryInterface
	inventoryRepo companyrepository.InventoryRepositoryInterface
	gamedataCache *cache.GamedataCache
	auditRepo     auditRepository.AuditRepositoryInterface
}

// NewMarketService creates a new MarketService
func NewMarketService(
	companyRepo companyrepository.CompanyRepositoryInterface,
	inventoryRepo companyrepository.InventoryRepositoryInterface,
	gamedataCache *cache.GamedataCache,
	auditRepo auditRepository.AuditRepositoryInterface,
) *MarketService {
	return &MarketService{
		companyRepo:   companyRepo,
		inventoryRepo: inventoryRepo,
		gamedataCache: gamedataCache,
		auditRepo:     auditRepo,
	}
}

// BuyResource purchases a resource from the market for the given company.
// quantity must be a positive multiple of resource.MarketSaleQty.
// Total cost = resource.MarketPrice * (quantity / resource.MarketSaleQty).
func (s *MarketService) BuyResource(ctx context.Context, companyID, resourceMasterID string, quantity int64) (*MarketTransactionResult, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("resource_master_id", resourceMasterID).
		Int64("quantity", quantity).
		Logger()

	// Lookup resource in cache
	resource, found := s.gamedataCache.GetResource(resourceMasterID)
	if !found {
		logger.Warn().Msg("Resource not found in cache")
		return nil, ErrResourceNotFound
	}

	// Validate quantity
	if quantity <= 0 {
		return nil, companyModels.ErrInvalidQuantity
	}
	if resource.MarketSaleQty <= 0 || quantity%resource.MarketSaleQty != 0 {
		logger.Warn().
			Int64("quantity", quantity).
			Int64("market_sale_qty", resource.MarketSaleQty).
			Msg("Quantity is not a valid multiple of market_sale_qty")
		return nil, ErrInvalidQuantityMultiple
	}

	totalCost := resource.MarketPrice * (quantity / resource.MarketSaleQty)

	// Get company
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get company")
		return nil, err
	}
	if company == nil {
		return nil, companyModels.ErrCompanyNotFound
	}

	// Check funds
	if !company.CanAfford(totalCost) {
		logger.Warn().
			Int64("money", company.Money).
			Int64("total_cost", totalCost).
			Msg("Insufficient funds")
		return nil, companyModels.ErrInsufficientFunds
	}

	// Deduct money
	newMoney := company.Money - totalCost
	updatedCompany, err := s.companyRepo.UpdateCompanyMoney(ctx, companyID, newMoney)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update company money")
		return nil, err
	}

	// Add to inventory
	if _, err = s.inventoryRepo.AddToInventory(ctx, companyID, resource.ID, quantity); err != nil {
		logger.Error().Err(err).Msg("Failed to add resource to inventory")
		return nil, err
	}

	// Return full inventory
	inventory, err := s.inventoryRepo.GetInventory(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve inventory after purchase")
		return nil, err
	}

	logger.Info().
		Int64("total_cost", totalCost).
		Int64("new_money", updatedCompany.Money).
		Msg("Resource purchased successfully")

	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "BUY_RESOURCE", "resource", resourceMasterID, map[string]interface{}{
			"quantity":    quantity,
			"total_cost":  totalCost,
			"after_money": updatedCompany.Money,
		})
	}

	return &MarketTransactionResult{
		Company:   updatedCompany,
		Inventory: inventory,
	}, nil
}

// SellResource sells a resource to the market for the given company.
// quantity must be a positive multiple of resource.MarketSaleQty.
// Revenue = resource.MarketPrice * (quantity / resource.MarketSaleQty).
func (s *MarketService) SellResource(ctx context.Context, companyID, resourceMasterID string, quantity int64) (*MarketTransactionResult, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("resource_master_id", resourceMasterID).
		Int64("quantity", quantity).
		Logger()

	// Lookup resource in cache
	resource, found := s.gamedataCache.GetResource(resourceMasterID)
	if !found {
		logger.Warn().Msg("Resource not found in cache")
		return nil, ErrResourceNotFound
	}

	// Validate quantity
	if quantity <= 0 {
		return nil, companyModels.ErrInvalidQuantity
	}
	if resource.MarketSaleQty <= 0 || quantity%resource.MarketSaleQty != 0 {
		logger.Warn().
			Int64("quantity", quantity).
			Int64("market_sale_qty", resource.MarketSaleQty).
			Msg("Quantity is not a valid multiple of market_sale_qty")
		return nil, ErrInvalidQuantityMultiple
	}

	// Check inventory
	inventoryItem, err := s.inventoryRepo.GetInventoryItem(ctx, companyID, resource.ID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get inventory item")
		return nil, err
	}
	available := int64(0)
	if inventoryItem != nil {
		available = inventoryItem.Quantity
	}
	if available < quantity {
		logger.Warn().
			Int64("available", available).
			Int64("requested", quantity).
			Msg("Insufficient inventory")
		return nil, companyModels.ErrInsufficientInventory
	}

	revenue := resource.MarketPrice * (quantity / resource.MarketSaleQty)

	// Remove from inventory
	if _, err = s.inventoryRepo.RemoveFromInventory(ctx, companyID, resource.ID, quantity); err != nil {
		logger.Error().Err(err).Msg("Failed to remove resource from inventory")
		return nil, err
	}

	// Add revenue to company
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get company")
		return nil, err
	}
	if company == nil {
		return nil, companyModels.ErrCompanyNotFound
	}

	newMoney := company.Money + revenue
	updatedCompany, err := s.companyRepo.UpdateCompanyMoney(ctx, companyID, newMoney)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update company money")
		return nil, err
	}

	// Return full inventory
	inventory, err := s.inventoryRepo.GetInventory(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve inventory after sale")
		return nil, err
	}

	logger.Info().
		Int64("revenue", revenue).
		Int64("new_money", updatedCompany.Money).
		Msg("Resource sold successfully")

	if s.auditRepo != nil {
		userID, _ := ctx.Value("user_id").(string)
		_ = s.auditRepo.Log(ctx, userID, companyID, "SELL_RESOURCE", "resource", resourceMasterID, map[string]interface{}{
			"quantity":    quantity,
			"revenue":     revenue,
			"after_money": updatedCompany.Money,
		})
	}

	return &MarketTransactionResult{
		Company:   updatedCompany,
		Inventory: inventory,
	}, nil
}
