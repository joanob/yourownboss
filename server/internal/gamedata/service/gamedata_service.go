package service

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/gamedata/repository"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// GamedataService maneja la carga y caché de datos maestros
type GamedataService struct {
	repo  *repository.GamedataRepository
	cache *cache.GamedataCache
}

// NewGamedataService crea una nueva instancia del servicio
func NewGamedataService(filePath string, cache *cache.GamedataCache) *GamedataService {
	return &GamedataService{
		repo:  repository.NewGamedataRepository(filePath),
		cache: cache,
	}
}

// Load carga los datos maestros desde archivo y los almacena en el cache
func (gs *GamedataService) Load() error {
	logger := log.Logger

	logger.Info().Msg("Cargando datos maestros...")

	// Cargar desde archivo
	gamedata, err := gs.repo.LoadFromFile()
	if err != nil {
		return fmt.Errorf("error al cargar gamedata: %w", err)
	}

	// Almacenar en cache
	gs.cache.SetResources(gamedata.Resources)
	gs.cache.SetProductionBuildings(gamedata.ProductionBuildings)
	gs.cache.SetSaleBuildings(gamedata.SaleBuildings)

	// Log resumen
	logger.Info().
		Int("resources", len(gamedata.Resources)).
		Int("production_buildings", len(gamedata.ProductionBuildings)).
		Int("sale_buildings", len(gamedata.SaleBuildings)).
		Msg("Gamedata cargado exitosamente")

	return nil
}

// IsLoaded verifica si el cache tiene datos cargados
func (gs *GamedataService) IsLoaded() bool {
	return gs.cache.IsLoaded()
}

// GetAllResources devuelve todos los recursos del cache
func (gs *GamedataService) GetAllResources() []cache.Resource {
	return gs.cache.GetAllResources()
}

// GetAllProductionBuildings devuelve todos los edificios de producción del cache
func (gs *GamedataService) GetAllProductionBuildings() []cache.ProductionBuilding {
	return gs.cache.GetAllProductionBuildings()
}

// GetAllSaleBuildings devuelve todos los edificios de venta del cache
func (gs *GamedataService) GetAllSaleBuildings() []cache.SaleBuilding {
	return gs.cache.GetAllSaleBuildings()
}

// GetResource obtiene un recurso por master_id
func (gs *GamedataService) GetResource(masterID string) (cache.Resource, bool) {
	return gs.cache.GetResource(masterID)
}

// GetProductionBuilding obtiene un edificio de producción por master_id
func (gs *GamedataService) GetProductionBuilding(masterID string) (cache.ProductionBuilding, bool) {
	return gs.cache.GetProductionBuilding(masterID)
}

// GetSaleBuilding obtiene un edificio de venta por master_id
func (gs *GamedataService) GetSaleBuilding(masterID string) (cache.SaleBuilding, bool) {
	return gs.cache.GetSaleBuilding(masterID)
}
