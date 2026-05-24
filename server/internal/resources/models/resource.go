package models

import "github.com/joanob/yourownboss/internal/db/dbqueries"

// ResourceDBO represents a resource as stored in the database
type ResourceDBO struct {
	ID            string `db:"id"`
	MasterID      string `db:"master_id"`
	Name          string `db:"name"`
	MarketPrice   int64  `db:"market_price"`
	MarketSaleQty int64  `db:"market_sale_qty"`
}

// Resource represents a resource in the domain
type Resource struct {
	ID            string `json:"id"`
	MasterID      string `json:"master_id"`
	Name          string `json:"name"`
	MarketPrice   int64  `json:"market_price"`
	MarketSaleQty int64  `json:"market_sale_qty"`
}

// NewResource creates a new resource from a DBO
func NewResource(dbo *ResourceDBO) *Resource {
	if dbo == nil {
		return nil
	}
	return &Resource{
		ID:            dbo.ID,
		MasterID:      dbo.MasterID,
		Name:          dbo.Name,
		MarketPrice:   dbo.MarketPrice,
		MarketSaleQty: dbo.MarketSaleQty,
	}
}

// ToResourceDBO converts a Resource to a DBO
func (r *Resource) ToResourceDBO() *ResourceDBO {
	return &ResourceDBO{
		ID:            r.ID,
		MasterID:      r.MasterID,
		Name:          r.Name,
		MarketPrice:   r.MarketPrice,
		MarketSaleQty: r.MarketSaleQty,
	}
}

// FromDBQueriesResource converts a sqlc-generated Resource to a model DBO
func FromDBQueriesResource(dbo *dbqueries.Resource) *ResourceDBO {
	if dbo == nil {
		return nil
	}
	return &ResourceDBO{
		ID:            dbo.ID,
		MasterID:      dbo.MasterID,
		Name:          dbo.Name,
		MarketPrice:   dbo.MarketPrice,
		MarketSaleQty: dbo.MarketSaleQty,
	}
}
