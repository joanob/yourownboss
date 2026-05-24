package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// CompanyRepository handles all company-related database operations
type CompanyRepository struct {
	queries *dbqueries.Queries
}

// NewCompanyRepository creates a new company repository
func NewCompanyRepository(queries *dbqueries.Queries) *CompanyRepository {
	return &CompanyRepository{
		queries: queries,
	}
}

// CreateCompany creates a new company for a user with initial money
// Only one company per user is allowed
func (r *CompanyRepository) CreateCompany(ctx context.Context, userID, name string, initialMoney int64) (*models.Company, error) {
	logger := log.With().
		Str("user_id", userID).
		Str("company_name", name).
		Logger()

	companyID := uuid.New().String()

	// Check if company already exists for this user
	count, err := r.queries.CheckCompanyExists(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to check if company exists")
		return nil, err
	}
	if count > 0 {
		logger.Warn().Msg("Company already exists for user")
		return nil, models.ErrCompanyAlreadyExists
	}

	// Create company
	dbo, err := r.queries.CreateCompany(ctx, dbqueries.CreateCompanyParams{
		ID:     companyID,
		UserID: userID,
		Name:   name,
		Money:  initialMoney,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create company")
		return nil, err
	}

	logger.Info().Str("company_id", companyID).Msg("Company created successfully")
	return dboDomainModel(&dbo), nil
}

// GetCompanyByID retrieves a company by its ID
func (r *CompanyRepository) GetCompanyByID(ctx context.Context, companyID string) (*models.Company, error) {
	dbo, err := r.queries.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return dboDomainModel(&dbo), nil
}

// GetCompanyByUserID retrieves the company associated with a user
// Users can have at most one company
func (r *CompanyRepository) GetCompanyByUserID(ctx context.Context, userID string) (*models.Company, error) {
	logger := log.With().Str("user_id", userID).Logger()

	dbo, err := r.queries.GetCompanyByUserID(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get company by user ID")
		return nil, err
	}

	return dboDomainModel(&dbo), nil
}

// UpdateCompanyName updates the company name
func (r *CompanyRepository) UpdateCompanyName(ctx context.Context, companyID, newName string) (*models.Company, error) {
	logger := log.With().
		Str("company_id", companyID).
		Str("new_name", newName).
		Logger()

	dbo, err := r.queries.UpdateCompanyName(ctx, dbqueries.UpdateCompanyNameParams{
		Name: newName,
		ID:   companyID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update company name")
		return nil, err
	}

	logger.Info().Msg("Company name updated successfully")
	return dboDomainModel(&dbo), nil
}

// UpdateCompanyMoney updates the company's money balance
// This is used for atomic money operations (buying, selling, construction costs)
func (r *CompanyRepository) UpdateCompanyMoney(ctx context.Context, companyID string, newMoney int64) (*models.Company, error) {
	logger := log.With().
		Str("company_id", companyID).
		Int64("new_money", newMoney).
		Logger()

	if newMoney < 0 {
		logger.Error().Int64("amount", newMoney).Msg("Money cannot be negative")
		return nil, models.ErrInvalidAmount
	}

	dbo, err := r.queries.UpdateCompanyMoney(ctx, dbqueries.UpdateCompanyMoneyParams{
		Money: newMoney,
		ID:    companyID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update company money")
		return nil, err
	}

	logger.Info().Msg("Company money updated successfully")
	return dboDomainModel(&dbo), nil
}

// CheckCompanyExists checks if a user has a company
func (r *CompanyRepository) CheckCompanyExists(ctx context.Context, userID string) (bool, error) {
	count, err := r.queries.CheckCompanyExists(ctx, userID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SoftDeleteCompany performs a soft delete on the company
func (r *CompanyRepository) SoftDeleteCompany(ctx context.Context, companyID string) error {
	logger := log.With().Str("company_id", companyID).Logger()

	err := r.queries.SoftDeleteCompany(ctx, companyID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to soft delete company")
		return err
	}

	logger.Info().Msg("Company soft deleted successfully")
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// dboDomainModel converts a DBO to a domain model
func dboDomainModel(dbo *dbqueries.Company) *models.Company {
	if dbo == nil {
		return nil
	}

	return &models.Company{
		ID:        dbo.ID,
		UserID:    dbo.UserID,
		Name:      dbo.Name,
		Money:     dbo.Money,
		CreatedAt: dbo.CreatedAt,
		IsDeleted: dbo.IsDeleted == 1,
		DeletedAt: dbo.DeletedAt,
	}
}
