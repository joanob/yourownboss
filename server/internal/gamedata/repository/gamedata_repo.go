package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// GamedataRepository maneja la carga de datos maestros
type GamedataRepository struct {
	filePath string
}

// NewGamedataRepository crea una nueva instancia del repository
func NewGamedataRepository(filePath string) *GamedataRepository {
	return &GamedataRepository{
		filePath: filePath,
	}
}

// GamedataDTO es la estructura de datos para deserializar el JSON
type GamedataDTO struct {
	Resources           []cache.Resource           `json:"resources"`
	ProductionBuildings []cache.ProductionBuilding `json:"production_buildings"`
	SaleBuildings       []cache.SaleBuilding       `json:"sale_buildings"`
}

// LoadFromFile carga los datos maestros desde un archivo JSON
func (gr *GamedataRepository) LoadFromFile() (*GamedataDTO, error) {
	logger := log.Logger

	// Leer archivo JSON
	data, err := os.ReadFile(gr.filePath)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer archivo gamedata en %s: %w", gr.filePath, err)
	}

	// Deserializar JSON
	var gamedata GamedataDTO
	if err := json.Unmarshal(data, &gamedata); err != nil {
		return nil, fmt.Errorf("error al parsear gamedata JSON: %w", err)
	}

	// Validar estructura
	if err := gr.validate(&gamedata); err != nil {
		return nil, fmt.Errorf("gamedata no válido: %w", err)
	}

	logger.Info().
		Int("resources", len(gamedata.Resources)).
		Int("production_buildings", len(gamedata.ProductionBuildings)).
		Int("sale_buildings", len(gamedata.SaleBuildings)).
		Msg("Gamedata cargado desde archivo")

	return &gamedata, nil
}

