package models

import "github.com/joanob/yourownboss/internal/db/dbqueries"

// SaleBuildingDBO represents a sale building as stored in the database
type SaleBuildingDBO struct {
	ID                  string `db:"id"`
	MasterID            string `db:"master_id"`
	Name                string `db:"name"`
	ConstructionCost    int64  `db:"construction_cost"`
	ConstructionTimeSec int64  `db:"construction_time_s"`
}

// SaleResource represents a resource that can be sold in a sale building
type SaleResource struct {
	ResourceID         string `json:"resource_id"`
	PricePerUnit       int64  `json:"price_per_unit"`
	UnitsSoldPerSecond int64  `json:"units_sold_per_second"`
}

// SaleBuilding represents a sale building in the domain
type SaleBuilding struct {
	ID                  string          `json:"id"`
	MasterID            string          `json:"master_id"`
	Name                string          `json:"name"`
	ConstructionCost    int64           `json:"construction_cost"`
	ConstructionTimeSec int64           `json:"construction_time_s"`
	Resources           []*SaleResource `json:"resources"`
}

// NewSaleBuilding creates a new sale building from a DBO
func NewSaleBuilding(dbo *SaleBuildingDBO) *SaleBuilding {
	if dbo == nil {
		return nil
	}
	return &SaleBuilding{
		ID:                  dbo.ID,
		MasterID:            dbo.MasterID,
		Name:                dbo.Name,
		ConstructionCost:    dbo.ConstructionCost,
		ConstructionTimeSec: dbo.ConstructionTimeSec,
		Resources:           make([]*SaleResource, 0),
	}
}

// ToSaleBuildingDBO converts a SaleBuilding to a DBO
func (sb *SaleBuilding) ToSaleBuildingDBO() *SaleBuildingDBO {
	return &SaleBuildingDBO{
		ID:                  sb.ID,
		MasterID:            sb.MasterID,
		Name:                sb.Name,
		ConstructionCost:    sb.ConstructionCost,
		ConstructionTimeSec: sb.ConstructionTimeSec,
	}
}

// FromDBQueriesSaleBuilding converts a sqlc-generated SaleBuilding to a model DBO
func FromDBQueriesSaleBuilding(dbo *dbqueries.SaleBuilding) *SaleBuildingDBO {
	if dbo == nil {
		return nil
	}
	return &SaleBuildingDBO{
		ID:                  dbo.ID,
		MasterID:            dbo.MasterID,
		Name:                dbo.Name,
		ConstructionCost:    dbo.ConstructionCost,
		ConstructionTimeSec: dbo.ConstructionTimeS,
	}
}

// AddResource adds a resource that can be sold in this building
func (sb *SaleBuilding) AddResource(resourceID string, pricePerUnit, unitsSoldPerSecond int64) {
	if sb.Resources == nil {
		sb.Resources = make([]*SaleResource, 0)
	}
	sb.Resources = append(sb.Resources, &SaleResource{
		ResourceID:         resourceID,
		PricePerUnit:       pricePerUnit,
		UnitsSoldPerSecond: unitsSoldPerSecond,
	})
}
