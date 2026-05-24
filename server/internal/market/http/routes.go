package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/market/service"
)

// RegisterMarketRoutes registers market routes under the given router.
// The router should already have authentication middleware applied.
func RegisterMarketRoutes(router chi.Router, marketService *service.MarketService) {
	router.Post("/market/buy", BuyResourceHandler(marketService))
	router.Post("/market/sell", SellResourceHandler(marketService))
}
