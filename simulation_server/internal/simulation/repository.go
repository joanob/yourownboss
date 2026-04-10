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

// ProcessResDetailed includes resource name
type ProcessResDetailed struct {
	ResourceID int64  `json:"resource_id"`
	IsOutput   bool   `json:"is_output"`
	Name       string `json:"name"`
}

// GetProcessResourcesDetailed returns resources linked to a process with names
func (r *Repository) GetProcessResourcesDetailed(ctx context.Context, processID int64) ([]ProcessResDetailed, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT pr.resource_id, pr.is_output, rs.name FROM process_resources pr JOIN resources rs ON pr.resource_id = rs.id WHERE pr.process_id = ?`, processID)
	if err != nil {
		return nil, fmt.Errorf("query process resources detailed: %w", err)
	}
	defer rows.Close()

	out := make([]ProcessResDetailed, 0)
	for rows.Next() {
		var rid int64
		var isOut int
		var name string
		if err := rows.Scan(&rid, &isOut, &name); err != nil {
			return nil, fmt.Errorf("scan process resource detailed: %w", err)
		}
		out = append(out, ProcessResDetailed{ResourceID: rid, IsOutput: isOut == 1, Name: name})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// SimulationResult represents a simulation and its resources
type SimulationResult struct {
	ID             int64   `json:"id"`
	ProcessID      int64   `json:"process_id"`
	TimeMs         int     `json:"time_ms"`
	BenefitPerHour float64 `json:"benefit_per_hour"`
	Resources      []struct {
		ResourceID int64 `json:"resource_id"`
		IsOutput   bool  `json:"is_output"`
		Price      int64 `json:"price"`
		Quantity   int64 `json:"quantity"`
	} `json:"resources"`
}

// ListSimulationsByProcess returns simulations and their resources for a process
func (r *Repository) ListSimulationsByProcess(ctx context.Context, processID int64) ([]SimulationResult, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, process_id, time_ms, benefit_per_hour FROM simulations WHERE process_id = ? ORDER BY id DESC", processID)
	if err != nil {
		return nil, fmt.Errorf("query simulations: %w", err)
	}
	defer rows.Close()

	var out []SimulationResult
	for rows.Next() {
		var sr SimulationResult
		if err := rows.Scan(&sr.ID, &sr.ProcessID, &sr.TimeMs, &sr.BenefitPerHour); err != nil {
			return nil, fmt.Errorf("scan simulation: %w", err)
		}
		// load resources for this simulation
		rrows, err := r.db.QueryContext(ctx, "SELECT resource_id, is_output, price, quantity FROM simulation_resources WHERE simulation_id = ?", sr.ID)
		if err != nil {
			return nil, fmt.Errorf("query simulation resources: %w", err)
		}
		sr.Resources = make([]struct {
			ResourceID int64 `json:"resource_id"`
			IsOutput   bool  `json:"is_output"`
			Price      int64 `json:"price"`
			Quantity   int64 `json:"quantity"`
		}, 0)
		for rrows.Next() {
			var rid int64
			var isOut int
			var price int64
			var qty int64
			if err := rrows.Scan(&rid, &isOut, &price, &qty); err != nil {
				rrows.Close()
				return nil, fmt.Errorf("scan simulation resource: %w", err)
			}
			sr.Resources = append(sr.Resources, struct {
				ResourceID int64 `json:"resource_id"`
				IsOutput   bool  `json:"is_output"`
				Price      int64 `json:"price"`
				Quantity   int64 `json:"quantity"`
			}{ResourceID: rid, IsOutput: isOut == 1, Price: price, Quantity: qty})
		}
		rrows.Close()
		out = append(out, sr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