// validate verifica que la estructura de gamedata sea válida
func (gr *GamedataRepository) validate(gamedata *GamedataDTO) error {
	logger := log.Logger

	// Validar que no está vacío
	if len(gamedata.Resources) == 0 {
		return fmt.Errorf("no hay recursos definidos")
	}

	if len(gamedata.ProductionBuildings) == 0 {
		logger.Warn().Msg("No hay edificios de producción definidos")
	}

	if len(gamedata.SaleBuildings) == 0 {
		logger.Warn().Msg("No hay edificios de venta definidos")
	}

	// Validar que master_id son únicos en recursos
	resourceMasterIds := make(map[string]bool)
	for i, r := range gamedata.Resources {
		if r.MasterID == "" {
			return fmt.Errorf("recurso %d: master_id está vacío", i)
		}
		if r.ID == "" {
			return fmt.Errorf("recurso %d (%s): id está vacío", i, r.MasterID)
		}
		if r.Name == "" {
			return fmt.Errorf("recurso %d (%s): name está vacío", i, r.MasterID)
		}
		if r.MarketPrice < 0 {
			return fmt.Errorf("recurso %d (%s): market_price no puede ser negativo", i, r.MasterID)
		}
		if r.MarketSaleQty <= 0 {
			return fmt.Errorf("recurso %d (%s): market_sale_qty debe ser mayor a 0", i, r.MasterID)
		}

		if resourceMasterIds[r.MasterID] {
			return fmt.Errorf("recurso %s: master_id duplicado", r.MasterID)
		}
		resourceMasterIds[r.MasterID] = true
	}

	// Validar edificios de producción
	prodBuildingIds := make(map[string]bool)
	processIds := make(map[string]bool)

	for i, b := range gamedata.ProductionBuildings {
		if b.MasterID == "" {
			return fmt.Errorf("edificio producción %d: master_id está vacío", i)
		}
		if b.ID == "" {
			return fmt.Errorf("edificio producción %d (%s): id está vacío", i, b.MasterID)
		}
		if b.Name == "" {
			return fmt.Errorf("edificio producción %d (%s): name está vacío", i, b.MasterID)
		}

		if prodBuildingIds[b.MasterID] {
			return fmt.Errorf("edificio producción %s: master_id duplicado", b.MasterID)
		}
		prodBuildingIds[b.MasterID] = true

		// Validar procesos
		for j, p := range b.Processes {
			if p.MasterID == "" {
				return fmt.Errorf("proceso %d en edificio %s: master_id está vacío", j, b.MasterID)
			}
			if p.ID == "" {
				return fmt.Errorf("proceso %d (%s) en edificio %s: id está vacío", j, p.MasterID, b.MasterID)
			}
			if p.CycleTimeS <= 0 {
				return fmt.Errorf("proceso %s: cycle_time_s debe ser mayor a 0", p.MasterID)
			}

			// Validar ventana horaria si existe
			if (p.WindowStartHour != nil && p.WindowEndHour == nil) || (p.WindowStartHour == nil && p.WindowEndHour != nil) {
				return fmt.Errorf("proceso %s: window_start_hour y window_end_hour deben estar ambos presentes o ausentes", p.MasterID)
			}
			if p.WindowStartHour != nil && p.WindowEndHour != nil {
				if *p.WindowStartHour < 0 || *p.WindowStartHour > 86400 {
					return fmt.Errorf("proceso %s: window_start_hour debe estar entre 0 y 86400", p.MasterID)
				}
				if *p.WindowEndHour < 0 || *p.WindowEndHour > 86400 {
					return fmt.Errorf("proceso %s: window_end_hour debe estar entre 0 y 86400", p.MasterID)
				}
			}

			if processIds[p.MasterID] {
				return fmt.Errorf("proceso %s: master_id duplicado", p.MasterID)
			}
			processIds[p.MasterID] = true

			// Validar recursos del proceso
			if len(p.Resources) == 0 {
				return fmt.Errorf("proceso %s: debe tener al menos un recurso", p.MasterID)
			}

			hasInput := false
			hasOutput := false
			for _, pr := range p.Resources {
				if pr.ResourceID == "" {
					return fmt.Errorf("proceso %s: resource_id está vacío", p.MasterID)
				}
				if !resourceMasterIds[pr.ResourceID] {
					return fmt.Errorf("proceso %s: resource_id %s no existe", p.MasterID, pr.ResourceID)
				}
				if pr.Quantity <= 0 {
					return fmt.Errorf("proceso %s: quantity debe ser mayor a 0", p.MasterID)
				}

				if pr.IsOutput {
					hasOutput = true
				} else {
					hasInput = true
				}
			}

			if !hasInput {
				return fmt.Errorf("proceso %s: debe tener al menos un recurso de entrada (is_output=false)", p.MasterID)
			}
			if !hasOutput {
				return fmt.Errorf("proceso %s: debe tener al menos un recurso de salida (is_output=true)", p.MasterID)
			}
		}
	}

	// Validar edificios de venta
	saleBuildingIds := make(map[string]bool)

	for i, b := range gamedata.SaleBuildings {
		if b.MasterID == "" {
			return fmt.Errorf("edificio venta %d: master_id está vacío", i)
		}
		if b.ID == "" {
			return fmt.Errorf("edificio venta %d (%s): id está vacío", i, b.MasterID)
		}
		if b.Name == "" {
			return fmt.Errorf("edificio venta %d (%s): name está vacío", i, b.MasterID)
		}

		if saleBuildingIds[b.MasterID] {
			return fmt.Errorf("edificio venta %s: master_id duplicado", b.MasterID)
		}
		saleBuildingIds[b.MasterID] = true

		// Validar recursos de venta
		if len(b.Resources) == 0 {
			return fmt.Errorf("edificio venta %s: debe tener al menos un recurso", b.MasterID)
		}

		for _, sr := range b.Resources {
			if sr.ResourceID == "" {
				return fmt.Errorf("edificio venta %s: resource_id está vacío", b.MasterID)
			}
			if !resourceMasterIds[sr.ResourceID] {
				return fmt.Errorf("edificio venta %s: resource_id %s no existe", b.MasterID, sr.ResourceID)
			}
			if sr.PricePerUnit <= 0 {
				return fmt.Errorf("edificio venta %s: price_per_unit debe ser mayor a 0", b.MasterID)
			}
			if sr.UnitsSoldPerSecond <= 0 {
				return fmt.Errorf("edificio venta %s: units_sold_per_second debe ser mayor a 0", b.MasterID)
			}
		}
	}

	logger.Info().Msg("Validación de gamedata exitosa")
	return nil
}
