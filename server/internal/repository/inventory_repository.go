package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

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
	Create(ctx context.Context, name, icon, description string, price int64, packSize int64) (*db.Resource, error)
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
	row := r.db.QueryRow(
		`SELECT id, name, icon, description, price, pack_size, created_at FROM resources WHERE id = ?`,
		id,
	)

	var resource db.Resource
	if err := row.Scan(&resource.ID, &resource.Name, &resource.Icon, &resource.Description, &resource.Price, &resource.PackSize, &resource.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	return &resource, nil
}

func (r *resourceRepository) GetAll(ctx context.Context) ([]db.Resource, error) {
	query := `SELECT id, name, icon, description, price, pack_size, created_at FROM resources`
	log.Printf("DEBUG: Executing query: %s", query)

	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("ERROR: Query failed: %v", err)
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var resources []db.Resource
	rowCount := 0

	for rows.Next() {
		rowCount++
		var resource db.Resource
		if err := rows.Scan(&resource.ID, &resource.Name, &resource.Icon, &resource.Description, &resource.Price, &resource.PackSize, &resource.CreatedAt); err != nil {
			log.Printf("ERROR: Scan failed on row %d: %v", rowCount, err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		log.Printf("DEBUG: Scanned row %d: ID=%d, Name=%s", rowCount, resource.ID, resource.Name)
		resources = append(resources, resource)
	}

	log.Printf("DEBUG: Total rows scanned: %d", rowCount)

	if err := rows.Err(); err != nil {
		log.Printf("ERROR: rows.Err(): %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	log.Printf("DEBUG: Returning %d resources", len(resources))
	return resources, nil
}

func (r *resourceRepository) Create(ctx context.Context, name, icon, description string, price int64, packSize int64) (*db.Resource, error) {
	result, err := r.db.Exec(
		`INSERT INTO resources (name, icon, description, price, pack_size) VALUES (?, ?, ?, ?, ?)`,
		name, icon, description, price, packSize,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
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
	row := i.db.QueryRow(
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
	rows, err := i.db.Query(
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
	rows, err := i.db.Query(
		`SELECT ci.id, ci.resource_id, r.name, r.icon, ci.quantity, r.price, r.pack_size
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
		if err := rows.Scan(&item.ID, &item.ResourceID, &item.Name, &item.Icon, &item.Quantity, &item.Price, &item.PackSize); err != nil {
			return nil, err
		}
		details = append(details, item)
	}

	return details, rows.Err()
}

func (i *inventoryRepository) AddItem(ctx context.Context, companyID, resourceID int64, quantity int64) error {
	_, err := i.db.Exec(
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

	_, err = i.db.Exec(
		`UPDATE company_inventory SET quantity = quantity - ?, updated_at = CURRENT_TIMESTAMP WHERE company_id = ? AND resource_id = ?`,
		quantity, companyID, resourceID,
	)
	return err
}

func (i *inventoryRepository) SetQuantity(ctx context.Context, companyID, resourceID int64, quantity int64) error {
	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	_, err := i.db.Exec(
		`INSERT INTO company_inventory (company_id, resource_id, quantity, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(company_id, resource_id) DO UPDATE SET quantity = ?, updated_at = CURRENT_TIMESTAMP`,
		companyID, resourceID, quantity, quantity,
	)
	return err
}
