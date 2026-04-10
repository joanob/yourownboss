package production

import (
	"context"
	"database/sql"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// CreateBuilding inserts building and nested processes/resources
func (s *Service) CreateBuilding(ctx context.Context, db *sql.DB, dto BuildingDTO) error {
	if dto.ID == 0 {
		return errors.New("building id required and must be non-zero")
	}
	if dto.Name == "" {
		return errors.New("building name required")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// check exists
	exists, err := s.repo.BuildingExists(ctx, tx, dto.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if exists {
		tx.Rollback()
		return errors.New("building id already exists")
	}

	if err := s.repo.InsertBuilding(ctx, tx, dto.ID, dto.Name); err != nil {
		tx.Rollback()
		return err
	}

	// processes
	for _, p := range dto.Processes {
		if p.ID == nil {
			tx.Rollback()
			return errors.New("process id required and must be non-zero")
		}
		if p.Name == "" {
			tx.Rollback()
			return errors.New("process name required")
		}
		if err := s.repo.InsertProcess(ctx, tx, *p.ID, dto.ID, p.Name, p.StartHour, p.EndHour); err != nil {
			tx.Rollback()
			return err
		}

		// resources: insert after deleting any existing (fresh create no existing)
		if len(p.Resources) > 0 {
			for _, pr := range p.Resources {
				if pr.ResourceID == 0 {
					tx.Rollback()
					return errors.New("process resource resource_id required")
				}
				if err := s.repo.InsertProcessResource(ctx, tx, *p.ID, pr.ResourceID, pr.IsOutput); err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// UpdateBuilding replaces nested collections: processes not in payload are deleted
func (s *Service) UpdateBuilding(ctx context.Context, db *sql.DB, buildingID int64, dto BuildingDTO) error {
	if dto.Name == "" {
		return errors.New("building name required")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	exists, err := s.repo.BuildingExists(ctx, tx, buildingID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !exists {
		tx.Rollback()
		return errors.New("building not found")
	}

	if err := s.repo.UpdateBuilding(ctx, tx, buildingID, dto.Name); err != nil {
		tx.Rollback()
		return err
	}

	// process handling: collect keep IDs and upsert each process; delete others
	keep := make([]int64, 0, len(dto.Processes))
	for _, p := range dto.Processes {
		if p.ID == nil {
			tx.Rollback()
			return errors.New("process id required for update")
		}
		pid := *p.ID
		keep = append(keep, pid)
		// check if process exists by trying update; if not exists, insert
		if err := s.repo.UpdateProcess(ctx, tx, pid, p.Name, p.StartHour, p.EndHour); err != nil {
			// try insert
			if err := s.repo.InsertProcess(ctx, tx, pid, buildingID, p.Name, p.StartHour, p.EndHour); err != nil {
				tx.Rollback()
				return err
			}
		}

		// replace process resources: delete then insert
		if err := s.repo.DeleteProcessResources(ctx, tx, pid); err != nil {
			tx.Rollback()
			return err
		}
		for _, pr := range p.Resources {
			if pr.ResourceID == 0 {
				tx.Rollback()
				return errors.New("process resource resource_id required")
			}
			if err := s.repo.InsertProcessResource(ctx, tx, pid, pr.ResourceID, pr.IsOutput); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if err := s.repo.DeleteProcessesNotIn(ctx, tx, buildingID, keep); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (s *Service) DeleteBuilding(ctx context.Context, db *sql.DB, id int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteBuilding(ctx, tx, id); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
