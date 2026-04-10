package resources

import (
	"context"
	"database/sql"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// UpsertResources validates payload and performs transactional upsert.
func (s *Service) UpsertResources(ctx context.Context, db *sql.DB, items []ResourceDTO) error {
	if len(items) == 0 {
		return errors.New("no resources")
	}

	// detect duplicate ids in payload
	seen := make(map[int64]bool)
	for _, it := range items {
		if it.ID == 0 {
			return errors.New("resource id required and must be non-zero")
		}
		if it.Name == "" {
			return errors.New("resource name required")
		}
		if seen[it.ID] {
			return errors.New("duplicate id in payload")
		}
		seen[it.ID] = true
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	applied := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		exists, err := s.repo.Exists(ctx, tx, it.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
		if exists {
			if err := s.repo.Update(ctx, tx, it.ID, it.Name); err != nil {
				tx.Rollback()
				return err
			}
			applied = append(applied, map[string]interface{}{"id": it.ID, "name": it.Name, "status": "updated"})
		} else {
			if err := s.repo.Insert(ctx, tx, it.ID, it.Name); err != nil {
				tx.Rollback()
				return err
			}
			applied = append(applied, map[string]interface{}{"id": it.ID, "name": it.Name, "status": "created"})
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
