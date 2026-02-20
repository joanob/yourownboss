package service

import (
	"context"
	"errors"

	"yourownboss/internal/db"
	"yourownboss/internal/repository"
)

const (
	MinCompanyNameLength = 3
	MaxCompanyNameLength = 50
)

var (
	ErrCompanyAlreadyExists = errors.New("user already has a company")
	ErrCompanyNotFound      = errors.New("company not found")
	ErrInvalidCompanyName   = errors.New("company name must be between 3 and 50 characters")
	ErrInsufficientFunds    = errors.New("insufficient funds")
)

// CompanyService handles company business logic
type CompanyService interface {
	CreateCompany(ctx context.Context, userID int64, name string) (*db.Company, error)
	GetCompanyByUserID(ctx context.Context, userID int64) (*db.Company, error)
	AddMoney(ctx context.Context, companyID int64, amount int64) error
	SubtractMoney(ctx context.Context, companyID int64, amount int64) error
}

type companyService struct {
	companyRepo  repository.CompanyRepository
	initialMoney int64
}

// NewCompanyService creates a new company service
func NewCompanyService(companyRepo repository.CompanyRepository, initialMoney int64) CompanyService {
	return &companyService{
		companyRepo:  companyRepo,
		initialMoney: initialMoney,
	}
}

func (s *companyService) CreateCompany(ctx context.Context, userID int64, name string) (*db.Company, error) {
	// Validate company name
	if len(name) < MinCompanyNameLength || len(name) > MaxCompanyNameLength {
		return nil, ErrInvalidCompanyName
	}

	// Check if user already has a company
	existing, err := s.companyRepo.GetByUserID(ctx, userID)
	if err != nil && err != repository.ErrCompanyNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCompanyAlreadyExists
	}

	// Create company with initial money from config
	company, err := s.companyRepo.Create(ctx, userID, name, s.initialMoney)
	if err != nil {
		if err == repository.ErrCompanyAlreadyExists {
			return nil, ErrCompanyAlreadyExists
		}
		return nil, err
	}

	return company, nil
}

func (s *companyService) GetCompanyByUserID(ctx context.Context, userID int64) (*db.Company, error) {
	company, err := s.companyRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == repository.ErrCompanyNotFound {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}
	return company, nil
}

func (s *companyService) AddMoney(ctx context.Context, companyID int64, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	newMoney := company.Money + amount
	return s.companyRepo.UpdateMoney(ctx, companyID, newMoney)
}

func (s *companyService) SubtractMoney(ctx context.Context, companyID int64, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	if company.Money < amount {
		return ErrInsufficientFunds
	}

	newMoney := company.Money - amount
	return s.companyRepo.UpdateMoney(ctx, companyID, newMoney)
}
