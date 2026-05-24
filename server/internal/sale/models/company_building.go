package models

import (
	"time"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// CompanySaleBuilding represents a company's owned sale building in the domain
type CompanySaleBuilding struct {
	ID                 string
	CompanyID          string
	SaleBuildingID     string
	Level              int64
	ConstructionEndsAt *time.Time
	CreatedAt          time.Time
	IsDeleted          bool
	DeletedAt          *time.Time
	ActiveRun          *SaleRun
}

// IsUnderConstruction returns true if the building is still being constructed
func (b *CompanySaleBuilding) IsUnderConstruction() bool {
	if b.ConstructionEndsAt == nil {
		return false
	}
	return time.Now().Before(*b.ConstructionEndsAt)
}

// IsSelling returns true if the building has an active sale run
func (b *CompanySaleBuilding) IsSelling() bool {
	return b.ActiveRun != nil
}

// IsIdle returns true if the building is ready to start a sale
func (b *CompanySaleBuilding) IsIdle() bool {
	return !b.IsUnderConstruction() && !b.IsSelling()
}

// CompanySaleBuildingDTO is the data transfer object for API responses
type CompanySaleBuildingDTO struct {
	ID                 string      `json:"id"`
	CompanyID          string      `json:"company_id"`
	SaleBuildingID     string      `json:"sale_building_id"`
	Level              int64       `json:"level"`
	ConstructionEndsAt *time.Time  `json:"construction_ends_at"`
	CreatedAt          time.Time   `json:"created_at"`
	Status             string      `json:"status"`
	ActiveRun          *SaleRunDTO `json:"active_run,omitempty"`
}

// ToDTO converts a CompanySaleBuilding to its DTO representation
func (b *CompanySaleBuilding) ToDTO() *CompanySaleBuildingDTO {
	dto := &CompanySaleBuildingDTO{
		ID:                 b.ID,
		CompanyID:          b.CompanyID,
		SaleBuildingID:     b.SaleBuildingID,
		Level:              b.Level,
		ConstructionEndsAt: b.ConstructionEndsAt,
		CreatedAt:          b.CreatedAt,
	}

	switch {
	case b.IsUnderConstruction():
		dto.Status = "constructing"
	case b.IsSelling():
		dto.Status = "selling"
		dto.ActiveRun = b.ActiveRun.ToDTO()
	default:
		dto.Status = "idle"
	}

	return dto
}

// FromDBCompanySaleBuilding converts a dbqueries.CompanySaleBuilding to the domain model
func FromDBCompanySaleBuilding(dbo *dbqueries.CompanySaleBuilding) *CompanySaleBuilding {
	return &CompanySaleBuilding{
		ID:                 dbo.ID,
		CompanyID:          dbo.CompanyID,
		SaleBuildingID:     dbo.SaleBuildingID,
		Level:              dbo.Level,
		ConstructionEndsAt: dbo.ConstructionEndsAt,
		CreatedAt:          dbo.CreatedAt,
		IsDeleted:          dbo.IsDeleted != 0,
		DeletedAt:          dbo.DeletedAt,
	}
}
