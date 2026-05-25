package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/joanob/yourownboss/internal/market/service"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

// RegisterMarketRoutes registers market routes under the given router.
// The router should already have authentication middleware applied.
func RegisterMarketRoutes(router chi.Router, marketService *service.MarketService, rateLimiter *cache.RateLimiter) {
	router.Post("/market/buy", BuyResourceHandler(marketService, rateLimiter))
	router.Post("/market/sell", SellResourceHandler(marketService, rateLimiter))
}
