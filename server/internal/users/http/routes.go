package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joanob/yourownboss/internal/users/service"
)

// RegisterUsersRoutes registers all user-related routes.
func RegisterUsersRoutes(
	router chi.Router,
	userService service.UserService,
	validator *validator.Validate,
) {
	getMeHandler := NewGetMeHandler(userService)
	updateMeHandler := NewUpdateMeHandler(userService, validator)

	router.Get("/api/v1/users/me", getMeHandler.Handle)
	router.Put("/api/v1/users/me", updateMeHandler.Handle)
}
