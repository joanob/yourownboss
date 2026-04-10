package simulation

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// StartSimulation launches the simulation in background and returns immediately.
// It generates combinations for the requested time and resource ranges, computes
// benefit per hour and persists positive results concurrently using a worker pool.
func (s *Service) StartSimulation(ctx context.Context, db *sql.DB, req SimulationRequest) error {
	go func(r SimulationRequest) {
		start := time.Now()
		log.Printf("simulation started: process=%d time=%d-%d step=%d", r.ProcessID, r.TimeMinMs, r.TimeMaxMs, r.TimeStepMs)

		if r.TimeStepMs <= 0 || r.TimeMinMs < 0 || r.TimeMaxMs < r.TimeMinMs {
			log.Printf("invalid time range")
			return
		}

		pres, err := s.repo.GetProcessResources(context.Background(), r.ProcessID)
		if err != nil {
			log.Printf("failed to get process resources: %v", err)
			return
		}
		if len(pres) == 0 {
			log.Printf("no resources defined for process %d", r.ProcessID)
			return
		}

		presMap := make(map[int64]bool)
		for _, p := range pres {
			presMap[p.ResourceID] = p.IsOutput
		}

		rrMap := make(map[int64]SimulationResourceRange)
		for _, rr := range r.ResourceRanges {
			rrMap[rr.ResourceID] = rr
		}
		for _, p := range pres {
			if _, ok := rrMap[p.ResourceID]; !ok {
				log.Printf("missing resource range for resource %d", p.ResourceID)
				return
			}
		}

		priceLists := make([][]int64, len(pres))
		resIDs := make([]int64, len(pres))
		for i, p := range pres {
			resIDs[i] = p.ResourceID
			rr := rrMap[p.ResourceID]
			if rr.Step <= 0 || rr.MaxPrice < rr.MinPrice {
				log.Printf("invalid price range for resource %d", p.ResourceID)
				return
			}
			steps := 0
			if rr.Step > 0 {
				steps = int((rr.MaxPrice - rr.MinPrice) / rr.Step)
				if steps < 0 {
					steps = 0
				}
			}
			list := make([]int64, 0, steps+1)
			for si := 0; si <= steps; si++ {
				v := rr.MinPrice + int64(si)*rr.Step
				if v > rr.MaxPrice {
					v = rr.MaxPrice
				}
				list = append(list, v)
			}
			priceLists[i] = list
		}

		workers := 2
		if env := os.Getenv("SIM_WORKERS"); env != "" {
			if n, e := strconv.Atoi(env); e == nil && n > 0 {
				workers = n
			}
		}

		type job struct {
			timeMs int
			prices []int64
		}

		jobs := make(chan job, workers*2)
		var wg sync.WaitGroup

		// output channel for completed positive results to be persisted in batches
		type resResource struct {
			ResourceID int64
			IsOutput   bool
			Price      int64
			Quantity   int64
		}
		type resItem struct {
			TimeMs      int
			BenefitHour float64
			Resources   []resResource
		}

		out := make(chan resItem, workers*4)

		// saver goroutine: persist batches of results
		batchSize := 1000
		if benv := os.Getenv("SIM_BATCH_SIZE"); benv != "" {
			if n, e := strconv.Atoi(benv); e == nil && n > 0 {
				batchSize = n
			}
		}

		var saverWg sync.WaitGroup
		saverWg.Add(1)
		go func() {
			defer saverWg.Done()
			batch := make([]resItem, 0, batchSize)
			persist := func(items []resItem) {
				if len(items) == 0 {
					return
				}
				tx, err := db.BeginTx(context.Background(), nil)
				if err != nil {
					log.Printf("batch begin tx error: %v", err)
					return
				}
				for _, it := range items {
					sid, err := s.repo.SaveSimulationResult(context.Background(), tx, r.ProcessID, it.TimeMs, it.BenefitHour)
					if err != nil {
						log.Printf("batch save simulation error: %v", err)
						_ = tx.Rollback()
						return
					}
					for _, rr := range it.Resources {
						if err := s.repo.SaveSimulationResource(context.Background(), tx, sid, rr.ResourceID, rr.IsOutput, rr.Price, rr.Quantity); err != nil {
							log.Printf("batch save simulation resource error: %v", err)
							_ = tx.Rollback()
							return
						}
					}
				}
				if err := tx.Commit(); err != nil {
					log.Printf("batch commit error: %v", err)
					_ = tx.Rollback()
				}
			}

			for it := range out {
				batch = append(batch, it)
				if len(batch) >= batchSize {
					persist(batch)
					batch = batch[:0]
				}
			}
			if len(batch) > 0 {
				persist(batch)
			}
		}()

		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range jobs {
					var outputsSum float64
					var inputsSum float64
					for idx, price := range j.prices {
						rid := resIDs[idx]
						if presMap[rid] {
							outputsSum += float64(price)
						} else {
							inputsSum += float64(price)
						}
					}
					profitPerCycle := outputsSum - inputsSum
					if profitPerCycle <= 0 {
						continue
					}

					cyclesPerHour := float64(3600000) / float64(j.timeMs)
					benefitPerHour := profitPerCycle * cyclesPerHour
					// send to saver (use quantity=1 by default)
					resList := make([]resResource, 0, len(j.prices))
					for idx, price := range j.prices {
						rid := resIDs[idx]
						resList = append(resList, resResource{ResourceID: rid, IsOutput: presMap[rid], Price: price, Quantity: 1})
					}
					select {
					case out <- resItem{TimeMs: j.timeMs, BenefitHour: benefitPerHour, Resources: resList}:
					default:
						// if saver is busy, block briefly (backpressure)
						out <- resItem{TimeMs: j.timeMs, BenefitHour: benefitPerHour, Resources: resList}
					}
				}
			}()
		}

		var gen func(idx int, cur []int64)
		gen = func(idx int, cur []int64) {
			if idx == len(priceLists) {
				for t := r.TimeMinMs; t <= r.TimeMaxMs; t += r.TimeStepMs {
					prices := make([]int64, len(cur))
					copy(prices, cur)
					jobs <- job{timeMs: t, prices: prices}
				}
				return
			}
			for _, v := range priceLists[idx] {
				gen(idx+1, append(cur, v))
			}
		}

		gen(0, []int64{})
		close(jobs)
		wg.Wait()
		close(out)
		saverWg.Wait()

		log.Printf("simulation finished: process=%d duration=%s", r.ProcessID, time.Since(start))
	}(req)
	return nil
}
