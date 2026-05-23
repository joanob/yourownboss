package repository

import (
	"database/sql"
	"fmt"

	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// ResourceRepo maneja operaciones con recursos en la BD
type ResourceRepo struct {
	db *sql.DB
}

// NewResourceRepo crea una nueva instancia del repository
func NewResourceRepo(db *sql.DB) *ResourceRepo {
	return &ResourceRepo{db: db}
}

// GetAll obtiene todos los recursos de la BD
func (r *ResourceRepo) GetAll() ([]cache.Resource, error) {
	query := `
		SELECT id, master_id, name, market_price, market_sale_qty
		FROM resources
		WHERE is_deleted = 0
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar recursos: %w", err)
	}
	defer rows.Close()

	var resources []cache.Resource
	for rows.Next() {
		var res cache.Resource
		if err := rows.Scan(&res.ID, &res.MasterID, &res.Name, &res.MarketPrice, &res.MarketSaleQty); err != nil {
			return nil, fmt.Errorf("error al scanear recurso: %w", err)
		}
		resources = append(resources, res)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando recursos: %w", err)
	}

	return resources, nil
}

// InsertBatch inserta múltiples recursos en la BD
func (r *ResourceRepo) InsertBatch(resources []cache.Resource) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO resources (id, master_id, name, market_price, market_sale_qty)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparando statement: %w", err)
	}
	defer stmt.Close()

	for _, res := range resources {
		if _, err := stmt.Exec(res.ID, res.MasterID, res.Name, res.MarketPrice, res.MarketSaleQty); err != nil {
			return fmt.Errorf("error insertando recurso %s: %w", res.MasterID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiteando transacción: %w", err)
	}

	return nil
}

// Count retorna el número de recursos en la BD
func (r *ResourceRepo) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM resources WHERE is_deleted = 0").Scan(&count)
	return count, err
}
