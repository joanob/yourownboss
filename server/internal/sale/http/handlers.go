package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	companyModels "github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/sale/service"
)

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

// GetSaleBuildingsHandler handles GET /api/v1/company/sale/buildings
func GetSaleBuildingsHandler(svc *service.SaleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		buildings, err := svc.GetBuildings(r.Context(), companyID)
		if err != nil {
			log.Error().Err(err).Msg("GetSaleBuildings failed")
			respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get buildings")
			return
		}

		respondData(w, http.StatusOK, buildings)
	}
}

// BuildSaleBuildingRequest is the request body for constructing a sale building
type BuildSaleBuildingRequest struct {
	SaleBuildingMasterID string `json:"sale_building_master_id"`
}

// BuildSaleBuildingHandler handles POST /api/v1/company/sale/buildings
func BuildSaleBuildingHandler(svc *service.SaleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		var req BuildSaleBuildingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}
		if req.SaleBuildingMasterID == "" {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "sale_building_master_id is required")
			return
		}

		building, err := svc.BuildSaleBuilding(r.Context(), companyID, req.SaleBuildingMasterID)
		if err != nil {
			handleSaleError(w, err)
			return
		}

		respondData(w, http.StatusCreated, building)
	}
}

// UpgradeSaleBuildingRequest is the request body for upgrading a sale building
type UpgradeSaleBuildingRequest struct {
	Levels int64 `json:"levels"`
}

// UpgradeSaleBuildingHandler handles POST /api/v1/company/sale/buildings/:id/upgrade
func UpgradeSaleBuildingHandler(svc *service.SaleService) http.HandlerFunc {
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

		var req UpgradeSaleBuildingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}
		if req.Levels <= 0 {
			req.Levels = 1
		}

		building, err := svc.UpgradeBuilding(r.Context(), companyID, buildingID, req.Levels)
		if err != nil {
			handleSaleError(w, err)
			return
		}

		respondData(w, http.StatusOK, building)
	}
}

// StartSaleRequest is the request body for starting a sale run
type StartSaleRequest struct {
	ResourceID string `json:"resource_id"`
	Units      int64  `json:"units"`
}

// StartSaleHandler handles POST /api/v1/company/sale/buildings/:id/start
func StartSaleHandler(svc *service.SaleService) http.HandlerFunc {
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

		var req StartSaleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}
		if req.ResourceID == "" {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "resource_id is required")
			return
		}
		if req.Units <= 0 {
			respondError(w, http.StatusBadRequest, "INVALID_INPUT", "units must be a positive number")
			return
		}

		building, err := svc.StartSale(r.Context(), companyID, buildingID, req.ResourceID, req.Units)
		if err != nil {
			handleSaleError(w, err)
			return
		}

		respondData(w, http.StatusOK, building)
	}
}

// CollectSaleHandler handles POST /api/v1/company/sale/buildings/:id/collect
func CollectSaleHandler(svc *service.SaleService) http.HandlerFunc {
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

		building, err := svc.CollectSale(r.Context(), companyID, buildingID)
		if err != nil {
			handleSaleError(w, err)
			return
		}

		respondData(w, http.StatusOK, building)
	}
}

// handleSaleError maps sale service errors to HTTP responses
func handleSaleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrBuildingNotFound):
		respondError(w, http.StatusNotFound, "BUILDING_NOT_FOUND", "Sale building not found")
	case errors.Is(err, service.ErrBuildingNotIdle):
		respondError(w, http.StatusConflict, "BUILDING_NOT_IDLE", "Building is under construction or already selling")
	case errors.Is(err, service.ErrBuildingNotSelling):
		respondError(w, http.StatusConflict, "BUILDING_NOT_SELLING", "Building has no active sale run to collect")
	case errors.Is(err, service.ErrSaleNotComplete):
		respondError(w, http.StatusConflict, "SALE_NOT_COMPLETE", "Sale run has not finished yet")
	case errors.Is(err, service.ErrResourceNotFound):
		respondError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Resource not available in this sale building")
	case errors.Is(err, service.ErrInvalidUnits):
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Units must be a positive number")
	case errors.Is(err, service.ErrInvalidLevels):
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "Levels must be a positive number")
	case errors.Is(err, companyModels.ErrInsufficientFunds):
		respondError(w, http.StatusConflict, "INSUFFICIENT_FUNDS", "Insufficient company funds")
	case errors.Is(err, companyModels.ErrInsufficientInventory):
		respondError(w, http.StatusConflict, "INSUFFICIENT_INVENTORY", "Insufficient resources in inventory")
	case errors.Is(err, companyModels.ErrCompanyNotFound):
		respondError(w, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
	default:
		log.Error().Err(err).Msg("Unhandled sale error")
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
