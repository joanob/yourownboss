package service

import (
	"context"
	"errors"

	"yourownboss/internal/db"
	"yourownboss/internal/repository"
)

var (
	ErrMarketInsufficientFunds = errors.New("insufficient funds to buy")
	ErrResourceDoesNotExist    = errors.New("resource does not exist")
)

// InventoryService handles inventory business logic
type InventoryService interface {
	GetInventory(ctx context.Context, companyID int64) ([]db.InventoryWithDetails, error)
	GetResource(ctx context.Context, resourceID int64) (*db.Resource, error)
	GetAllResources(ctx context.Context) ([]db.Resource, error)
}

// MarketService handles buying and selling
type MarketService interface {
	BuyResource(ctx context.Context, companyID, resourceID int64, packCount int64) error
	SellResource(ctx context.Context, companyID, resourceID int64, packCount int64) error
}

type inventoryService struct {
	resourceRepo   repository.ResourceRepository
	inventoryRepo  repository.InventoryRepository
}

type marketService struct {
	resourceRepo  repository.ResourceRepository
	companyRepo   repository.CompanyRepository
	inventoryRepo repository.InventoryRepository
}

// NewInventoryService creates a new inventory service
func NewInventoryService(
	resourceRepo repository.ResourceRepository,
	inventoryRepo repository.InventoryRepository,
) InventoryService {
	return &inventoryService{
		resourceRepo:   resourceRepo,
		inventoryRepo:  inventoryRepo,
	}
}

// NewMarketService creates a new market service
func NewMarketService(
	resourceRepo repository.ResourceRepository,
	companyRepo repository.CompanyRepository,
	inventoryRepo repository.InventoryRepository,
) MarketService {
	return &marketService{
		resourceRepo:  resourceRepo,
		companyRepo:   companyRepo,
		inventoryRepo: inventoryRepo,
	}
}

// --- Inventory Service Implementation ---

func (s *inventoryService) GetInventory(ctx context.Context, companyID int64) ([]db.InventoryWithDetails, error) {
	return s.inventoryRepo.GetAllByCompanyWithDetails(ctx, companyID)
}

func (s *inventoryService) GetResource(ctx context.Context, resourceID int64) (*db.Resource, error) {
	return s.resourceRepo.GetByID(ctx, resourceID)
}

func (s *inventoryService) GetAllResources(ctx context.Context) ([]db.Resource, error) {
	return s.resourceRepo.GetAll(ctx)
}

// --- Market Service Implementation ---

// BuyResource buys packCount number of packs of a resource
// Each pack contains resource.PackSize units and costs resource.Price
func (s *marketService) BuyResource(ctx context.Context, companyID, resourceID int64, packCount int64) error {
	if packCount <= 0 {
		return errors.New("pack count must be positive")
	}

	// Get the resource
	resource, err := s.resourceRepo.GetByID(ctx, resourceID)
	if err != nil {
		return err
	}

	totalCost := resource.Price * packCount
	totalUnits := resource.PackSize * packCount

	// Get company
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	// Check if company has enough funds
	if company.Money < totalCost {
		return ErrMarketInsufficientFunds
	}

	// Deduct money from company
	if err := s.companyRepo.UpdateMoney(ctx, companyID, company.Money-totalCost); err != nil {
		return err
	}

	// Add items to inventory
	if err := s.inventoryRepo.AddItem(ctx, companyID, resourceID, totalUnits); err != nil {
		// Rollback: return money if inventory add fails
		_ = s.companyRepo.UpdateMoney(ctx, companyID, company.Money)
		return err
	}

	return nil
}

// SellResource sells packCount number of packs of a resource
// Each pack contains resource.PackSize units and is sold for resource.Price
func (s *marketService) SellResource(ctx context.Context, companyID, resourceID int64, packCount int64) error {
	if packCount <= 0 {
		return errors.New("pack count must be positive")
	}

	// Get the resource
	resource, err := s.resourceRepo.GetByID(ctx, resourceID)
	if err != nil {
		return err
	}

	totalRevenue := resource.Price * packCount
	totalUnits := resource.PackSize * packCount

	// Check if company has enough items to sell
	inv, err := s.inventoryRepo.GetByCompanyAndResource(ctx, companyID, resourceID)
	if err != nil {
		if err == repository.ErrInventoryNotFound {
			return repository.ErrInsufficientStock
		}
		return err
	}

	if inv.Quantity < totalUnits {
		return repository.ErrInsufficientStock
	}

	// Get company current money
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	// Add money to company
	if err := s.companyRepo.UpdateMoney(ctx, companyID, company.Money+totalRevenue); err != nil {
		return err
	}

	// Remove items from inventory
	if err := s.inventoryRepo.RemoveItem(ctx, companyID, resourceID, totalUnits); err != nil {
		// Rollback: return money if removal fails
		_ = s.companyRepo.UpdateMoney(ctx, companyID, company.Money)
		return err
	}

	return nil
}
