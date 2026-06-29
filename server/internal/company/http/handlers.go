package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/company/service"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	appvalidator "github.com/joanob/yourownboss/internal/pkg/validator"
)

// validate es el validador usado por los handlers de company. Se inicializa con
// validación activa y RegisterCompanyRoutes lo reemplaza por el validador de la
// aplicación (que puede omitir la validación en development).
var validate = appvalidator.New(false)

// errNoContext is returned when a required value is missing from the request context.
var errNoContext = errors.New("value not in context")

// companyHTTPResponse wraps all API responses for the company domain.
type companyHTTPResponse struct {
	Data  interface{}  `json:"data,omitempty"`
	Error *errorDetail `json:"error,omitempty"`
}

// errorDetail follows the project-wide error format: {"code": "...", "message": "..."}.
type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ============================================================================
// Requests and Responses
// ============================================================================

// CreateCompanyRequest represents the request body for creating a company.
// Initial money is set server-side from INITIAL_COMPANY_MONEY env var.
type CreateCompanyRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateCompanyRequest represents the request body for updating a company.
type UpdateCompanyRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CompanyDTO represents a company in API responses.
type CompanyDTO struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Money     int64  `json:"money"`
	CreatedAt string `json:"created_at"`
}

// InventoryItemDTO represents an inventory item in API responses.
type InventoryItemDTO struct {
	ID         string `json:"id"`
	ResourceID string `json:"resource_id"`
	Quantity   int64  `json:"quantity"`
}

// InventoryDTO represents the entire inventory in API responses.
type InventoryDTO struct {
	Items []InventoryItemDTO `json:"items"`
}

// ============================================================================
// Utility Functions
// ============================================================================

func companyToDTO(c *models.Company) CompanyDTO {
	return CompanyDTO{
		ID:        c.ID,
		UserID:    c.UserID,
		Name:      c.Name,
		Money:     c.Money,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func inventoryToDTO(inv *models.CompanyInventory) InventoryDTO {
	items := make([]InventoryItemDTO, 0, len(inv.Items))
	for _, item := range inv.Items {
		items = append(items, InventoryItemDTO{
			ID:         item.ID,
			ResourceID: item.ResourceID,
			Quantity:   item.Quantity,
		})
	}
	return InventoryDTO{Items: items}
}

// writeError writes an error response in the project-standard format.
func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(companyHTTPResponse{
		Error: &errorDetail{Code: code, Message: message},
	})
}

// writeJSON writes a successful JSON response.
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(companyHTTPResponse{Data: data})
}

// getUserIDFromContext extracts userID from the request context (set by auth middleware).
func getUserIDFromContext(r *http.Request) (string, error) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		return "", errNoContext
	}
	return userID, nil
}

// ============================================================================
// Handlers
// ============================================================================

// GetCompanyHandler handles GET /api/v1/company
func GetCompanyHandler(svc service.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "GetCompany").Logger()

		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		company, err := svc.GetCompany(r.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get company")
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get company")
			return
		}

		if company == nil {
			logger.Warn().Str("user_id", userID).Msg("Company not found")
			writeError(w, http.StatusForbidden, "COMPANY_NOT_FOUND", "User does not have a company")
			return
		}

		logger.Debug().Str("company_id", company.ID).Msg("Company retrieved successfully")
		writeJSON(w, http.StatusOK, companyToDTO(company))
	}
}

// CreateCompanyHandler handles POST /api/v1/company.
// initialMoney is configured server-side via INITIAL_COMPANY_MONEY env var.
func CreateCompanyHandler(svc service.CompanyService, initialMoney int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "CreateCompany").Logger()

		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		var req CreateCompanyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().Err(err).Msg("Failed to parse request body")
			writeError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			logger.Warn().Err(err).Msg("Request validation failed")
			writeError(w, http.StatusBadRequest, "INVALID_INPUT", "Validation failed: name is required (1–100 chars)")
			return
		}

		company, err := svc.CreateCompany(r.Context(), userID, req.Name, initialMoney)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrCompanyAlreadyExists):
				logger.Warn().Str("user_id", userID).Msg("Company already exists")
				writeError(w, http.StatusConflict, "COMPANY_ALREADY_EXISTS", "User already has a company")
			default:
				logger.Error().Err(err).Msg("Failed to create company")
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create company")
			}
			return
		}

		logger.Info().Str("company_id", company.ID).Msg("Company created successfully")
		writeJSON(w, http.StatusCreated, companyToDTO(company))
	}
}

// UpdateCompanyHandler handles PUT /api/v1/company
func UpdateCompanyHandler(svc service.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "UpdateCompany").Logger()

		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		company, err := svc.GetCompany(r.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get company")
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get company")
			return
		}
		if company == nil {
			logger.Warn().Str("user_id", userID).Msg("Company not found")
			writeError(w, http.StatusForbidden, "COMPANY_NOT_FOUND", "User does not have a company")
			return
		}

		var req UpdateCompanyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().Err(err).Msg("Failed to parse request body")
			writeError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			logger.Warn().Err(err).Msg("Request validation failed")
			writeError(w, http.StatusBadRequest, "INVALID_INPUT", "Validation failed: name is required (1–100 chars)")
			return
		}

		updatedCompany, err := svc.UpdateCompanyName(r.Context(), company.ID, req.Name)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to update company")
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update company")
			return
		}

		logger.Info().Str("company_id", company.ID).Msg("Company updated successfully")
		writeJSON(w, http.StatusOK, companyToDTO(updatedCompany))
	}
}

// DeleteCompanyHandler handles DELETE /api/v1/company.
// After deletion, clears company_id from all active session cache entries for the user
// so that renewed session tokens reflect the deleted company (per spec).
func DeleteCompanyHandler(svc service.CompanyService, sessionCache *cache.SessionCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "DeleteCompany").Logger()

		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		company, err := svc.GetCompany(r.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get company")
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get company")
			return
		}
		if company == nil {
			logger.Warn().Str("user_id", userID).Msg("Company not found")
			writeError(w, http.StatusForbidden, "COMPANY_NOT_FOUND", "User does not have a company")
			return
		}

		if err := svc.DeleteCompany(r.Context(), company.ID); err != nil {
			logger.Error().Err(err).Msg("Failed to delete company")
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete company")
			return
		}

		// Clear company_id from all active sessions so renewed tokens have company_id = null.
		sessionCache.ClearCompanyID(userID)

		logger.Info().Str("company_id", company.ID).Msg("Company deleted successfully")
		writeJSON(w, http.StatusOK, map[string]interface{}{})
	}
}

// GetInventoryHandler handles GET /api/v1/company/inventory
func GetInventoryHandler(companySvc service.CompanyService, invSvc service.InventoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "GetInventory").Logger()

		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		// Look up company by userID (company_id in JWT may be stale or nil).
		company, err := companySvc.GetCompany(r.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get company")
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get company")
			return
		}
		if company == nil {
			logger.Warn().Str("user_id", userID).Msg("Company not found")
			writeError(w, http.StatusForbidden, "COMPANY_NOT_FOUND", "User does not have a company")
			return
		}

		inventory, err := invSvc.GetInventory(r.Context(), company.ID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get inventory")
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get inventory")
			return
		}
		if inventory == nil {
			inventory = models.NewCompanyInventory(company.ID)
		}

		logger.Debug().Str("company_id", company.ID).Int("item_count", len(inventory.Items)).Msg("Inventory retrieved successfully")
		writeJSON(w, http.StatusOK, inventoryToDTO(inventory))
	}
}
