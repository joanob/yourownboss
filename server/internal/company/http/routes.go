package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/company/service"
)

// RegisterCompanyRoutes registers all company-related routes
// The router should already be scoped to /api/v1
func RegisterCompanyRoutes(router chi.Router, companyService service.CompanyService, inventoryService service.InventoryService) {
	// Note: All routes require authentication middleware to be applied at parent router level

	// Company routes
	router.Get("/company", GetCompanyHandler(companyService))
	router.Post("/company", CreateCompanyHandler(companyService))
	router.Put("/company", UpdateCompanyHandler(companyService))
	router.Delete("/company", DeleteCompanyHandler(companyService))

	// Inventory routes
	router.Get("/company/inventory", GetInventoryHandler(inventoryService))
}
