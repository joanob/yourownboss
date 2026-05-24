package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/company/service"
)

var validate = validator.New()

// APIResponse wraps all API responses
type APIResponse[T any] struct {
	Data T `json:"data"`
}

// ErrorResponse wraps error responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ============================================================================
// Requests and Responses
// ============================================================================

// CreateCompanyRequest represents the request body for creating a company
type CreateCompanyRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100"`
	InitialMoney int64  `json:"initial_money" validate:"required,min=0"`
}

// UpdateCompanyRequest represents the request body for updating a company
type UpdateCompanyRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CompanyDTO represents a company in API responses
type CompanyDTO struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Money     int64  `json:"money"`
	CreatedAt string `json:"created_at"`
}

// InventoryItemDTO represents an inventory item in API responses
type InventoryItemDTO struct {
	ID         string `json:"id"`
	ResourceID string `json:"resource_id"`
	Quantity   int64  `json:"quantity"`
}

// InventoryDTO represents the entire inventory in API responses
type InventoryDTO struct {
	Items []InventoryItemDTO `json:"items"`
}

// ============================================================================
// Utility Functions
// ============================================================================

// companyToDTO converts a domain Company to API DTO
func companyToDTO(c *models.Company) CompanyDTO {
	return CompanyDTO{
		ID:        c.ID,
		UserID:    c.UserID,
		Name:      c.Name,
		Money:     c.Money,
		CreatedAt: c.CreatedAt,
	}
}

// inventoryToDTO converts a domain CompanyInventory to API DTO
func inventoryToDTO(inv *models.CompanyInventory) InventoryDTO {
	items := make([]InventoryItemDTO, 0, len(*inv))
	for _, item := range *inv {
		items = append(items, InventoryItemDTO{
			ID:         item.ID,
			ResourceID: item.ResourceID,
			Quantity:   item.Quantity,
		})
	}
	return InventoryDTO{Items: items}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, statusCode int, errorMsg string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorMsg,
		Message: details,
	})
}

// writeJSON writes a successful JSON response
func writeJSON[T any](w http.ResponseWriter, statusCode int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse[T]{Data: data})
}

// getUserIDFromContext extracts userID from request context
// In a real application, this would come from the authentication middleware
func getUserIDFromContext(r *http.Request) (string, error) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		return "", models.ErrInsufficientFunds // Reusing error type
	}
	return userID.(string), nil
}

// getCompanyIDFromContext extracts companyID from request context
func getCompanyIDFromContext(r *http.Request) (string, error) {
	companyID := r.Context().Value("company_id")
	if companyID == nil {
		return "", models.ErrInsufficientFunds // Reusing error type
	}
	return companyID.(string), nil
}

// ============================================================================
// Handlers
// ============================================================================

// GetCompanyHandler handles GET /api/v1/company
func GetCompanyHandler(svc service.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "GetCompany").Logger()

		// Get userID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
			return
		}

		// Get company
		company, err := svc.GetCompany(r.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get company")
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get company")
			return
		}

		if company == nil {
			logger.Warn().Str("user_id", userID).Msg("Company not found")
			writeError(w, http.StatusNotFound, "not_found", "Company not found")
			return
		}

		logger.Debug().Str("company_id", company.ID).Msg("Company retrieved successfully")
		writeJSON(w, http.StatusOK, companyToDTO(company))
	}
}

// CreateCompanyHandler handles POST /api/v1/company
func CreateCompanyHandler(svc service.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "CreateCompany").Logger()

		// Get userID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
			return
		}

		// Parse request body
		var req CreateCompanyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().Err(err).Msg("Failed to parse request body")
			writeError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
			return
		}

		// Validate request
		if err := validate.Struct(req); err != nil {
			logger.Warn().Err(err).Msg("Request validation failed")
			writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
			return
		}

		// Create company
		company, err := svc.CreateCompany(r.Context(), userID, req.Name, req.InitialMoney)
		if err != nil {
			if err == models.ErrCompanyAlreadyExists {
				logger.Warn().Str("user_id", userID).Msg("Company already exists")
				writeError(w, http.StatusConflict, "company_exists", "Company already exists for this user")
				return
			}
			logger.Error().Err(err).Msg("Failed to create company")
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create company")
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

		// Get userID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
			return
		}

		// Get company to find its ID
		company, err := svc.GetCompany(r.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get company")
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get company")
			return
		}

		if company == nil {
			logger.Warn().Str("user_id", userID).Msg("Company not found")
			writeError(w, http.StatusNotFound, "not_found", "Company not found")
			return
		}

		// Parse request body
		var req UpdateCompanyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().Err(err).Msg("Failed to parse request body")
			writeError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
			return
		}

		// Validate request
		if err := validate.Struct(req); err != nil {
			logger.Warn().Err(err).Msg("Request validation failed")
			writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
			return
		}

		// Update company
		updatedCompany, err := svc.UpdateCompanyName(r.Context(), company.ID, req.Name)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to update company")
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update company")
			return
		}

		logger.Info().Str("company_id", company.ID).Msg("Company updated successfully")
		writeJSON(w, http.StatusOK, companyToDTO(updatedCompany))
	}
}

// DeleteCompanyHandler handles DELETE /api/v1/company
func DeleteCompanyHandler(svc service.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "DeleteCompany").Logger()

		// Get userID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
			return
		}

		// Get company to find its ID
		company, err := svc.GetCompany(r.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get company")
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get company")
			return
		}

		if company == nil {
			logger.Warn().Str("user_id", userID).Msg("Company not found")
			writeError(w, http.StatusNotFound, "not_found", "Company not found")
			return
		}

		// Delete company
		err = svc.DeleteCompany(r.Context(), company.ID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to delete company")
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete company")
			return
		}

		logger.Info().Str("company_id", company.ID).Msg("Company deleted successfully")
		w.WriteHeader(http.StatusNoContent)
	}
}

// GetInventoryHandler handles GET /api/v1/company/inventory
func GetInventoryHandler(svc service.InventoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "GetInventory").Logger()

		// Get userID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r)
		if err != nil {
			logger.Warn().Msg("User not authenticated")
			writeError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
			return
		}

		// For now, get companyID from context (in a real app, we'd look up company by userID)
		// This would normally come from auth middleware or be looked up from userID
		companyID, err := getCompanyIDFromContext(r)
		if err != nil {
			// Fallback: Try to get from user's company
			// This is a temporary solution - in production we'd have this in auth middleware
			logger.Warn().Str("user_id", userID).Msg("Company ID not in context")
			writeError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
			return
		}

		// Get inventory
		inventory, err := svc.GetInventory(r.Context(), companyID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get inventory")
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get inventory")
			return
		}

		if inventory == nil {
			logger.Warn().Str("company_id", companyID).Msg("Inventory not found")
			writeError(w, http.StatusNotFound, "not_found", "Inventory not found")
			return
		}

		logger.Debug().Str("company_id", companyID).Int("item_count", len(*inventory)).Msg("Inventory retrieved successfully")
		writeJSON(w, http.StatusOK, inventoryToDTO(inventory))
	}
}
