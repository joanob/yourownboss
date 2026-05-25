package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	auditRepository "github.com/joanob/yourownboss/internal/audit/repository"
	companyModels "github.com/joanob/yourownboss/internal/company/models"
	companyrepository "github.com/joanob/yourownboss/internal/company/repository"
	"github.com/joanob/yourownboss/internal/db/dbqueries"
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
	db            *sql.DB
	queries       *dbqueries.Queries
}

// SetDB provides optional transaction support (call from main, not required in tests)
func (s *MarketService) SetDB(db *sql.DB, queries *dbqueries.Queries) {
	s.db = db
	s.queries = queries
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

	// Deduct money and add inventory atomically
	newMoney := company.Money - totalCost
	if s.db != nil && s.queries != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to begin transaction")
			return nil, err
		}
		defer tx.Rollback()
		txq := s.queries.WithTx(tx)

		if _, err := txq.UpdateCompanyMoney(ctx, dbqueries.UpdateCompanyMoneyParams{Money: newMoney, ID: companyID}); err != nil {
			logger.Error().Err(err).Msg("Failed to update company money in transaction")
			return nil, err
		}
		itemID := uuid.New().String()
		if _, err := txq.AddToInventory(ctx, dbqueries.AddToInventoryParams{ID: itemID, CompanyID: companyID, ResourceID: resource.ID, Quantity: quantity}); err != nil {
			logger.Error().Err(err).Msg("Failed to add to inventory in transaction")
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			logger.Error().Err(err).Msg("Failed to commit buy transaction")
			return nil, err
		}
	} else {
		// Fallback (no db configured — used in tests with mock repos)
		if _, err := s.companyRepo.UpdateCompanyMoney(ctx, companyID, newMoney); err != nil {
			logger.Error().Err(err).Msg("Failed to update company money")
			return nil, err
		}
		if _, err := s.inventoryRepo.AddToInventory(ctx, companyID, resource.ID, quantity); err != nil {
			logger.Error().Err(err).Msg("Failed to add resource to inventory")
			return nil, err
		}
	}

	updatedCompany, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil || updatedCompany == nil {
		logger.Error().Err(err).Msg("Failed to reload company after purchase")
		if err == nil {
			err = companyModels.ErrCompanyNotFound
		}
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
		if err := s.auditRepo.Log(ctx, userID, companyID, "BUY_RESOURCE", "resource", resourceMasterID, map[string]interface{}{
			"quantity":    quantity,
			"total_cost":  totalCost,
			"after_money": updatedCompany.Money,
		}); err != nil {
			log.Error().Err(err).Msg("Audit log failed: BUY_RESOURCE")
		}
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

	// Get company to compute new money
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get company")
		return nil, err
	}
	if company == nil {
		return nil, companyModels.ErrCompanyNotFound
	}

	// Remove inventory and add revenue atomically
	newMoney := company.Money + revenue
	newQty := available - quantity
	if s.db != nil && s.queries != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to begin transaction")
			return nil, err
		}
		defer tx.Rollback()
		txq := s.queries.WithTx(tx)

		if _, err := txq.RemoveFromInventory(ctx, dbqueries.RemoveFromInventoryParams{Quantity: newQty, CompanyID: companyID, ResourceID: resource.ID}); err != nil {
			logger.Error().Err(err).Msg("Failed to remove inventory in transaction")
			return nil, err
		}
		if _, err := txq.UpdateCompanyMoney(ctx, dbqueries.UpdateCompanyMoneyParams{Money: newMoney, ID: companyID}); err != nil {
			logger.Error().Err(err).Msg("Failed to update company money in transaction")
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			logger.Error().Err(err).Msg("Failed to commit sell transaction")
			return nil, err
		}
	} else {
		// Fallback (no db configured — used in tests with mock repos)
		if _, err := s.inventoryRepo.RemoveFromInventory(ctx, companyID, resource.ID, quantity); err != nil {
			logger.Error().Err(err).Msg("Failed to remove resource from inventory")
			return nil, err
		}
		if _, err := s.companyRepo.UpdateCompanyMoney(ctx, companyID, newMoney); err != nil {
			logger.Error().Err(err).Msg("Failed to update company money")
			return nil, err
		}
	}

	updatedCompany, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil || updatedCompany == nil {
		logger.Error().Err(err).Msg("Failed to reload company after sale")
		if err == nil {
			err = companyModels.ErrCompanyNotFound
		}
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
		if err := s.auditRepo.Log(ctx, userID, companyID, "SELL_RESOURCE", "resource", resourceMasterID, map[string]interface{}{
			"quantity":    quantity,
			"revenue":     revenue,
			"after_money": updatedCompany.Money,
		}); err != nil {
			log.Error().Err(err).Msg("Audit log failed: SELL_RESOURCE")
		}
	}

	return &MarketTransactionResult{
		Company:   updatedCompany,
		Inventory: inventory,
	}, nil
}
