package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/sale/service"
)

// RegisterSaleRoutes registers all sale building routes on the provided router.
// All routes expect an authenticated context (company_id in context).
func RegisterSaleRoutes(r chi.Router, svc *service.SaleService) {
	r.Route("/company/sale/buildings", func(r chi.Router) {
		r.Get("/", GetSaleBuildingsHandler(svc))
		r.Post("/", BuildSaleBuildingHandler(svc))

		r.Route("/{id}", func(r chi.Router) {
			r.Post("/upgrade", UpgradeSaleBuildingHandler(svc))
			r.Post("/start", StartSaleHandler(svc))
			r.Post("/collect", CollectSaleHandler(svc))
		})
	})
}
