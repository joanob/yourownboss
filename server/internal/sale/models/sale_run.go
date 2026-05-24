package models

import (
	"time"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// SaleRun represents an active or completed sale run in the domain
type SaleRun struct {
	ID                    string
	CompanySaleBuildingID string
	ResourceID            string
	UnitsToSell           int64
	StartedAt             time.Time
	EndsAt                time.Time
	IsCollected           bool
	CollectedAt           *time.Time
	IsDeleted             bool
	DeletedAt             *time.Time
}

// IsCompleted returns true if the sale run has finished
func (r *SaleRun) IsCompleted() bool {
	return !time.Now().Before(r.EndsAt)
}

// SaleRunDTO is the data transfer object for API responses
type SaleRunDTO struct {
	ID                    string     `json:"id"`
	CompanySaleBuildingID string     `json:"company_sale_building_id"`
	ResourceID            string     `json:"resource_id"`
	UnitsToSell           int64      `json:"units_to_sell"`
	StartedAt             time.Time  `json:"started_at"`
	EndsAt                time.Time  `json:"ends_at"`
	IsCollected           bool       `json:"is_collected"`
	CollectedAt           *time.Time `json:"collected_at,omitempty"`
	IsCompleted           bool       `json:"is_completed"`
}

// ToDTO converts a SaleRun to its DTO representation
func (r *SaleRun) ToDTO() *SaleRunDTO {
	return &SaleRunDTO{
		ID:                    r.ID,
		CompanySaleBuildingID: r.CompanySaleBuildingID,
		ResourceID:            r.ResourceID,
		UnitsToSell:           r.UnitsToSell,
		StartedAt:             r.StartedAt,
		EndsAt:                r.EndsAt,
		IsCollected:           r.IsCollected,
		CollectedAt:           r.CollectedAt,
		IsCompleted:           r.IsCompleted(),
	}
}

// FromDBSaleRun converts a dbqueries.SaleRun to the domain model
func FromDBSaleRun(dbo *dbqueries.SaleRun) *SaleRun {
	return &SaleRun{
		ID:                    dbo.ID,
		CompanySaleBuildingID: dbo.CompanySaleBuildingID,
		ResourceID:            dbo.ResourceID,
		UnitsToSell:           dbo.UnitsToSell,
		StartedAt:             dbo.StartedAt,
		EndsAt:                dbo.EndsAt,
		IsCollected:           dbo.IsCollected != 0,
		CollectedAt:           dbo.CollectedAt,
		IsDeleted:             dbo.IsDeleted != 0,
		DeletedAt:             dbo.DeletedAt,
	}
}
