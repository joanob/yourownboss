package repository

import (
	"context"

	"yourownboss/internal/db"
)

// ProductionProcessResourceRepository handles process resource data access.
type ProductionProcessResourceRepository interface {
	GetAllByProcess(ctx context.Context, processID int64) ([]db.ProductionProcessResource, error)
	Upsert(ctx context.Context, processID, resourceID int64, direction string, quantity int64) error
	Delete(ctx context.Context, processID, resourceID int64, direction string) error
}

type productionProcessResourceRepository struct {
	db *db.DB
}

// NewProductionProcessResourceRepository creates a new process resource repository.
func NewProductionProcessResourceRepository(database *db.DB) ProductionProcessResourceRepository {
	return &productionProcessResourceRepository{db: database}
}

func (r *productionProcessResourceRepository) GetAllByProcess(ctx context.Context, processID int64) ([]db.ProductionProcessResource, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT process_id, resource_id, direction, quantity
		 FROM production_process_resources
		 WHERE process_id = ?
		 ORDER BY resource_id, direction`,
		processID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []db.ProductionProcessResource
	for rows.Next() {
		var resource db.ProductionProcessResource
		if err := rows.Scan(&resource.ProcessID, &resource.ResourceID, &resource.Direction, &resource.Quantity); err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *productionProcessResourceRepository) Upsert(
	ctx context.Context,
	processID int64,
	resourceID int64,
	direction string,
	quantity int64,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO production_process_resources (process_id, resource_id, direction, quantity)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(process_id, resource_id, direction)
		 DO UPDATE SET quantity = excluded.quantity`,
		processID,
		resourceID,
		direction,
		quantity,
	)
	return err
}

func (r *productionProcessResourceRepository) Delete(
	ctx context.Context,
	processID int64,
	resourceID int64,
	direction string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM production_process_resources WHERE process_id = ? AND resource_id = ? AND direction = ?`,
		processID,
		resourceID,
		direction,
	)
	return err
}
