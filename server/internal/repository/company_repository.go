package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"yourownboss/internal/db"
)

var (
	ErrCompanyAlreadyExists = errors.New("user already has a company")
	ErrCompanyNotFound      = errors.New("company not found")
)

// CompanyRepository handles company data access
type CompanyRepository interface {
	Create(ctx context.Context, userID int64, name string, initialMoney int64) (*db.Company, error)
	GetByUserID(ctx context.Context, userID int64) (*db.Company, error)
	GetByID(ctx context.Context, id int64) (*db.Company, error)
	UpdateMoney(ctx context.Context, id int64, newMoney int64) error
	Update(ctx context.Context, company *db.Company) error
}

type companyRepository struct {
	db *db.DB
}

// NewCompanyRepository creates a new company repository
func NewCompanyRepository(database *db.DB) CompanyRepository {
	return &companyRepository{db: database}
}

func (r *companyRepository) Create(ctx context.Context, userID int64, name string, initialMoney int64) (*db.Company, error) {
	result, err := r.db.ExecContext(
		ctx,
		"INSERT INTO companies (user_id, name, money) VALUES (?, ?, ?)",
		userID, name, initialMoney,
	)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "UNIQUE constraint failed: companies.user_id" {
			return nil, ErrCompanyAlreadyExists
		}
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &db.Company{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Money:     initialMoney,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (r *companyRepository) GetByUserID(ctx context.Context, userID int64) (*db.Company, error) {
	var company db.Company
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, name, money, created_at, updated_at FROM companies WHERE user_id = ?",
		userID,
	).Scan(&company.ID, &company.UserID, &company.Name, &company.Money, &company.CreatedAt, &company.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrCompanyNotFound
	}
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (r *companyRepository) GetByID(ctx context.Context, id int64) (*db.Company, error) {
	var company db.Company
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, name, money, created_at, updated_at FROM companies WHERE id = ?",
		id,
	).Scan(&company.ID, &company.UserID, &company.Name, &company.Money, &company.CreatedAt, &company.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrCompanyNotFound
	}
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (r *companyRepository) UpdateMoney(ctx context.Context, id int64, newMoney int64) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE companies SET money = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newMoney, id,
	)
	return err
}

func (r *companyRepository) Update(ctx context.Context, company *db.Company) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE companies SET name = ?, money = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		company.Name, company.Money, company.ID,
	)
	return err
}
