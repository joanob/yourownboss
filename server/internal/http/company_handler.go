package http

import (
	"encoding/json"
	"net/http"
	"time"

	"yourownboss/internal/auth"
	"yourownboss/internal/service"
)

type CompanyHandler struct {
	companyService service.CompanyService
}

func NewCompanyHandler(companyService service.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
	}
}

type CreateCompanyRequest struct {
	Name string `json:"name"`
}

type CompanyResponse struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	Money     int64  `json:"money"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	company, err := h.companyService.CreateCompany(ctx, userID, req.Name)
	if err != nil {
		switch err {
		case service.ErrCompanyAlreadyExists:
			http.Error(w, "User already has a company", http.StatusConflict)
		case service.ErrInvalidCompanyName:
			http.Error(w, "Company name must be between 3 and 50 characters", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := CompanyResponse{
		ID:        company.ID,
		UserID:    company.UserID,
		Name:      company.Name,
		Money:     company.Money,
		CreatedAt: company.CreatedAt.Format(time.RFC3339),
		UpdatedAt: company.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *CompanyHandler) GetMyCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	company, err := h.companyService.GetCompanyByUserID(ctx, userID)
	if err != nil {
		switch err {
		case service.ErrCompanyNotFound:
			http.Error(w, "Company not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := CompanyResponse{
		ID:        company.ID,
		UserID:    company.UserID,
		Name:      company.Name,
		Money:     company.Money,
		CreatedAt: company.CreatedAt.Format(time.RFC3339),
		UpdatedAt: company.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
