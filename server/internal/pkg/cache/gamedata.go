package cache

import (
	"sync"
)

// Resource representa un recurso del juego
type Resource struct {
	ID            string `json:"id"`
	MasterID      string `json:"master_id"`
	Name          string `json:"name"`
	MarketPrice   int64  `json:"market_price"`
	MarketSaleQty int64  `json:"market_sale_qty"`
}

// ProductionProcess representa un proceso de producción
type ProductionProcess struct {
	ID                   string                      `json:"id"`
	MasterID             string                      `json:"master_id"`
	ProductionBuildingID string                      `json:"production_building_id"`
	Name                 string                      `json:"name"`
	CycleTimeS           int64                       `json:"cycle_time_s"`
	WindowStartHour      *int64                      `json:"window_start_hour"`
	WindowEndHour        *int64                      `json:"window_end_hour"`
	Resources            []ProductionProcessResource `json:"resources"`
}

// ProductionProcessResource representa un recurso en un proceso de producción
type ProductionProcessResource struct {
	ResourceID string `json:"resource_id"`
	IsOutput   bool   `json:"is_output"`
	Quantity   int64  `json:"quantity"`
}

// ProductionBuilding representa un edificio de producción
type ProductionBuilding struct {
	ID                string              `json:"id"`
	MasterID          string              `json:"master_id"`
	Name              string              `json:"name"`
	ConstructionCost  int64               `json:"construction_cost"`
	ConstructionTimeS int64               `json:"construction_time_s"`
	Processes         []ProductionProcess `json:"processes"`
}

// SaleResource representa un recurso que se puede vender
type SaleResource struct {
	ResourceID         string `json:"resource_id"`
	PricePerUnit       int64  `json:"price_per_unit"`
	UnitsSoldPerSecond int64  `json:"units_sold_per_second"`
}

// SaleBuilding representa un edificio de venta
type SaleBuilding struct {
	ID                string         `json:"id"`
	MasterID          string         `json:"master_id"`
	Name              string         `json:"name"`
	ConstructionCost  int64          `json:"construction_cost"`
	ConstructionTimeS int64          `json:"construction_time_s"`
	Resources         []SaleResource `json:"resources"`
}

// GamedataCache almacena datos maestros del juego en memoria
type GamedataCache struct {
	resources           map[string]Resource
	productionBuildings map[string]ProductionBuilding
	saleBuildings       map[string]SaleBuilding
	mu                  sync.RWMutex
}

// NewGamedataCache crea una nueva instancia del cache de gamedata
func NewGamedataCache() *GamedataCache {
	return &GamedataCache{
		resources:           make(map[string]Resource),
		productionBuildings: make(map[string]ProductionBuilding),
		saleBuildings:       make(map[string]SaleBuilding),
	}
}

// SetResources establece los recursos en el cache
func (gc *GamedataCache) SetResources(resources []Resource) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.resources = make(map[string]Resource)
	for _, r := range resources {
		gc.resources[r.MasterID] = r
	}
}

// SetProductionBuildings establece los edificios de producción en el cache
func (gc *GamedataCache) SetProductionBuildings(buildings []ProductionBuilding) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.productionBuildings = make(map[string]ProductionBuilding)
	for _, b := range buildings {
		gc.productionBuildings[b.MasterID] = b
	}
}

// SetSaleBuildings establece los edificios de venta en el cache
func (gc *GamedataCache) SetSaleBuildings(buildings []SaleBuilding) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.saleBuildings = make(map[string]SaleBuilding)
	for _, b := range buildings {
		gc.saleBuildings[b.MasterID] = b
	}
}

// GetResource obtiene un recurso por master_id
func (gc *GamedataCache) GetResource(masterID string) (Resource, bool) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	resource, exists := gc.resources[masterID]
	return resource, exists
}

// GetProductionBuilding obtiene un edificio de producción por master_id
func (gc *GamedataCache) GetProductionBuilding(masterID string) (ProductionBuilding, bool) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	building, exists := gc.productionBuildings[masterID]
	return building, exists
}

// GetSaleBuilding obtiene un edificio de venta por master_id
func (gc *GamedataCache) GetSaleBuilding(masterID string) (SaleBuilding, bool) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	building, exists := gc.saleBuildings[masterID]
	return building, exists
}

// GetAllResources devuelve todos los recursos
func (gc *GamedataCache) GetAllResources() []Resource {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	resources := make([]Resource, 0, len(gc.resources))
	for _, r := range gc.resources {
		resources = append(resources, r)
	}
	return resources
}

// GetAllProductionBuildings devuelve todos los edificios de producción
func (gc *GamedataCache) GetAllProductionBuildings() []ProductionBuilding {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	buildings := make([]ProductionBuilding, 0, len(gc.productionBuildings))
	for _, b := range gc.productionBuildings {
		buildings = append(buildings, b)
	}
	return buildings
}

// GetAllSaleBuildings devuelve todos los edificios de venta
func (gc *GamedataCache) GetAllSaleBuildings() []SaleBuilding {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	buildings := make([]SaleBuilding, 0, len(gc.saleBuildings))
	for _, b := range gc.saleBuildings {
		buildings = append(buildings, b)
	}
	return buildings
}

// IsLoaded devuelve true si el cache tiene datos cargados
func (gc *GamedataCache) IsLoaded() bool {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	return len(gc.resources) > 0 && len(gc.productionBuildings) > 0 && len(gc.saleBuildings) > 0
}
