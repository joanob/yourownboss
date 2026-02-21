package repository

import (
	"context"
	"database/sql"
	"errors"

	"yourownboss/internal/db"
)

var (
	ErrProductionProcessNotFound = errors.New("production process not found")
)

// ProductionProcessRepository handles production process data access.
type ProductionProcessRepository interface {
	GetByID(ctx context.Context, id int64) (*db.ProductionProcess, error)
	GetAllByBuilding(ctx context.Context, buildingID int64) ([]db.ProductionProcess, error)
	Create(
		ctx context.Context,
		id int64,
		name string,
		processingTimeMs int64,
		buildingID int64,
		windowStartHour *int64,
		windowEndHour *int64,
	) (*db.ProductionProcess, error)
	Update(
		ctx context.Context,
		id int64,
		name string,
		processingTimeMs int64,
		buildingID int64,
		windowStartHour *int64,
		windowEndHour *int64,
	) (*db.ProductionProcess, error)
}

type productionProcessRepository struct {
	db *db.DB
}

// NewProductionProcessRepository creates a new production process repository.
func NewProductionProcessRepository(database *db.DB) ProductionProcessRepository {
	return &productionProcessRepository{db: database}
}

func (r *productionProcessRepository) GetByID(ctx context.Context, id int64) (*db.ProductionProcess, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, processing_time_ms, building_id, window_start_hour, window_end_hour
		 FROM production_processes
		 WHERE id = ?`,
		id,
	)

	var process db.ProductionProcess
	var startHour sql.NullInt64
	var endHour sql.NullInt64
	if err := row.Scan(
		&process.ID,
		&process.Name,
		&process.ProcessingTimeMs,
		&process.BuildingID,
		&startHour,
		&endHour,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductionProcessNotFound
		}
		return nil, err
	}

	if startHour.Valid {
		value := startHour.Int64
		process.WindowStartHour = &value
	}
	if endHour.Valid {
		value := endHour.Int64
		process.WindowEndHour = &value
	}

	return &process, nil
}

func (r *productionProcessRepository) GetAllByBuilding(ctx context.Context, buildingID int64) ([]db.ProductionProcess, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, processing_time_ms, building_id, window_start_hour, window_end_hour
		 FROM production_processes
		 WHERE building_id = ?
		 ORDER BY id`,
		buildingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var processes []db.ProductionProcess
	for rows.Next() {
		var process db.ProductionProcess
		var startHour sql.NullInt64
		var endHour sql.NullInt64
		if err := rows.Scan(
			&process.ID,
			&process.Name,
			&process.ProcessingTimeMs,
			&process.BuildingID,
			&startHour,
			&endHour,
		); err != nil {
			return nil, err
		}

		if startHour.Valid {
			value := startHour.Int64
			process.WindowStartHour = &value
		}
		if endHour.Valid {
			value := endHour.Int64
			process.WindowEndHour = &value
		}

		processes = append(processes, process)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return processes, nil
}

func (r *productionProcessRepository) Create(
	ctx context.Context,
	id int64,
	name string,
	processingTimeMs int64,
	buildingID int64,
	windowStartHour *int64,
	windowEndHour *int64,
) (*db.ProductionProcess, error) {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO production_processes (
			id,
			name,
			processing_time_ms,
			building_id,
			window_start_hour,
			window_end_hour
		) VALUES (?, ?, ?, ?, ?, ?)`,
		id,
		name,
		processingTimeMs,
		buildingID,
		nullableInt64(windowStartHour),
		nullableInt64(windowEndHour),
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *productionProcessRepository) Update(
	ctx context.Context,
	id int64,
	name string,
	processingTimeMs int64,
	buildingID int64,
	windowStartHour *int64,
	windowEndHour *int64,
) (*db.ProductionProcess, error) {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE production_processes
		 SET name = ?,
			processing_time_ms = ?,
			building_id = ?,
			window_start_hour = ?,
			window_end_hour = ?
		 WHERE id = ?`,
		name,
		processingTimeMs,
		buildingID,
		nullableInt64(windowStartHour),
		nullableInt64(windowEndHour),
		id,
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func nullableInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}
