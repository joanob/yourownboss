package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"

	companyModels "github.com/joanob/yourownboss/internal/company/models"
	"github.com/joanob/yourownboss/internal/market/service"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

const marketRateLimit = 20

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

// MarketRequest is the request body for buy and sell operations
type MarketRequest struct {
	ResourceMasterID string `json:"resource_master_id"`
	Quantity         int64  `json:"quantity"`
}

// MarketResponse is the response body for buy and sell operations
type MarketResponse struct {
	Company   interface{} `json:"company"`
	Inventory interface{} `json:"inventory"`
}

// BuyResourceHandler handles POST /api/v1/market/buy
func BuyResourceHandler(marketService *service.MarketService, rateLimiter *cache.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "BuyResource").Logger()

		userID, ok := r.Context().Value("user_id").(string)
		if !ok || userID == "" {
			logger.Warn().Msg("No user_id in context")
			respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			logger.Warn().Msg("No company_id in context")
			respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		if !rateLimiter.AllowAndRecord(userID, "market_buy", marketRateLimit) {
			logger.Warn().Str("user_id", userID).Msg("Rate limit exceeded for market buy")
			respondWithError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please slow down")
			return
		}

		var req MarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().Err(err).Msg("Invalid request body")
			respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}

		if req.ResourceMasterID == "" {
			respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "resource_master_id is required")
			return
		}
		if req.Quantity <= 0 {
			respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "quantity must be positive")
			return
		}

		result, err := marketService.BuyResource(r.Context(), companyID, req.ResourceMasterID, req.Quantity)
		if err != nil {
			handleMarketError(w, err)
			return
		}

		respondWithData(w, http.StatusOK, MarketResponse{
			Company:   result.Company.ToDTO(),
			Inventory: result.Inventory.ToDTO(),
		})
	}
}

// SellResourceHandler handles POST /api/v1/market/sell
func SellResourceHandler(marketService *service.MarketService, rateLimiter *cache.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("handler", "SellResource").Logger()

		userID, ok := r.Context().Value("user_id").(string)
		if !ok || userID == "" {
			logger.Warn().Msg("No user_id in context")
			respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		companyID, ok := r.Context().Value("company_id").(string)
		if !ok || companyID == "" {
			logger.Warn().Msg("No company_id in context")
			respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}

		if !rateLimiter.AllowAndRecord(userID, "market_sell", marketRateLimit) {
			logger.Warn().Str("user_id", userID).Msg("Rate limit exceeded for market sell")
			respondWithError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please slow down")
			return
		}

		var req MarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().Err(err).Msg("Invalid request body")
			respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body")
			return
		}

		if req.ResourceMasterID == "" {
			respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "resource_master_id is required")
			return
		}
		if req.Quantity <= 0 {
			respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "quantity must be positive")
			return
		}

		result, err := marketService.SellResource(r.Context(), companyID, req.ResourceMasterID, req.Quantity)
		if err != nil {
			handleMarketError(w, err)
			return
		}

		respondWithData(w, http.StatusOK, MarketResponse{
			Company:   result.Company.ToDTO(),
			Inventory: result.Inventory.ToDTO(),
		})
	}
}

func handleMarketError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrResourceNotFound):
		respondWithError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Resource not found")
	case errors.Is(err, service.ErrInvalidQuantityMultiple):
		respondWithError(w, http.StatusBadRequest, "INVALID_QUANTITY_MULTIPLE", "Quantity must be a multiple of market_sale_qty")
	case errors.Is(err, companyModels.ErrInsufficientFunds):
		respondWithError(w, http.StatusConflict, "INSUFFICIENT_FUNDS", "Insufficient funds")
	case errors.Is(err, companyModels.ErrInsufficientInventory):
		respondWithError(w, http.StatusConflict, "INSUFFICIENT_INVENTORY", "Insufficient inventory")
	case errors.Is(err, companyModels.ErrCompanyNotFound):
		respondWithError(w, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
	default:
		log.Error().Err(err).Msg("Market operation failed")
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Operation failed")
	}
}

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
