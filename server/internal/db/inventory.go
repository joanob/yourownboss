package db

import "time"

// Resource represents a resource type in the game
type Resource struct {
	ID       int64
	Name     string
	Price    int64 // Price in thousandths for pack_size units
	PackSize int64 // Number of units per pack
}

// CompanyInventory represents the quantity of a resource owned by a company
type CompanyInventory struct {
	ID         int64
	CompanyID  int64
	ResourceID int64
	Quantity   int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// InventoryWithDetails combines inventory and resource details
type InventoryWithDetails struct {
	ID         int64
	ResourceID int64
	Name       string
	Quantity   int64
	Price      int64 // price per pack in thousandths
	PackSize   int64 // units per pack
}
