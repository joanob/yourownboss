package http

import (
	"encoding/json"
	"net/http"

	"yourownboss/internal/auth"
	"yourownboss/internal/repository"
	"yourownboss/internal/service"
)

type InventoryHandler struct {
	inventoryService service.InventoryService
	companyRepo      repository.CompanyRepository
}

type MarketHandler struct {
	marketService service.MarketService
	companyRepo   repository.CompanyRepository
}

func NewInventoryHandler(inventoryService service.InventoryService, companyRepo repository.CompanyRepository) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
		companyRepo:      companyRepo,
	}
}

func NewMarketHandler(marketService service.MarketService, companyRepo repository.CompanyRepository) *MarketHandler {
	return &MarketHandler{
		marketService: marketService,
		companyRepo:   companyRepo,
	}
}

// --- Response Types ---

type InventoryItemResponse struct {
	ID         int64  `json:"id"`
	ResourceID int64  `json:"resource_id"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Quantity   int64  `json:"quantity"`
	Price      int64  `json:"price"`     // Price per pack
	PackSize   int64  `json:"pack_size"` // Units per pack
}

type ResourceResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Price       int64  `json:"price"`     // Price per pack
	PackSize    int64  `json:"pack_size"` // Units per pack
}

type BuyRequest struct {
	ResourceID int64 `json:"resource_id"`
	PackCount  int64 `json:"pack_count"` // Number of packs to buy
}

type SellRequest struct {
	ResourceID int64 `json:"resource_id"`
	PackCount  int64 `json:"pack_count"` // Number of packs to sell
}

// --- Inventory Handler Methods ---

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get company from user
	company, err := h.companyRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == repository.ErrCompanyNotFound {
			http.Error(w, "Company not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get company", http.StatusInternalServerError)
		}
		return
	}

	inventory, err := h.inventoryService.GetInventory(ctx, company.ID)
	if err != nil {
		http.Error(w, "Failed to get inventory", http.StatusInternalServerError)
		return
	}

	response := make([]InventoryItemResponse, 0)
	for _, item := range inventory {
		response = append(response, InventoryItemResponse{
			ID:         item.ID,
			ResourceID: item.ResourceID,
			Name:       item.Name,
			Icon:       item.Icon,
			Quantity:   item.Quantity,
			Price:      item.Price,
			PackSize:   item.PackSize,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *InventoryHandler) GetResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resources, err := h.inventoryService.GetAllResources(ctx)
	if err != nil {
		http.Error(w, "Failed to get resources", http.StatusInternalServerError)
		return
	}

	response := make([]ResourceResponse, 0)
	for _, res := range resources {
		response = append(response, ResourceResponse{
			ID:          res.ID,
			Name:        res.Name,
			Icon:        res.Icon,
			Description: res.Description,
			Price:       res.Price,
			PackSize:    res.PackSize,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// --- Market Handler Methods ---

func (h *MarketHandler) BuyResource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get company from user
	company, err := h.companyRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == repository.ErrCompanyNotFound {
			http.Error(w, "Company not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get company", http.StatusInternalServerError)
		}
		return
	}

	var req BuyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PackCount <= 0 {
		http.Error(w, "Pack count must be positive", http.StatusBadRequest)
		return
	}

	err = h.marketService.BuyResource(ctx, company.ID, req.ResourceID, req.PackCount)
	if err != nil {
		switch err {
		case service.ErrMarketInsufficientFunds:
			http.Error(w, "Insufficient funds", http.StatusBadRequest)
		case repository.ErrInsufficientStock:
			http.Error(w, "Insufficient stock", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to buy resource", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Resource purchased successfully"})
}

func (h *MarketHandler) SellResource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get company from user
	company, err := h.companyRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == repository.ErrCompanyNotFound {
			http.Error(w, "Company not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get company", http.StatusInternalServerError)
		}
		return
	}

	var req SellRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PackCount <= 0 {
		http.Error(w, "Pack count must be positive", http.StatusBadRequest)
		return
	}

	err = h.marketService.SellResource(ctx, company.ID, req.ResourceID, req.PackCount)
	if err != nil {
		switch err {
		case repository.ErrInsufficientStock:
			http.Error(w, "Insufficient stock", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to sell resource", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Resource sold successfully"})
}
