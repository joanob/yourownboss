package production

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) BuildingExists(ctx context.Context, tx *sql.Tx, id int64) (bool, error) {
	var cnt int
	row := tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM production_buildings WHERE id = ?", id)
	if err := row.Scan(&cnt); err != nil {
		return false, fmt.Errorf("building exists scan: %w", err)
	}
	return cnt > 0, nil
}

func (r *Repository) InsertBuilding(ctx context.Context, tx *sql.Tx, id int64, name string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO production_buildings (id, name) VALUES (?, ?)", id, name)
	return err
}

func (r *Repository) UpdateBuilding(ctx context.Context, tx *sql.Tx, id int64, name string) error {
	_, err := tx.ExecContext(ctx, "UPDATE production_buildings SET name = ? WHERE id = ?", name, id)
	return err
}

func (r *Repository) DeleteBuilding(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM production_buildings WHERE id = ?", id)
	return err
}

// Processes
func (r *Repository) InsertProcess(ctx context.Context, tx *sql.Tx, id int64, buildingID int64, name string, startHour *int, endHour *int) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO production_processes (id, building_id, name, start_hour, end_hour) VALUES (?, ?, ?, ?, ?)", id, buildingID, name, startHour, endHour)
	return err
}

func (r *Repository) UpdateProcess(ctx context.Context, tx *sql.Tx, id int64, name string, startHour *int, endHour *int) error {
	_, err := tx.ExecContext(ctx, "UPDATE production_processes SET name = ?, start_hour = ?, end_hour = ? WHERE id = ?", name, startHour, endHour, id)
	return err
}

func (r *Repository) DeleteProcessesNotIn(ctx context.Context, tx *sql.Tx, buildingID int64, keepIDs []int64) error {
	// If keepIDs empty, delete all processes for building
	if len(keepIDs) == 0 {
		_, err := tx.ExecContext(ctx, "DELETE FROM production_processes WHERE building_id = ?", buildingID)
		return err
	}
	// build query
	query := "DELETE FROM production_processes WHERE building_id = ? AND id NOT IN ("
	args := make([]interface{}, 0, len(keepIDs)+1)
	args = append(args, buildingID)
	for i, id := range keepIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

// Process resources: we'll delete existing for a process and re-insert
func (r *Repository) DeleteProcessResources(ctx context.Context, tx *sql.Tx, processID int64) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM process_resources WHERE process_id = ?", processID)
	return err
}

func (r *Repository) InsertProcessResource(ctx context.Context, tx *sql.Tx, processID int64, resourceID int64, isOutput bool) error {
	out := 0
	if isOutput {
		out = 1
	}
	_, err := tx.ExecContext(ctx, "INSERT INTO process_resources (process_id, resource_id, is_output) VALUES (?, ?, ?)", processID, resourceID, out)
	return err
}
