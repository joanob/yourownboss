package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/gamedata/service"
)

// RegisterGamedataRoutes registers all public (unauthenticated) gamedata read endpoints.
func RegisterGamedataRoutes(router chi.Router, gamedataService *service.GamedataService) {
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/gamedata", GetGamedataHandler(gamedataService))
		r.Get("/resources", GetResourcesHandler(gamedataService))
		r.Get("/resources/{resourceID}", GetResourceHandler(gamedataService))
		r.Get("/production/buildings", GetProductionBuildingsHandler(gamedataService))
		r.Get("/production/buildings/{buildingID}", GetProductionBuildingHandler(gamedataService))
		r.Get("/sale/buildings", GetSaleBuildingsHandler(gamedataService))
		r.Get("/sale/buildings/{buildingID}", GetSaleBuildingHandler(gamedataService))
	})
}

// RegisterAdminGamedataRoutes registers the admin-only gamedata import endpoint.
// The router must already have AuthMiddleware, RequireAuth, and RequireAdmin applied.
func RegisterAdminGamedataRoutes(router chi.Router, gamedataService *service.GamedataService) {
	router.Post("/gamedata", PostGamedataHandler(gamedataService))
}
