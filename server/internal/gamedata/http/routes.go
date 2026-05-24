package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/gamedata/service"
)

// RegisterGamedataRoutes registers all gamedata-related routes
func RegisterGamedataRoutes(router chi.Router, gamedataService *service.GamedataService) {
	router.Route("/api/v1", func(r chi.Router) {
		// Public gamedata endpoints
		r.Get("/gamedata", GetGamedataHandler(gamedataService))
		r.Get("/resources", GetResourcesHandler(gamedataService))
		r.Get("/resources/{resourceID}", GetResourceHandler(gamedataService))
		r.Get("/production/buildings", GetProductionBuildingsHandler(gamedataService))
		r.Get("/production/buildings/{buildingID}", GetProductionBuildingHandler(gamedataService))
		r.Get("/sale/buildings", GetSaleBuildingsHandler(gamedataService))
		r.Get("/sale/buildings/{buildingID}", GetSaleBuildingHandler(gamedataService))

		// Admin-only endpoints - will have admin middleware applied
		r.Post("/gamedata", PostGamedataHandler(gamedataService))
	})
}
