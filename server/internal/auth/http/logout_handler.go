package http

import (
	"encoding/json"
	"net/http"

	"github.com/joanob/yourownboss/internal/auth/service"
	"github.com/rs/zerolog/log"
)

// LogoutHandler handles user logout.
type LogoutHandler struct {
	authService service.AuthService
}

// NewLogoutHandler creates a new logout handler.
func NewLogoutHandler(authService service.AuthService) *LogoutHandler {
	return &LogoutHandler{
		authService: authService,
	}
}

// Handle processes logout requests.
func (h *LogoutHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session ID from context (set by auth middleware)
	sessionID, ok := ctx.Value("session_id").(string)
	if !ok || sessionID == "" {
		log.Debug().Msg("Logout attempt without valid session")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(GenericResponse{
			Error: ErrorResponse{Code: "UNAUTHORIZED", Message: "No active session"},
		})
		return
	}

	// Call auth service to revoke session
	if err := h.authService.Logout(ctx, sessionID); err != nil {
		log.Error().Err(err).Str("session_id", sessionID).Msg("Logout failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{
			Error: ErrorResponse{Code: "INTERNAL_ERROR", Message: "Logout failed"},
		})
		return
	}

	// Clear session token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // Delete cookie
	})

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // Delete cookie
	})

	log.Info().Str("session_id", sessionID).Msg("User successfully logged out")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Data: "logged out"})
}
