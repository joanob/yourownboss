package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	companyModels "github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	"github.com/joanob/yourownboss/internal/production/service"
)

const productionRateLimit = 20

// httpResponse wraps all HTTP responses
type httpResponse struct {
	Data  interface{}   `json:"data,omitempty"`
	Error *errorDetails `json:"error,omitempty"`
}

// errorDetails contains error information
type errorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GetBuildingsHandler handles GET /api/v1/company/production/buildings
func GetBuildingsHandler(svc *service.ProductionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		buildings, err := svc.GetBuildings(r.Context(), companyID)
		if err != nil {
			log.Error().Err(err).Msg("GetBuildings failed")
			respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get buildings")
			return
		}

		respondData(w, http.StatusOK, buildings)
	}
}

// BuildProductionBuildingRequest is the request body for building a new production building
type BuildProductionBuildingRequest struct {
	ProductionBuildingMasterID string `json:"production_building_master_id"`
}

// BuildProductionBuildingHandler handles POST /api/v1/company/production/buildings
func BuildProductionBuildingHandler(svc *service.ProductionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		var req BuildProductionBuildingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}
		if req.ProductionBuildingMasterID == "" {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "production_building_master_id is required")
			return
		}

		building, err := svc.BuildProductionBuilding(r.Context(), companyID, req.ProductionBuildingMasterID)
		if err != nil {
			handleProductionError(w, err)
			return
		}

		respondData(w, http.StatusCreated, building)
	}
}

// UpgradeBuildingRequest is the request body for upgrading a building
type UpgradeBuildingRequest struct {
	Levels int64 `json:"levels"`
}

// UpgradeBuildingHandler handles POST /api/v1/company/production/buildings/:id/upgrade
func UpgradeBuildingHandler(svc *service.ProductionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		buildingID := chi.URLParam(r, "id")
		if buildingID == "" {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "building id is required")
			return
		}

		var req UpgradeBuildingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}
		if req.Levels <= 0 {
			req.Levels = 1
		}

		building, err := svc.UpgradeBuilding(r.Context(), companyID, buildingID, req.Levels)
		if err != nil {
			handleProductionError(w, err)
			return
		}

		respondData(w, http.StatusOK, building)
	}
}

// StartProductionRequest is the request body for starting a production run
type StartProductionRequest struct {
	ProcessMasterID string `json:"process_master_id"`
	Cycles          int64  `json:"cycles"`
}

// StartProductionHandler handles POST /api/v1/company/production/buildings/:id/start
func StartProductionHandler(svc *service.ProductionService, rateLimiter *cache.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("user_id").(string)
		if !ok || userID == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		if !rateLimiter.Allow(userID, "production_start", productionRateLimit) {
			log.Warn().Str("user_id", userID).Msg("Rate limit exceeded for production start")
			respondError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please slow down")
			return
		}

		buildingID := chi.URLParam(r, "id")
		if buildingID == "" {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "building id is required")
			return
		}

		var req StartProductionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}
		if req.ProcessMasterID == "" {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "process_master_id is required")
			return
		}
		if req.Cycles <= 0 {
			req.Cycles = 1
		}

		building, err := svc.StartProduction(r.Context(), companyID, buildingID, req.ProcessMasterID, req.Cycles)
		if err != nil {
			handleProductionError(w, err)
			return
		}

		rateLimiter.Record(userID, "production_start")

		respondData(w, http.StatusOK, building)
	}
}

// CollectProductionHandler handles POST /api/v1/company/production/buildings/:id/collect
func CollectProductionHandler(svc *service.ProductionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		buildingID := chi.URLParam(r, "id")
		if buildingID == "" {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "building id is required")
			return
		}

		building, err := svc.CollectProduction(r.Context(), companyID, buildingID)
		if err != nil {
			handleProductionError(w, err)
			return
		}

		respondData(w, http.StatusOK, building)
	}
}

// handleProductionError maps production service errors to HTTP responses
func handleProductionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrBuildingNotFound):
		respondError(w, http.StatusNotFound, "BUILDING_NOT_FOUND", "Production building not found")
	case errors.Is(err, service.ErrBuildingNotIdle):
		respondError(w, http.StatusConflict, "BUILDING_NOT_IDLE", "Building is under construction or already producing")
	case errors.Is(err, service.ErrBuildingNotProducing):
		respondError(w, http.StatusConflict, "BUILDING_NOT_PRODUCING", "Building has no active production run to collect")
	case errors.Is(err, service.ErrProductionNotComplete):
		respondError(w, http.StatusConflict, "PRODUCTION_NOT_COMPLETE", "Production run has not finished yet")
	case errors.Is(err, service.ErrProcessNotFound):
		respondError(w, http.StatusNotFound, "PROCESS_NOT_FOUND", "Production process not found")
	case errors.Is(err, service.ErrProcessNotAvailable):
		respondError(w, http.StatusBadRequest, "PROCESS_NOT_AVAILABLE_NOW", "Process is outside its allowed time window")
	case errors.Is(err, service.ErrInvalidCycles):
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Cycles must be a positive number")
	case errors.Is(err, companyModels.ErrInsufficientFunds):
		respondError(w, http.StatusConflict, "INSUFFICIENT_FUNDS", "Insufficient company funds")
	case errors.Is(err, companyModels.ErrInsufficientInventory):
		respondError(w, http.StatusConflict, "INSUFFICIENT_INVENTORY", "Insufficient resources in inventory")
	case errors.Is(err, companyModels.ErrCompanyNotFound):
		respondError(w, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
	default:
		log.Error().Err(err).Msg("Unhandled production error")
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
	}
}

func respondData(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(httpResponse{Data: data})
}

func respondError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(httpResponse{Error: &errorDetails{Code: code, Message: message}})
}
