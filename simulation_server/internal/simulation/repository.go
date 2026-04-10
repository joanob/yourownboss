package simulation

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// SaveSimulationResult inserts a simulation row and returns the inserted id
func (r *Repository) SaveSimulationResult(ctx context.Context, tx *sql.Tx, processID int64, timeMs int, benefitPerHour float64) (int64, error) {
	res, err := tx.ExecContext(ctx, "INSERT INTO simulations (process_id, time_ms, benefit_per_hour) VALUES (?, ?, ?)", processID, timeMs, benefitPerHour)
	if err != nil {
		return 0, fmt.Errorf("insert simulation: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("lastinsertid: %w", err)
	}
	return id, nil
}

func (r *Repository) SaveSimulationResource(ctx context.Context, tx *sql.Tx, simulationID int64, resourceID int64, isOutput bool, price int64, quantity int64) error {
	out := 0
	if isOutput {
		out = 1
	}
	_, err := tx.ExecContext(ctx, "INSERT INTO simulation_resources (simulation_id, resource_id, is_output, price, quantity) VALUES (?, ?, ?, ?, ?)", simulationID, resourceID, out, price, quantity)
	return err
}

type ProcessRes struct {
	ResourceID int64
	IsOutput   bool
}

// GetProcessResources returns resources linked to a process, with is_output flag
func (r *Repository) GetProcessResources(ctx context.Context, processID int64) ([]ProcessRes, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT resource_id, is_output FROM process_resources WHERE process_id = ?", processID)
	if err != nil {
		return nil, fmt.Errorf("query process resources: %w", err)
	}
	defer rows.Close()

	out := make([]ProcessRes, 0)
	for rows.Next() {
		var rid int64
		var isOut int
		if err := rows.Scan(&rid, &isOut); err != nil {
			return nil, fmt.Errorf("scan process resource: %w", err)
		}
		out = append(out, ProcessRes{ResourceID: rid, IsOutput: isOut == 1})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
