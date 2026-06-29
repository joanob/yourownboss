package http

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joanob/yourownboss/internal/auth/service"
	appvalidator "github.com/joanob/yourownboss/internal/pkg/validator"
	"github.com/rs/zerolog/log"
)

// LoginHandler handles user login.
type LoginHandler struct {
	authService service.AuthService
	validator   *appvalidator.Validator
}

// NewLoginHandler creates a new login handler.
func NewLoginHandler(authService service.AuthService, validator *appvalidator.Validator) *LoginHandler {
	return &LoginHandler{
		authService: authService,
		validator:   validator,
	}
}

// Handle processes login requests.
func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode login request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenericResponse{
			Error: ErrorResponse{Code: "INVALID_INPUT", Message: "Invalid request body"},
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		log.Debug().Err(err).Str("username", req.Username).Msg("Login validation failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenericResponse{
			Error: ErrorResponse{Code: "INVALID_INPUT", Message: "Validation failed"},
		})
		return
	}

	// Call auth service
	loginResp, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		log.Debug().Err(err).Str("username", req.Username).Msg("Login failed")
		w.Header().Set("Content-Type", "application/json")
		// SEC-07: too_many_attempts devuelve 429, no 401
		if strings.Contains(err.Error(), "too_many_attempts") {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(GenericResponse{
				Error: ErrorResponse{Code: "RATE_LIMIT_EXCEEDED", Message: "Too many failed login attempts, please try again later"},
			})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(GenericResponse{
				Error: ErrorResponse{Code: "INVALID_CREDENTIALS", Message: "Invalid username or password"},
			})
		}
		return
	}

	// Build user DTO
	timezone := ""
	if loginResp.User.Timezone != nil {
		timezone = *loginResp.User.Timezone
	}
	userDTO := &UserDTO{
		ID:        loginResp.User.ID,
		Username:  loginResp.User.Username,
		Email:     loginResp.User.Email,
		Timezone:  timezone,
		CreatedAt: loginResp.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	secureCookie := os.Getenv("ENVIRONMENT") == "production"

	// Set session token cookie (httpOnly, 1 minute)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    loginResp.SessionToken,
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   60, // 1 minute
	})

	// Set refresh token cookie (httpOnly, 300 days)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    loginResp.RefreshToken,
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int((300 * 24 * time.Hour).Seconds()),
	})

	log.Info().
		Str("user_id", loginResp.User.ID).
		Str("username", req.Username).
		Msg("User successfully logged in")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(GenericResponse{
		Data: LoginResponse{
			User:             userDTO,
			SessionExpiresAt: loginResp.ExpiresAt,
		},
	})
}
