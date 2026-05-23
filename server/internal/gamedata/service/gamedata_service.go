package service

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/gamedata/repository"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// GamedataService maneja la carga de datos maestros desde BD.
// La cache se sincroniza desde fuera mediante RefreshCache()
type GamedataService struct {
	db                     *sql.DB
	fileRepo               *repository.GamedataRepository
	resourceRepo           *repository.ResourceRepo
	productionBuildingRepo *repository.ProductionBuildingRepo
	saleBuildingRepo       *repository.SaleBuildingRepo
}

// NewGamedataService crea una nueva instancia del servicio
// Requiere: filePath (para importar de JSON), db (conexión a BD)
func NewGamedataService(filePath string, db *sql.DB) *GamedataService {
	return &GamedataService{
		db:                     db,
		fileRepo:               repository.NewGamedataRepository(filePath),
		resourceRepo:           repository.NewResourceRepo(db),
		productionBuildingRepo: repository.NewProductionBuildingRepo(db),
		saleBuildingRepo:       repository.NewSaleBuildingRepo(db),
	}
}

// Load asegura que la BD tiene datos maestros.
// Si BD está vacía, importa desde JSON.
// NO sincroniza el cache (ver RefreshCache()).
func (gs *GamedataService) Load() error {
	logger := log.Logger

	logger.Info().Msg("Verificando datos maestros en BD...")

	// Contar datos en BD
	resourceCount, err := gs.resourceRepo.Count()
	if err != nil {
		return fmt.Errorf("error contando recursos en BD: %w", err)
	}

	// Si BD está vacía, importar desde JSON
	if resourceCount == 0 {
		logger.Info().Msg("BD vacía, importando datos desde JSON...")
		if err := gs.importFromFile(); err != nil {
			return fmt.Errorf("error importando datos: %w", err)
		}
	} else {
		logger.Info().Int("resources_in_db", resourceCount).Msg("Datos maestros encontrados en BD")
	}

	logger.Info().Msg("Verificación de datos maestros completada")
	return nil
}

// importFromFile carga datos desde JSON e importa a BD
func (gs *GamedataService) importFromFile() error {
	logger := log.Logger

	// Cargar desde JSON
	gamedata, err := gs.fileRepo.LoadFromFile()
	if err != nil {
		return fmt.Errorf("error cargando JSON: %w", err)
	}

	// Insertar en BD (en transacciones independientes)
	if err := gs.resourceRepo.InsertBatch(gamedata.Resources); err != nil {
		return fmt.Errorf("error insertando recursos: %w", err)
	}
	logger.Info().Int("count", len(gamedata.Resources)).Msg("Recursos importados a BD")

	if err := gs.productionBuildingRepo.InsertBatch(gamedata.ProductionBuildings); err != nil {
		return fmt.Errorf("error insertando edificios de producción: %w", err)
	}
	logger.Info().Int("count", len(gamedata.ProductionBuildings)).Msg("Edificios de producción importados a BD")

	if err := gs.saleBuildingRepo.InsertBatch(gamedata.SaleBuildings); err != nil {
		return fmt.Errorf("error insertando edificios de venta: %w", err)
	}
	logger.Info().Int("count", len(gamedata.SaleBuildings)).Msg("Edificios de venta importados a BD")

	return nil
}

// RefreshCache sincroniza el cache desde los datos en BD
// Requiere la cache como parámetro para actualizarla
func (gs *GamedataService) RefreshCache(gameCache *cache.GamedataCache) error {
	logger := log.Logger

	logger.Info().Msg("Sincronizando cache desde BD...")

	// Cargar recursos desde BD
	resources, err := gs.resourceRepo.GetAll()
	if err != nil {
		return fmt.Errorf("error cargando recursos: %w", err)
	}

	// Cargar edificios de producción desde BD
	productionBuildings, err := gs.productionBuildingRepo.GetAll()
	if err != nil {
		return fmt.Errorf("error cargando edificios de producción: %w", err)
	}

	// Cargar edificios de venta desde BD
	saleBuildings, err := gs.saleBuildingRepo.GetAll()
	if err != nil {
		return fmt.Errorf("error cargando edificios de venta: %w", err)
	}

	// Actualizar cache
	gameCache.SetResources(resources)
	gameCache.SetProductionBuildings(productionBuildings)
	gameCache.SetSaleBuildings(saleBuildings)

	logger.Info().
		Int("resources", len(resources)).
		Int("production_buildings", len(productionBuildings)).
		Int("sale_buildings", len(saleBuildings)).
		Msg("Cache sincronizado exitosamente desde BD")

	return nil
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
