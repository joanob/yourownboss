package dbqueries

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ─── Company Sale Buildings ───────────────────────────────────────────────────

const createCompanySaleBuilding = `
INSERT INTO company_sale_buildings
  (id, company_id, sale_building_id, level, construction_ends_at, created_at)
VALUES (?, ?, ?, ?, ?, ?)
`

// CreateCompanySaleBuildingParams holds parameters for creating a company sale building
type CreateCompanySaleBuildingParams struct {
	ID                 string
	CompanyID          string
	SaleBuildingID     string
	Level              int64
	ConstructionEndsAt *time.Time
	CreatedAt          time.Time
}

// CreateCompanySaleBuilding inserts a new company sale building record
func (q *Queries) CreateCompanySaleBuilding(ctx context.Context, arg CreateCompanySaleBuildingParams) error {
	_, err := q.db.ExecContext(ctx, createCompanySaleBuilding,
		arg.ID,
		arg.CompanyID,
		arg.SaleBuildingID,
		arg.Level,
		arg.ConstructionEndsAt,
		arg.CreatedAt,
	)
	return err
}

const getCompanySaleBuildingByID = `
SELECT id, company_id, sale_building_id, level, construction_ends_at, created_at, is_deleted, deleted_at
FROM company_sale_buildings
WHERE id = ? AND is_deleted = 0
`

