package models

import (
	"time"

	"github.com/joanob/yourownboss/internal/db/dbqueries"
)

// CompanyProductionBuilding represents a company's instance of a production building
type CompanyProductionBuilding struct {
	ID                   string
	CompanyID            string
	ProductionBuildingID string // DB primary key of the master production building
	Level                int64
	ConstructionEndsAt   *time.Time
	CreatedAt            time.Time
	IsDeleted            bool
	DeletedAt            *time.Time
	ActiveRun            *ProductionRun // loaded separately; nil if no active run
}

// IsUnderConstruction returns true if the building is still being built or upgraded
func (b *CompanyProductionBuilding) IsUnderConstruction() bool {
	if b.ConstructionEndsAt == nil {
		return false
	}
	return b.ConstructionEndsAt.After(time.Now())
}

// IsProducing returns true if the building has an active (non-collected) production run
func (b *CompanyProductionBuilding) IsProducing() bool {
	return b.ActiveRun != nil
}

// IsIdle returns true if the building is available to start production
func (b *CompanyProductionBuilding) IsIdle() bool {
	return !b.IsUnderConstruction() && !b.IsProducing()
}

// CompanyProductionBuildingDTO is the data transfer object for CompanyProductionBuilding
type CompanyProductionBuildingDTO struct {
	ID                   string            `json:"id"`
	CompanyID            string            `json:"company_id"`
	ProductionBuildingID string            `json:"production_building_id"`
	Level                int64             `json:"level"`
	ConstructionEndsAt   *time.Time        `json:"construction_ends_at,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	Status               string            `json:"status"` // "constructing", "idle", "producing"
	ActiveRun            *ProductionRunDTO `json:"active_run,omitempty"`
}

// ToDTO converts a CompanyProductionBuilding to its DTO
func (b *CompanyProductionBuilding) ToDTO() *CompanyProductionBuildingDTO {
	status := "idle"
	if b.IsUnderConstruction() {
		status = "constructing"
	} else if b.IsProducing() {
		status = "producing"
	}

	dto := &CompanyProductionBuildingDTO{
		ID:                   b.ID,
		CompanyID:            b.CompanyID,
		ProductionBuildingID: b.ProductionBuildingID,
		Level:                b.Level,
		ConstructionEndsAt:   b.ConstructionEndsAt,
		CreatedAt:            b.CreatedAt,
		Status:               status,
	}

	if b.ActiveRun != nil {
		dto.ActiveRun = b.ActiveRun.ToDTO()
	}

	return dto
}

// FromDBCompanyProductionBuilding converts a sqlc CompanyProductionBuilding to the domain model
func FromDBCompanyProductionBuilding(dbo *dbqueries.CompanyProductionBuilding) *CompanyProductionBuilding {
	if dbo == nil {
		return nil
	}
	return &CompanyProductionBuilding{
		ID:                   dbo.ID,
		CompanyID:            dbo.CompanyID,
		ProductionBuildingID: dbo.ProductionBuildingID,
		Level:                dbo.Level,
		ConstructionEndsAt:   dbo.ConstructionEndsAt,
		CreatedAt:            dbo.CreatedAt,
		IsDeleted:            dbo.IsDeleted != 0,
		DeletedAt:            dbo.DeletedAt,
	}
}
