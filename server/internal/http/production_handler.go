package http

import (
	"encoding/json"
	"net/http"

	"yourownboss/internal/service"
)

// ProductionHandler handles HTTP requests for production buildings.
type ProductionHandler struct {
	productionService service.ProductionService
}

// NewProductionHandler creates a new production handler.
func NewProductionHandler(productionService service.ProductionService) *ProductionHandler {
	return &ProductionHandler{productionService: productionService}
}

type ProductionBuildingResponse struct {
	ID        int64                       `json:"id"`
	Name      string                      `json:"name"`
	Cost      int64                       `json:"cost"`
	Processes []ProductionProcessResponse `json:"processes"`
}

type ProductionProcessResponse struct {
	ID               int64                               `json:"id"`
	Name             string                              `json:"name"`
	ProcessingTimeMs int64                               `json:"processing_time_ms"`
	WindowStartHour  *int64                              `json:"window_start_hour"`
	WindowEndHour    *int64                              `json:"window_end_hour"`
	Resources        []ProductionProcessResourceResponse `json:"resources"`
}

type ProductionProcessResourceResponse struct {
	ResourceID   int64  `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	Direction    string `json:"direction"`
	Quantity     int64  `json:"quantity"`
}

// GetProductionBuildings returns buildings with processes and resources.
func (h *ProductionHandler) GetProductionBuildings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	buildings, err := h.productionService.GetProductionBuildings(ctx)
	if err != nil {
		http.Error(w, "Failed to get production buildings", http.StatusInternalServerError)
		return
	}

	response := make([]ProductionBuildingResponse, 0, len(buildings))
	for _, building := range buildings {
		processes := make([]ProductionProcessResponse, 0, len(building.Processes))
		for _, process := range building.Processes {
			resources := make([]ProductionProcessResourceResponse, 0, len(process.Resources))
			for _, resource := range process.Resources {
				resources = append(resources, ProductionProcessResourceResponse{
					ResourceID:   resource.ResourceID,
					ResourceName: resource.ResourceName,
					Direction:    resource.Direction,
					Quantity:     resource.Quantity,
				})
			}

			processes = append(processes, ProductionProcessResponse{
				ID:               process.ID,
				Name:             process.Name,
				ProcessingTimeMs: process.ProcessingTimeMs,
				WindowStartHour:  process.WindowStartHour,
				WindowEndHour:    process.WindowEndHour,
				Resources:        resources,
			})
		}

		response = append(response, ProductionBuildingResponse{
			ID:        building.ID,
			Name:      building.Name,
			Cost:      building.Cost,
			Processes: processes,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
