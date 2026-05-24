package service

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/company/repository"
)

// CompanyService defines operations for company business logic
type CompanyService interface {
	CreateCompany(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error)
	GetCompany(ctx context.Context, userID string) (*models.Company, error)
	UpdateCompanyName(ctx context.Context, companyID, newName string) (*models.Company, error)
	DeleteCompany(ctx context.Context, companyID string) error
}

// companyService implements CompanyService
type companyService struct {
	companyRepo   *repository.CompanyRepository
	inventoryRepo *repository.InventoryRepository
}

// NewCompanyService creates a new company service
func NewCompanyService(companyRepo *repository.CompanyRepository, inventoryRepo *repository.InventoryRepository) CompanyService {
	return &companyService{
		companyRepo:   companyRepo,
		inventoryRepo: inventoryRepo,
	}
}

// CreateCompany creates a new company for a user
// Each user can only have one company
func (s *companyService) CreateCompany(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error) {
	logger := log.With().
		Str("user_id", userID).
		Str("company_name", name).
		Int64("initial_money", initialMoney).
		Logger()

	// Validate inputs
	if name == "" {
		logger.Warn().Msg("Company name cannot be empty")
		return nil, models.ErrInvalidAmount // Reusing for invalid input
	}

	if initialMoney < 0 {
		logger.Warn().Int64("money", initialMoney).Msg("Initial money cannot be negative")
		return nil, models.ErrInvalidAmount
	}

	// Delegate to repository
	company, err := s.companyRepo.CreateCompany(ctx, userID, name, initialMoney)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create company")
		return nil, err
	}

	logger.Info().Str("company_id", company.ID).Msg("Company created successfully")
	return company, nil
}

// GetCompany retrieves a user's company
func (s *companyService) GetCompany(ctx context.Context, userID string) (*models.Company, error) {
	logger := log.With().Str("user_id", userID).Logger()

	company, err := s.companyRepo.GetCompanyByUserID(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get company")
		return nil, err
	}

	if company == nil {
		logger.Warn().Msg("Company not found for user")
		return nil, nil
	}

	logger.Debug().Str("company_id", company.ID).Msg("Company retrieved successfully")
	return company, nil
}

// UpdateCompanyName updates a company's name
func (s *companyService) UpdateCompanyName(ctx context.Context, companyID, newName string) (*models.Company, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("new_name", newName).
		Logger()

	// Validate inputs
	if newName == "" {
		logger.Warn().Msg("Company name cannot be empty")
		return nil, models.ErrInvalidAmount
	}

	// Verify company exists
	company, err := s.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to verify company exists")
		return nil, err
	}

	if company == nil {
		logger.Warn().Msg("Company not found")
		return nil, models.ErrInsufficientFunds // Reusing domain error
	}

	// Update name
	updatedCompany, err := s.companyRepo.UpdateCompanyName(ctx, companyID, newName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update company name")
		return nil, err
	}

	logger.Info().Msg("Company name updated successfully")
	return updatedCompany, nil
}

// DeleteCompany soft-deletes a company
func (s *companyService) DeleteCompany(ctx context.Context, companyID string) error {
	logger := log.With().Str("company_id", companyID).Logger()

	// Verify company exists
	exists, err := s.companyRepo.CheckCompanyExists(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to check if company exists")
		return err
	}

	if !exists {
		logger.Warn().Msg("Company not found")
		return models.ErrInsufficientFunds // Reusing domain error
	}

	// Soft delete
	err = s.companyRepo.SoftDeleteCompany(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete company")
		return err
	}

	logger.Info().Msg("Company deleted successfully")
	return nil
}
