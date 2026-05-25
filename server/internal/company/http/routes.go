package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/company/service"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// RegisterCompanyRoutes registers all company-related routes.
// The router should already be scoped to /api/v1 with auth middleware applied.
func RegisterCompanyRoutes(router chi.Router, companyService service.CompanyService, inventoryService service.InventoryService, initialMoney int64, sessionCache *cache.SessionCache) {
	router.Get("/company", GetCompanyHandler(companyService))
	router.Post("/company", CreateCompanyHandler(companyService, initialMoney))
	router.Put("/company", UpdateCompanyHandler(companyService))
	router.Delete("/company", DeleteCompanyHandler(companyService, sessionCache))

	router.Get("/company/inventory", GetInventoryHandler(companyService, inventoryService))
}
