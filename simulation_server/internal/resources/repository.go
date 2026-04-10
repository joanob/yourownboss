package resources

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// UpsertResources executes inserts/updates within provided tx.
// Caller is responsible for transaction.
func (r *Repository) Exists(ctx context.Context, tx *sql.Tx, id int64) (bool, error) {
	var cnt int
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM resources WHERE id = ?", id)
	} else {
		row = r.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM resources WHERE id = ?", id)
	}
	if err := row.Scan(&cnt); err != nil {
		return false, fmt.Errorf("exists scan: %w", err)
	}
	return cnt > 0, nil
}

func (r *Repository) Insert(ctx context.Context, tx *sql.Tx, id int64, name string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO resources (id, name) VALUES (?, ?)", id, name)
	return err
}

func (r *Repository) Update(ctx context.Context, tx *sql.Tx, id int64, name string) error {
	_, err := tx.ExecContext(ctx, "UPDATE resources SET name = ? WHERE id = ?", name, id)
	return err
}

// List returns all resources ordered by id
func (r *Repository) List(ctx context.Context) ([]ResourceDTO, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM resources ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ResourceDTO
	for rows.Next() {
		var it ResourceDTO
		if err := rows.Scan(&it.ID, &it.Name); err != nil {
			return nil, err
		}
		result = append(result, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
