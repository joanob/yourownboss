package models

import (
	"time"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// ProductionRun represents a single production run for a company building
type ProductionRun struct {
	ID                string
	CompanyBuildingID string
	ProcessID         string
	ProductionCycles  int64
	StartedAt         time.Time
	EndsAt            time.Time
	IsCollected       bool
	CollectedAt       *time.Time
	IsDeleted         bool
	DeletedAt         *time.Time
}

// IsCompleted returns true if the production run has finished
func (r *ProductionRun) IsCompleted() bool {
	return !time.Now().Before(r.EndsAt)
}

// ProductionRunDTO is the data transfer object for ProductionRun
type ProductionRunDTO struct {
	ID                string     `json:"id"`
	CompanyBuildingID string     `json:"company_building_id"`
	ProcessID         string     `json:"process_id"`
	ProductionCycles  int64      `json:"production_cycles"`
	StartedAt         time.Time  `json:"started_at"`
	EndsAt            time.Time  `json:"ends_at"`
	IsCollected       bool       `json:"is_collected"`
	CollectedAt       *time.Time `json:"collected_at,omitempty"`
	IsCompleted       bool       `json:"is_completed"`
}

// ToDTO converts a ProductionRun to its DTO
func (r *ProductionRun) ToDTO() *ProductionRunDTO {
	return &ProductionRunDTO{
		ID:                r.ID,
		CompanyBuildingID: r.CompanyBuildingID,
		ProcessID:         r.ProcessID,
		ProductionCycles:  r.ProductionCycles,
		StartedAt:         r.StartedAt,
		EndsAt:            r.EndsAt,
		IsCollected:       r.IsCollected,
		CollectedAt:       r.CollectedAt,
		IsCompleted:       r.IsCompleted(),
	}
}

// FromDBProductionRun converts a sqlc ProductionRun to the domain model
func FromDBProductionRun(dbo *dbqueries.ProductionRun) *ProductionRun {
	if dbo == nil {
		return nil
	}
	return &ProductionRun{
		ID:                dbo.ID,
		CompanyBuildingID: dbo.CompanyBuildingID,
		ProcessID:         dbo.ProcessID,
		ProductionCycles:  dbo.ProductionCycles,
		StartedAt:         dbo.StartedAt,
		EndsAt:            dbo.EndsAt,
		IsCollected:       dbo.IsCollected != 0,
		CollectedAt:       dbo.CollectedAt,
		IsDeleted:         dbo.IsDeleted != 0,
		DeletedAt:         dbo.DeletedAt,
	}
}
