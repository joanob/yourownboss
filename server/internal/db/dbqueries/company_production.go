package dbqueries

import (
	"context"
	"time"
)

// ─── Company Production Buildings ────────────────────────────────────────────

const createCompanyProductionBuilding = `
INSERT INTO company_production_buildings
  (id, company_id, production_building_id, level, construction_ends_at, created_at)
VALUES (?, ?, ?, ?, ?, ?)
`

// CreateCompanyProductionBuildingParams holds parameters for creating a company production building
type CreateCompanyProductionBuildingParams struct {
	ID                   string
	CompanyID            string
	ProductionBuildingID string
	Level                int64
	ConstructionEndsAt   *time.Time
	CreatedAt            time.Time
}

// CreateCompanyProductionBuilding inserts a new company production building record
func (q *Queries) CreateCompanyProductionBuilding(ctx context.Context, arg CreateCompanyProductionBuildingParams) error {
	_, err := q.db.ExecContext(ctx, createCompanyProductionBuilding,
		arg.ID,
		arg.CompanyID,
		arg.ProductionBuildingID,
		arg.Level,
		arg.ConstructionEndsAt,
		arg.CreatedAt,
	)
	return err
}

const getCompanyProductionBuildingByID = `
SELECT id, company_id, production_building_id, level, construction_ends_at, created_at, is_deleted, deleted_at
FROM company_production_buildings
WHERE id = ? AND is_deleted = 0
`

// GetCompanyProductionBuildingByID retrieves a company production building by its ID
func (q *Queries) GetCompanyProductionBuildingByID(ctx context.Context, id string) (CompanyProductionBuilding, error) {
	row := q.db.QueryRowContext(ctx, getCompanyProductionBuildingByID, id)
	var i CompanyProductionBuilding
	err := row.Scan(
		&i.ID,
		&i.CompanyID,
		&i.ProductionBuildingID,
		&i.Level,
		&i.ConstructionEndsAt,
		&i.CreatedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return i, err
}

const getCompanyProductionBuildingsByCompanyID = `
SELECT id, company_id, production_building_id, level, construction_ends_at, created_at, is_deleted, deleted_at
FROM company_production_buildings
WHERE company_id = ? AND is_deleted = 0
ORDER BY created_at ASC
`

// GetCompanyProductionBuildingsByCompanyID retrieves all non-deleted buildings for a company
func (q *Queries) GetCompanyProductionBuildingsByCompanyID(ctx context.Context, companyID string) ([]CompanyProductionBuilding, error) {
	rows, err := q.db.QueryContext(ctx, getCompanyProductionBuildingsByCompanyID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CompanyProductionBuilding
	for rows.Next() {
		var i CompanyProductionBuilding
		if err := rows.Scan(
			&i.ID,
			&i.CompanyID,
			&i.ProductionBuildingID,
			&i.Level,
			&i.ConstructionEndsAt,
			&i.CreatedAt,
			&i.IsDeleted,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const updateCompanyProductionBuildingLevel = `
UPDATE company_production_buildings
SET level = ?, construction_ends_at = ?
WHERE id = ?
`

// UpdateCompanyProductionBuildingLevelParams holds parameters for updating a building's level
type UpdateCompanyProductionBuildingLevelParams struct {
	Level              int64
	ConstructionEndsAt *time.Time
	ID                 string
}

// UpdateCompanyProductionBuildingLevel updates the level and construction_ends_at of a building
func (q *Queries) UpdateCompanyProductionBuildingLevel(ctx context.Context, arg UpdateCompanyProductionBuildingLevelParams) error {
	_, err := q.db.ExecContext(ctx, updateCompanyProductionBuildingLevel,
		arg.Level,
		arg.ConstructionEndsAt,
		arg.ID,
	)
	return err
}

// ─── Production Runs ─────────────────────────────────────────────────────────

const createProductionRun = `
INSERT INTO production_runs
  (id, company_building_id, process_id, production_cycles, started_at, ends_at)
VALUES (?, ?, ?, ?, ?, ?)
`

// CreateProductionRunParams holds parameters for creating a production run
type CreateProductionRunParams struct {
	ID                string
	CompanyBuildingID string
	ProcessID         string
	ProductionCycles  int64
	StartedAt         time.Time
	EndsAt            time.Time
}

// CreateProductionRun inserts a new production run record
func (q *Queries) CreateProductionRun(ctx context.Context, arg CreateProductionRunParams) error {
	_, err := q.db.ExecContext(ctx, createProductionRun,
		arg.ID,
		arg.CompanyBuildingID,
		arg.ProcessID,
		arg.ProductionCycles,
		arg.StartedAt,
		arg.EndsAt,
	)
	return err
}

const getProductionRunByID = `
SELECT id, company_building_id, process_id, production_cycles, started_at, ends_at,
       is_collected, collected_at, is_deleted, deleted_at
FROM production_runs
WHERE id = ? AND is_deleted = 0
`

// GetProductionRunByID retrieves a production run by its ID
func (q *Queries) GetProductionRunByID(ctx context.Context, id string) (ProductionRun, error) {
	row := q.db.QueryRowContext(ctx, getProductionRunByID, id)
	var i ProductionRun
	err := row.Scan(
		&i.ID,
		&i.CompanyBuildingID,
		&i.ProcessID,
		&i.ProductionCycles,
		&i.StartedAt,
		&i.EndsAt,
		&i.IsCollected,
		&i.CollectedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return i, err
}

const getActiveProductionRunByBuildingID = `
SELECT id, company_building_id, process_id, production_cycles, started_at, ends_at,
       is_collected, collected_at, is_deleted, deleted_at
FROM production_runs
WHERE company_building_id = ? AND is_collected = 0 AND is_deleted = 0
LIMIT 1
`

// GetActiveProductionRunByBuildingID retrieves the active (non-collected) production run for a building
func (q *Queries) GetActiveProductionRunByBuildingID(ctx context.Context, companyBuildingID string) (ProductionRun, error) {
	row := q.db.QueryRowContext(ctx, getActiveProductionRunByBuildingID, companyBuildingID)
	var i ProductionRun
	err := row.Scan(
		&i.ID,
		&i.CompanyBuildingID,
		&i.ProcessID,
		&i.ProductionCycles,
		&i.StartedAt,
		&i.EndsAt,
		&i.IsCollected,
		&i.CollectedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return i, err
}

const markProductionRunCollected = `
UPDATE production_runs
SET is_collected = 1, collected_at = ?
WHERE id = ?
`

// MarkProductionRunCollectedParams holds parameters for marking a run as collected
type MarkProductionRunCollectedParams struct {
	CollectedAt *time.Time
	ID          string
}

// MarkProductionRunCollected marks a production run as collected
func (q *Queries) MarkProductionRunCollected(ctx context.Context, arg MarkProductionRunCollectedParams) error {
	_, err := q.db.ExecContext(ctx, markProductionRunCollected, arg.CollectedAt, arg.ID)
	return err
}
