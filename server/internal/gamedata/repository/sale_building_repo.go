package repository

import (
	"database/sql"
	"fmt"

	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// SaleBuildingRepo maneja operaciones con edificios de venta en la BD
type SaleBuildingRepo struct {
	db *sql.DB
}

// NewSaleBuildingRepo crea una nueva instancia del repository
func NewSaleBuildingRepo(db *sql.DB) *SaleBuildingRepo {
	return &SaleBuildingRepo{db: db}
}

// GetAll obtiene todos los edificios de venta con sus recursos
func (s *SaleBuildingRepo) GetAll() ([]cache.SaleBuilding, error) {
	// Obtener edificios
	buildingQuery := `
		SELECT id, master_id, name, construction_cost, construction_time_s
		FROM sale_buildings
		WHERE is_deleted = 0
		ORDER BY name
	`

	rows, err := s.db.Query(buildingQuery)
	if err != nil {
		return nil, fmt.Errorf("error al consultar edificios de venta: %w", err)
	}
	defer rows.Close()

	var buildings []cache.SaleBuilding
	for rows.Next() {
		var b cache.SaleBuilding
		if err := rows.Scan(&b.ID, &b.MasterID, &b.Name, &b.ConstructionCost, &b.ConstructionTimeS); err != nil {
			return nil, fmt.Errorf("error al scanear edificio: %w", err)
		}

		// Obtener recursos de venta para este edificio
		resources, err := s.getSaleResources(b.ID)
		if err != nil {
			return nil, err
		}
		b.Resources = resources

		buildings = append(buildings, b)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando edificios: %w", err)
	}

	return buildings, nil
}

// getSaleResources obtiene los recursos de venta de un edificio
func (s *SaleBuildingRepo) getSaleResources(buildingID string) ([]cache.SaleResource, error) {
	resourceQuery := `
		SELECT resource_id, price_per_unit, units_sold_per_second
		FROM sale_resources
		WHERE sale_building_id = ?
		ORDER BY resource_id
	`

	rows, err := s.db.Query(resourceQuery, buildingID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar recursos de venta: %w", err)
	}
	defer rows.Close()

	var resources []cache.SaleResource
	for rows.Next() {
		var r cache.SaleResource
		if err := rows.Scan(&r.ResourceID, &r.PricePerUnit, &r.UnitsSoldPerSecond); err != nil {
			return nil, fmt.Errorf("error al scanear recurso de venta: %w", err)
		}
		resources = append(resources, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando recursos de venta: %w", err)
	}

	return resources, nil
}

// InsertBatch inserta múltiples edificios y sus recursos de venta en la BD (transacción)
func (s *SaleBuildingRepo) InsertBatch(buildings []cache.SaleBuilding) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback()

	// Insertar edificios
	buildingStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO sale_buildings (id, master_id, name, construction_cost, construction_time_s)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparando statement de edificios: %w", err)
	}
	defer buildingStmt.Close()

	// Insertar recursos de venta
	resourceStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO sale_resources (sale_building_id, resource_id, price_per_unit, units_sold_per_second)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparando statement de recursos: %w", err)
	}
	defer resourceStmt.Close()

	for _, b := range buildings {
		if _, err := buildingStmt.Exec(b.ID, b.MasterID, b.Name, b.ConstructionCost, b.ConstructionTimeS); err != nil {
			return fmt.Errorf("error insertando edificio %s: %w", b.MasterID, err)
		}

		for _, r := range b.Resources {
			if _, err := resourceStmt.Exec(b.ID, r.ResourceID, r.PricePerUnit, r.UnitsSoldPerSecond); err != nil {
				return fmt.Errorf("error insertando recurso de venta: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiteando transacción: %w", err)
	}

	return nil
}

// Count retorna el número de edificios en la BD
func (s *SaleBuildingRepo) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sale_buildings WHERE is_deleted = 0").Scan(&count)
	return count, err
}
