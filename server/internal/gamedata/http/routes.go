package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/gamedata/service"
)

// RegisterGamedataRoutes registers all public (unauthenticated) gamedata read endpoints.
// The router must already be within /api/v1 context.
func RegisterGamedataRoutes(router chi.Router, gamedataService *service.GamedataService) {
	router.Get("/gamedata", GetGamedataHandler(gamedataService))
	router.Get("/resources", GetResourcesHandler(gamedataService))
	router.Get("/resources/{resourceID}", GetResourceHandler(gamedataService))
	router.Get("/production/buildings", GetProductionBuildingsHandler(gamedataService))
	router.Get("/production/buildings/{buildingID}", GetProductionBuildingHandler(gamedataService))
	router.Get("/sale/buildings", GetSaleBuildingsHandler(gamedataService))
	router.Get("/sale/buildings/{buildingID}", GetSaleBuildingHandler(gamedataService))
}

// RegisterAdminGamedataRoutes registers the admin-only gamedata import endpoint.
// The router must already have AuthMiddleware, RequireAuth, and RequireAdmin applied.
func RegisterAdminGamedataRoutes(router chi.Router, gamedataService *service.GamedataService) {
	router.Post("/gamedata", PostGamedataHandler(gamedataService))
}
