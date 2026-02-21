package repository

import (
	"context"
	"database/sql"
	"errors"

	"yourownboss/internal/db"
)

var (
	ErrProductionBuildingNotFound = errors.New("production building not found")
)

// ProductionBuildingRepository handles production building data access.
type ProductionBuildingRepository interface {
	GetByID(ctx context.Context, id int64) (*db.ProductionBuilding, error)
	GetAll(ctx context.Context) ([]db.ProductionBuilding, error)
	Create(ctx context.Context, id int64, name string, cost int64) (*db.ProductionBuilding, error)
	Update(ctx context.Context, id int64, name string, cost int64) (*db.ProductionBuilding, error)
}

type productionBuildingRepository struct {
	db *db.DB
}

// NewProductionBuildingRepository creates a new production building repository.
func NewProductionBuildingRepository(database *db.DB) ProductionBuildingRepository {
	return &productionBuildingRepository{db: database}
}

func (r *productionBuildingRepository) GetByID(ctx context.Context, id int64) (*db.ProductionBuilding, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, cost FROM production_buildings WHERE id = ?`,
		id,
	)

	var building db.ProductionBuilding
	if err := row.Scan(&building.ID, &building.Name, &building.Cost); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductionBuildingNotFound
		}
		return nil, err
	}

	return &building, nil
}

func (r *productionBuildingRepository) GetAll(ctx context.Context) ([]db.ProductionBuilding, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, cost FROM production_buildings ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buildings []db.ProductionBuilding
	for rows.Next() {
		var building db.ProductionBuilding
		if err := rows.Scan(&building.ID, &building.Name, &building.Cost); err != nil {
			return nil, err
		}
		buildings = append(buildings, building)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buildings, nil
}

func (r *productionBuildingRepository) Create(ctx context.Context, id int64, name string, cost int64) (*db.ProductionBuilding, error) {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO production_buildings (id, name, cost) VALUES (?, ?, ?)`,
		id, name, cost,
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *productionBuildingRepository) Update(ctx context.Context, id int64, name string, cost int64) (*db.ProductionBuilding, error) {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE production_buildings SET name = ?, cost = ? WHERE id = ?`,
		name, cost, id,
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}
