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

func (r *Repository) SaveSimulationResource(ctx context.Context, tx *sql.Tx, simulationID int64, resourceID int64, isOutput bool, price float64) error {
	out := 0
	if isOutput {
		out = 1
	}
	_, err := tx.ExecContext(ctx, "INSERT INTO simulation_resources (simulation_id, resource_id, is_output, price) VALUES (?, ?, ?, ?)", simulationID, resourceID, out, price)
	return err
}
