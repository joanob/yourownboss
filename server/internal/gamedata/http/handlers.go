package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/gamedata/service"
)

// HTTPResponse wraps all HTTP responses
type HTTPResponse struct {
	Data  interface{}   `json:"data,omitempty"`
	Error *ErrorDetails `json:"error,omitempty"`
}

// ErrorDetails contains error information
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GetGamedataHandler returns all gamedata (resources, buildings, processes)
func GetGamedataHandler(gamedataService *service.GamedataService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.With().Str("handler", "GetGamedata").Logger()
		logger.Debug().Msg("GetGamedata request")

		gamedata, err := gamedataService.GetGameData(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get gamedata")
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve gamedata")
			return
		}

		respondWithData(w, http.StatusOK, gamedata)
	}
}

// PostGamedataHandler imports gamedata (admin only)
func PostGamedataHandler(gamedataService *service.GamedataService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.With().Str("handler", "PostGamedata").Logger()
		logger.Debug().Msg("PostGamedata request")

		var req service.GamedataImportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().Err(err).Msg("Invalid request body")
			respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}

		if err := gamedataService.ImportGameData(r.Context(), &req); err != nil {
			logger.Error().Err(err).Msg("Failed to import gamedata")
			respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
			return
		}

		respondWithData(w, http.StatusCreated, map[string]bool{"imported": true})
	}
}

// GetResourcesHandler returns all resources
func GetResourcesHandler(gamedataService *service.GamedataService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.With().Str("handler", "GetResources").Logger()
		logger.Debug().Msg("GetResources request")

		gamedata, err := gamedataService.GetGameData(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get resources")
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve resources")
			return
		}

		respondWithData(w, http.StatusOK, gamedata.Resources)
	}
}

// GetResourceHandler returns a specific resource by ID
func GetResourceHandler(gamedataService *service.GamedataService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.With().Str("handler", "GetResource").Logger()
		resourceID := chi.URLParam(r, "resourceID")
		logger.Debug().Str("resource_id", resourceID).Msg("GetResource request")

		gamedata, err := gamedataService.GetGameData(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get gamedata")
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve resource")
			return
		}

		// Search for resource in cache
		for _, res := range gamedata.Resources {
			if res.ID == resourceID {
				respondWithData(w, http.StatusOK, res)
				return
			}
		}

		logger.Debug().Str("resource_id", resourceID).Msg("Resource not found")
		respondWithError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Resource not found")
	}
}

// GetProductionBuildingsHandler returns all production buildings with their processes
func GetProductionBuildingsHandler(gamedataService *service.GamedataService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.With().Str("handler", "GetProductionBuildings").Logger()
		logger.Debug().Msg("GetProductionBuildings request")

		gamedata, err := gamedataService.GetGameData(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get gamedata")
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve production buildings")
			return
		}

		respondWithData(w, http.StatusOK, gamedata.ProductionBuildings)
	}
}

// GetProductionBuildingHandler returns a specific production building by ID
func GetProductionBuildingHandler(gamedataService *service.GamedataService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.With().Str("handler", "GetProductionBuilding").Logger()
		buildingID := chi.URLParam(r, "buildingID")
		logger.Debug().Str("building_id", buildingID).Msg("GetProductionBuilding request")

		gamedata, err := gamedataService.GetGameData(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get gamedata")
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve production building")
			return
		}

		// Search for building in cache
		for _, building := range gamedata.ProductionBuildings {
			if building.ID == buildingID {
				respondWithData(w, http.StatusOK, building)
				return
			}
		}

		logger.Debug().Str("building_id", buildingID).Msg("Production building not found")
		respondWithError(w, http.StatusNotFound, "BUILDING_NOT_FOUND", "Production building not found")
	}
}

// GetSaleBuildingsHandler returns all sale buildings with their resources
func GetSaleBuildingsHandler(gamedataService *service.GamedataService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.With().Str("handler", "GetSaleBuildings").Logger()
		logger.Debug().Msg("GetSaleBuildings request")

		gamedata, err := gamedataService.GetGameData(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get gamedata")
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve sale buildings")
			return
		}

		respondWithData(w, http.StatusOK, gamedata.SaleBuildings)
	}
}

// GetSaleBuildingHandler returns a specific sale building by ID
func GetSaleBuildingHandler(gamedataService *service.GamedataService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.With().Str("handler", "GetSaleBuilding").Logger()
		buildingID := chi.URLParam(r, "buildingID")
		logger.Debug().Str("building_id", buildingID).Msg("GetSaleBuilding request")

		gamedata, err := gamedataService.GetGameData(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get gamedata")
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve sale building")
			return
		}

		// Search for building in cache
		for _, building := range gamedata.SaleBuildings {
			if building.ID == buildingID {
				respondWithData(w, http.StatusOK, building)
				return
			}
		}

		logger.Debug().Str("building_id", buildingID).Msg("Sale building not found")
		respondWithError(w, http.StatusNotFound, "BUILDING_NOT_FOUND", "Sale building not found")
	}
}

// Helper functions

func respondWithData(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(HTTPResponse{Data: data})
}

func respondWithError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(HTTPResponse{
		Error: &ErrorDetails{
			Code:    code,
			Message: message,
		},
	})
}
