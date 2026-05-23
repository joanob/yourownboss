package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	authsvc "github.com/joanob/yourownboss/internal/auth/service"
	usersvc "github.com/joanob/yourownboss/internal/users/service"
)

// RegisterAuthRoutes registers all auth-related routes.
func RegisterAuthRoutes(
	router chi.Router,
	userService usersvc.UserService,
	authService authsvc.AuthService,
	validator *validator.Validate,
) {
	registerHandler := NewRegisterHandler(userService, validator)
	loginHandler := NewLoginHandler(authService, validator)
	logoutHandler := NewLogoutHandler(authService)

	router.Post("/api/v1/auth/register", registerHandler.Handle)
	router.Post("/api/v1/auth/login", loginHandler.Handle)
	router.Post("/api/v1/auth/logout", logoutHandler.Handle)
}
