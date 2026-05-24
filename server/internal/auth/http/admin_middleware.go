package http

import (
	"encoding/json"
	"net/http"

	"github.com/joanob/yourownboss/internal/users/models"
	"github.com/joanob/yourownboss/internal/users/repository"
	"github.com/rs/zerolog/log"
)

// RequireAdmin is a middleware that ensures the user is authenticated and has admin role.
// Must be used AFTER AuthMiddleware and RequireAuth to ensure user_id is in context.
// Returns 403 Forbidden if user is not an admin.
func RequireAdmin(userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user_id from context (set by RequireAuth)
			userID, ok := ctx.Value("user_id").(string)
			if !ok || userID == "" {
				log.Debug().Msg("RequireAdmin: no user_id in context")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(GenericResponse{
					Error: ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required"},
				})
				return
			}

			// Get user from database to check role
			userDBO, err := userRepo.GetByID(ctx, userID)
			if err != nil {
				log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user for admin check")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(GenericResponse{
					Error: ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to verify admin status"},
				})
				return
			}

			if userDBO == nil {
				log.Debug().Str("user_id", userID).Msg("User not found")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(GenericResponse{
					Error: ErrorResponse{Code: "INSUFFICIENT_PERMISSIONS", Message: "Admin access required"},
				})
				return
			}

			// Convert DBO role string to model Role type and check
			userRole := models.Role(userDBO.Role)
			if !userRole.IsAdmin() {
				log.Warn().Str("user_id", userID).Str("role", string(userRole)).Msg("Non-admin user attempted admin operation")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(GenericResponse{
					Error: ErrorResponse{Code: "INSUFFICIENT_PERMISSIONS", Message: "Admin access required"},
				})
				return
			}

			log.Debug().Str("user_id", userID).Msg("Admin access granted")
			next.ServeHTTP(w, r)
		})
	}
}
