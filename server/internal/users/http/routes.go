package http

import (
	"github.com/go-chi/chi/v5"
	appvalidator "github.com/joanob/yourownboss/internal/pkg/validator"
	"github.com/joanob/yourownboss/internal/users/service"
)

// RegisterUsersRoutes registers all user-related routes.
func RegisterUsersRoutes(
	router chi.Router,
	userService service.UserService,
	validator *appvalidator.Validator,
) {
	getMeHandler := NewGetMeHandler(userService)
	updateMeHandler := NewUpdateMeHandler(userService, validator)

	router.Get("/api/v1/users/me", getMeHandler.Handle)
	router.Put("/api/v1/users/me", updateMeHandler.Handle)
}