// GetCompanySaleBuildingByID retrieves a company sale building by its ID
func (q *Queries) GetCompanySaleBuildingByID(ctx context.Context, id string) (CompanySaleBuilding, error) {
	row := q.db.QueryRowContext(ctx, getCompanySaleBuildingByID, id)
	var i CompanySaleBuilding
	err := row.Scan(
		&i.ID,
		&i.CompanyID,
		&i.SaleBuildingID,
		&i.Level,
		&i.ConstructionEndsAt,
		&i.CreatedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return i, err
}

const getCompanySaleBuildingsByCompanyID = `
SELECT id, company_id, sale_building_id, level, construction_ends_at, created_at, is_deleted, deleted_at
FROM company_sale_buildings
WHERE company_id = ? AND is_deleted = 0
ORDER BY created_at ASC
`

// GetCompanySaleBuildingsByCompanyID retrieves all non-deleted sale buildings for a company
func (q *Queries) GetCompanySaleBuildingsByCompanyID(ctx context.Context, companyID string) ([]CompanySaleBuilding, error) {
	rows, err := q.db.QueryContext(ctx, getCompanySaleBuildingsByCompanyID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CompanySaleBuilding
	for rows.Next() {
		var i CompanySaleBuilding
		if err := rows.Scan(
			&i.ID,
			&i.CompanyID,
			&i.SaleBuildingID,
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

const updateCompanySaleBuildingLevel = `
UPDATE company_sale_buildings
SET level = ?, construction_ends_at = ?
WHERE id = ?
`

// UpdateCompanySaleBuildingLevelParams holds parameters for updating a sale building's level
type UpdateCompanySaleBuildingLevelParams struct {
	Level              int64
	ConstructionEndsAt *time.Time
	ID                 string
}

// UpdateCompanySaleBuildingLevel updates the level and construction_ends_at of a sale building
func (q *Queries) UpdateCompanySaleBuildingLevel(ctx context.Context, arg UpdateCompanySaleBuildingLevelParams) error {
	_, err := q.db.ExecContext(ctx, updateCompanySaleBuildingLevel,
		arg.Level,
		arg.ConstructionEndsAt,
		arg.ID,
	)
	return err
}

// ─── Sale Runs ────────────────────────────────────────────────────────────────

const createSaleRun = `
INSERT INTO sale_runs
  (id, company_sale_building_id, resource_id, units_to_sell, started_at, ends_at)
VALUES (?, ?, ?, ?, ?, ?)
`

// CreateSaleRunParams holds parameters for creating a sale run
type CreateSaleRunParams struct {
	ID                    string
	CompanySaleBuildingID string
	ResourceID            string
	UnitsToSell           int64
	StartedAt             time.Time
	EndsAt                time.Time
}

// CreateSaleRun inserts a new sale run record
func (q *Queries) CreateSaleRun(ctx context.Context, arg CreateSaleRunParams) error {
	_, err := q.db.ExecContext(ctx, createSaleRun,
		arg.ID,
		arg.CompanySaleBuildingID,
		arg.ResourceID,
		arg.UnitsToSell,
		arg.StartedAt,
		arg.EndsAt,
	)
	return err
}

const getSaleRunByID = `
SELECT id, company_sale_building_id, resource_id, units_to_sell, started_at, ends_at,
       is_collected, collected_at, is_deleted, deleted_at
FROM sale_runs
WHERE id = ? AND is_deleted = 0
`

// GetSaleRunByID retrieves a sale run by its ID
func (q *Queries) GetSaleRunByID(ctx context.Context, id string) (SaleRun, error) {
	row := q.db.QueryRowContext(ctx, getSaleRunByID, id)
	var i SaleRun
	err := row.Scan(
		&i.ID,
		&i.CompanySaleBuildingID,
		&i.ResourceID,
		&i.UnitsToSell,
		&i.StartedAt,
		&i.EndsAt,
		&i.IsCollected,
		&i.CollectedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return i, err
}

const getActiveSaleRunByBuildingID = `
SELECT id, company_sale_building_id, resource_id, units_to_sell, started_at, ends_at,
       is_collected, collected_at, is_deleted, deleted_at
FROM sale_runs
WHERE company_sale_building_id = ? AND is_collected = 0 AND is_deleted = 0
LIMIT 1
`

// GetActiveSaleRunByBuildingID retrieves the active (non-collected) sale run for a building
func (q *Queries) GetActiveSaleRunByBuildingID(ctx context.Context, companySaleBuildingID string) (SaleRun, error) {
	row := q.db.QueryRowContext(ctx, getActiveSaleRunByBuildingID, companySaleBuildingID)
	var i SaleRun
	err := row.Scan(
		&i.ID,
		&i.CompanySaleBuildingID,
		&i.ResourceID,
		&i.UnitsToSell,
		&i.StartedAt,
		&i.EndsAt,
		&i.IsCollected,
		&i.CollectedAt,
		&i.IsDeleted,
		&i.DeletedAt,
	)
	return i, err
}

const markSaleRunCollected = `
UPDATE sale_runs
SET is_collected = 1, collected_at = ?
WHERE id = ?
`

// MarkSaleRunCollectedParams holds parameters for marking a sale run as collected
type MarkSaleRunCollectedParams struct {
	CollectedAt *time.Time
	ID          string
}

// MarkSaleRunCollected marks a sale run as collected
func (q *Queries) MarkSaleRunCollected(ctx context.Context, arg MarkSaleRunCollectedParams) error {
	_, err := q.db.ExecContext(ctx, markSaleRunCollected, arg.CollectedAt, arg.ID)
	return err
}

// GetActiveSaleRunsByBuildingIDs returns all active (non-collected) sale runs
// for the given building IDs in a single query, avoiding the N+1 problem in GetBuildings.
func (q *Queries) GetActiveSaleRunsByBuildingIDs(ctx context.Context, buildingIDs []string) ([]SaleRun, error) {
	if len(buildingIDs) == 0 {
		return nil, nil
	}

	placeholders := strings.Repeat("?,", len(buildingIDs))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`
SELECT id, company_sale_building_id, resource_id, units_to_sell, started_at, ends_at,
       is_collected, collected_at, is_deleted, deleted_at
FROM sale_runs
WHERE company_sale_building_id IN (%s) AND is_collected = 0 AND is_deleted = 0
`, placeholders)

	args := make([]interface{}, len(buildingIDs))
	for i, id := range buildingIDs {
		args[i] = id
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SaleRun
	for rows.Next() {
		var i SaleRun
		if err := rows.Scan(
			&i.ID,
			&i.CompanySaleBuildingID,
			&i.ResourceID,
			&i.UnitsToSell,
			&i.StartedAt,
			&i.EndsAt,
			&i.IsCollected,
			&i.CollectedAt,
			&i.IsDeleted,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
