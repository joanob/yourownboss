package repository

import (
	"context"
	"database/sql"
	"errors"

	"yourownboss/internal/db"
)

var (
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrResourceNotFound  = errors.New("resource not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// ResourceRepository handles resource data access
type ResourceRepository interface {
	GetByID(ctx context.Context, id int64) (*db.Resource, error)
	GetAll(ctx context.Context) ([]db.Resource, error)
	Create(ctx context.Context, id int64, name string, price int64, packSize int64) (*db.Resource, error)
	Update(ctx context.Context, id int64, name string, price int64, packSize int64) (*db.Resource, error)
}

// InventoryRepository handles company inventory data access
type InventoryRepository interface {
	GetByCompanyAndResource(ctx context.Context, companyID, resourceID int64) (*db.CompanyInventory, error)
	GetAllByCompany(ctx context.Context, companyID int64) ([]db.CompanyInventory, error)
	GetAllByCompanyWithDetails(ctx context.Context, companyID int64) ([]db.InventoryWithDetails, error)
	AddItem(ctx context.Context, companyID, resourceID int64, quantity int64) error
	RemoveItem(ctx context.Context, companyID, resourceID int64, quantity int64) error
	SetQuantity(ctx context.Context, companyID, resourceID int64, quantity int64) error
}

// --- Resource Repository Implementation ---

type resourceRepository struct {
	db *db.DB
}

func NewResourceRepository(database *db.DB) ResourceRepository {
	return &resourceRepository{db: database}
}

func (r *resourceRepository) GetByID(ctx context.Context, id int64) (*db.Resource, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, price, pack_size, created_at FROM resources WHERE id = ?`,
		id,
	)

	var resource db.Resource
	if err := row.Scan(&resource.ID, &resource.Name, &resource.Price, &resource.PackSize, &resource.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	return &resource, nil
}

func (r *resourceRepository) GetAll(ctx context.Context) ([]db.Resource, error) {
	query := `SELECT id, name, price, pack_size, created_at FROM resources`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []db.Resource
	for rows.Next() {
		var resource db.Resource
		if err := rows.Scan(&resource.ID, &resource.Name, &resource.Price, &resource.PackSize, &resource.CreatedAt); err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *resourceRepository) Create(ctx context.Context, id int64, name string, price int64, packSize int64) (*db.Resource, error) {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO resources (id, name, price, pack_size) VALUES (?, ?, ?, ?)`,
		id, name, price, packSize,
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *resourceRepository) Update(ctx context.Context, id int64, name string, price int64, packSize int64) (*db.Resource, error) {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE resources SET name = ?, price = ?, pack_size = ? WHERE id = ?`,
		name, price, packSize, id,
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

// --- Inventory Repository Implementation ---

type inventoryRepository struct {
	db *db.DB
}

func NewInventoryRepository(database *db.DB) InventoryRepository {
	return &inventoryRepository{db: database}
}

func (i *inventoryRepository) GetByCompanyAndResource(ctx context.Context, companyID, resourceID int64) (*db.CompanyInventory, error) {
	row := i.db.QueryRowContext(
		ctx,
		`SELECT id, company_id, resource_id, quantity, created_at, updated_at FROM company_inventory WHERE company_id = ? AND resource_id = ?`,
		companyID, resourceID,
	)

	var inv db.CompanyInventory
	if err := row.Scan(&inv.ID, &inv.CompanyID, &inv.ResourceID, &inv.Quantity, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInventoryNotFound
		}
		return nil, err
	}

	return &inv, nil
}

func (i *inventoryRepository) GetAllByCompany(ctx context.Context, companyID int64) ([]db.CompanyInventory, error) {
	rows, err := i.db.QueryContext(
		ctx,
		`SELECT id, company_id, resource_id, quantity, created_at, updated_at FROM company_inventory WHERE company_id = ? ORDER BY resource_id`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []db.CompanyInventory
	for rows.Next() {
		var inv db.CompanyInventory
		if err := rows.Scan(&inv.ID, &inv.CompanyID, &inv.ResourceID, &inv.Quantity, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		inventories = append(inventories, inv)
	}

	return inventories, rows.Err()
}

func (i *inventoryRepository) GetAllByCompanyWithDetails(ctx context.Context, companyID int64) ([]db.InventoryWithDetails, error) {
	rows, err := i.db.QueryContext(
		ctx,
		`SELECT ci.id, ci.resource_id, r.name, ci.quantity, r.price, r.pack_size
		 FROM company_inventory ci
		 JOIN resources r ON ci.resource_id = r.id
		 WHERE ci.company_id = ?
		 ORDER BY r.name`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []db.InventoryWithDetails
	for rows.Next() {
		var item db.InventoryWithDetails
		if err := rows.Scan(&item.ID, &item.ResourceID, &item.Name, &item.Quantity, &item.Price, &item.PackSize); err != nil {
			return nil, err
		}
		details = append(details, item)
	}

	return details, rows.Err()
}

func (i *inventoryRepository) AddItem(ctx context.Context, companyID, resourceID int64, quantity int64) error {
	_, err := i.db.ExecContext(
		ctx,
		`INSERT INTO company_inventory (company_id, resource_id, quantity, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(company_id, resource_id) DO UPDATE SET quantity = quantity + ?, updated_at = CURRENT_TIMESTAMP`,
		companyID, resourceID, quantity, quantity,
	)
	return err
}

func (i *inventoryRepository) RemoveItem(ctx context.Context, companyID, resourceID int64, quantity int64) error {
	// Check if we have enough stock first
	inv, err := i.GetByCompanyAndResource(ctx, companyID, resourceID)
	if err != nil {
		if err == ErrInventoryNotFound {
			return ErrInsufficientStock
		}
		return err
	}

	if inv.Quantity < quantity {
		return ErrInsufficientStock
	}

	_, err = i.db.ExecContext(
		ctx,
		`UPDATE company_inventory SET quantity = quantity - ?, updated_at = CURRENT_TIMESTAMP WHERE company_id = ? AND resource_id = ?`,
		quantity, companyID, resourceID,
	)
	return err
}

func (i *inventoryRepository) SetQuantity(ctx context.Context, companyID, resourceID int64, quantity int64) error {
	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	_, err := i.db.ExecContext(
		ctx,
		`INSERT INTO company_inventory (company_id, resource_id, quantity, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(company_id, resource_id) DO UPDATE SET quantity = ?, updated_at = CURRENT_TIMESTAMP`,
		companyID, resourceID, quantity, quantity,
	)
	return err
}
