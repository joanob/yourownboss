package simulation

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// StartSimulation launches the simulation in background and returns immediately.
// Currently it performs a minimal placeholder work: logs start/finish and
// writes a single positive sample result to the DB so the pipeline is exercised.
func (s *Service) StartSimulation(ctx context.Context, db *sql.DB, req SimulationRequest) error {
	go func(r SimulationRequest) {
		start := time.Now()
		log.Printf("simulation started: process=%d time=%d-%d step=%d", r.ProcessID, r.TimeMinMs, r.TimeMaxMs, r.TimeStepMs)

		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			log.Printf("simulation begin tx error: %v", err)
			return
		}

		// Placeholder calculation: simple positive benefit so we persist one result.
		benefit := 1.0
		if benefit > 0 {
			sid, err := s.repo.SaveSimulationResult(context.Background(), tx, r.ProcessID, r.TimeMinMs, benefit)
			if err != nil {
				log.Printf("save simulation error: %v", err)
				_ = tx.Rollback()
				return
			}
			// Save resource snapshot using min price as a simple example
			for _, rr := range r.ResourceRanges {
				if err := s.repo.SaveSimulationResource(context.Background(), tx, sid, rr.ResourceID, rr.IsOutput, rr.MinPrice); err != nil {
					log.Printf("save simulation resource error: %v", err)
					_ = tx.Rollback()
					return
				}
			}
		}

		if err := tx.Commit(); err != nil {
			log.Printf("simulation commit error: %v", err)
			return
		}

		log.Printf("simulation finished: process=%d duration=%s", r.ProcessID, time.Since(start))
	}(req)
	return nil
}
