package repository

import (
	"database/sql"
	"fmt"

	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// ProductionBuildingRepo maneja operaciones con edificios de producción en la BD
type ProductionBuildingRepo struct {
	db *sql.DB
}

// NewProductionBuildingRepo crea una nueva instancia del repository
func NewProductionBuildingRepo(db *sql.DB) *ProductionBuildingRepo {
	return &ProductionBuildingRepo{db: db}
}

// GetAll obtiene todos los edificios de producción con sus procesos y recursos
func (p *ProductionBuildingRepo) GetAll() ([]cache.ProductionBuilding, error) {
	// Obtener edificios
	buildingQuery := `
		SELECT id, master_id, name, construction_cost, construction_time_s
		FROM production_buildings
		WHERE is_deleted = 0
		ORDER BY name
	`

	rows, err := p.db.Query(buildingQuery)
	if err != nil {
		return nil, fmt.Errorf("error al consultar edificios de producción: %w", err)
	}
	defer rows.Close()

	var buildings []cache.ProductionBuilding
	for rows.Next() {
		var b cache.ProductionBuilding
		if err := rows.Scan(&b.ID, &b.MasterID, &b.Name, &b.ConstructionCost, &b.ConstructionTimeS); err != nil {
			return nil, fmt.Errorf("error al scanear edificio: %w", err)
		}

		// Obtener procesos para este edificio
		processes, err := p.getProcessesForBuilding(b.ID)
		if err != nil {
			return nil, err
		}
		b.Processes = processes

		buildings = append(buildings, b)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando edificios: %w", err)
	}

	return buildings, nil
}

// getProcessesForBuilding obtiene los procesos de un edificio
func (p *ProductionBuildingRepo) getProcessesForBuilding(buildingID string) ([]cache.ProductionProcess, error) {
	processQuery := `
		SELECT id, master_id, production_building_id, name, cycle_time_s, window_start_hour, window_end_hour
		FROM production_processes
		WHERE production_building_id = ? AND is_deleted = 0
		ORDER BY name
	`

	rows, err := p.db.Query(processQuery, buildingID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar procesos: %w", err)
	}
	defer rows.Close()

	var processes []cache.ProductionProcess
	for rows.Next() {
		var proc cache.ProductionProcess
		if err := rows.Scan(&proc.ID, &proc.MasterID, &proc.ProductionBuildingID, &proc.Name, &proc.CycleTimeS, &proc.WindowStartHour, &proc.WindowEndHour); err != nil {
			return nil, fmt.Errorf("error al scanear proceso: %w", err)
		}

		// Obtener recursos del proceso
		resources, err := p.getProcessResources(proc.ID)
		if err != nil {
			return nil, err
		}
		proc.Resources = resources

		processes = append(processes, proc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando procesos: %w", err)
	}

	return processes, nil
}

// getProcessResources obtiene los recursos de un proceso
func (p *ProductionBuildingRepo) getProcessResources(processID string) ([]cache.ProductionProcessResource, error) {
	resourceQuery := `
		SELECT resource_id, is_output, quantity
		FROM production_process_resources
		WHERE process_id = ?
		ORDER BY is_output, resource_id
	`

	rows, err := p.db.Query(resourceQuery, processID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar recursos de proceso: %w", err)
	}
	defer rows.Close()

	var resources []cache.ProductionProcessResource
	for rows.Next() {
		var r cache.ProductionProcessResource
		if err := rows.Scan(&r.ResourceID, &r.IsOutput, &r.Quantity); err != nil {
			return nil, fmt.Errorf("error al scanear recurso de proceso: %w", err)
		}
		resources = append(resources, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando recursos de proceso: %w", err)
	}

	return resources, nil
}

// InsertBatch inserta múltiples edificios y sus procesos en la BD (transacción)
func (p *ProductionBuildingRepo) InsertBatch(buildings []cache.ProductionBuilding) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback()

	// Insertar edificios
	buildingStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO production_buildings (id, master_id, name, construction_cost, construction_time_s)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparando statement de edificios: %w", err)
	}
	defer buildingStmt.Close()

	// Insertar procesos
	processStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO production_processes (id, master_id, production_building_id, name, cycle_time_s, window_start_hour, window_end_hour)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparando statement de procesos: %w", err)
	}
	defer processStmt.Close()

	// Insertar recursos de procesos
	resourceStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO production_process_resources (process_id, resource_id, is_output, quantity)
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

		for _, proc := range b.Processes {
			if _, err := processStmt.Exec(proc.ID, proc.MasterID, proc.ProductionBuildingID, proc.Name, proc.CycleTimeS, proc.WindowStartHour, proc.WindowEndHour); err != nil {
				return fmt.Errorf("error insertando proceso %s: %w", proc.MasterID, err)
			}

			for _, r := range proc.Resources {
				if _, err := resourceStmt.Exec(proc.ID, r.ResourceID, r.IsOutput, r.Quantity); err != nil {
					return fmt.Errorf("error insertando recurso de proceso: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiteando transacción: %w", err)
	}

	return nil
}

// Count retorna el número de edificios en la BD
func (p *ProductionBuildingRepo) Count() (int, error) {
	var count int
	err := p.db.QueryRow("SELECT COUNT(*) FROM production_buildings WHERE is_deleted = 0").Scan(&count)
	return count, err
}
