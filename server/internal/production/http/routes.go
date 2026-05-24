package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/production/service"
)

// RegisterProductionRoutes registers all production building routes on the provided router.
// All routes expect an authenticated context (company_id in context).
func RegisterProductionRoutes(r chi.Router, svc *service.ProductionService) {
	r.Route("/company/production/buildings", func(r chi.Router) {
		r.Get("/", GetBuildingsHandler(svc))
		r.Post("/", BuildProductionBuildingHandler(svc))

		r.Route("/{id}", func(r chi.Router) {
			r.Post("/upgrade", UpgradeBuildingHandler(svc))
			r.Post("/start", StartProductionHandler(svc))
			r.Post("/collect", CollectProductionHandler(svc))
		})
	})
}
