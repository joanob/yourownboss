package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	"github.com/joanob/yourownboss/internal/users/service"
	"github.com/rs/zerolog/log"
)

// RegisterHandler handles user registration.
type RegisterHandler struct {
	userService service.UserService
	validator   *validator.Validate
	rateLimiter *cache.RateLimiter
}

// NewRegisterHandler creates a new register handler.
func NewRegisterHandler(userService service.UserService, validator *validator.Validate, rateLimiter *cache.RateLimiter) *RegisterHandler {
	return &RegisterHandler{
		userService: userService,
		validator:   validator,
		rateLimiter: rateLimiter,
	}
}

// Handle processes user registration requests.
func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Rate limit by IP: max 5 registrations per minute per IP (C-03)
	clientIP := r.RemoteAddr
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		clientIP = realIP
	}
	if !h.rateLimiter.AllowAndRecord(clientIP, "register", 5) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(GenericResponse{
			Error: ErrorResponse{Code: "RATE_LIMIT_EXCEEDED", Message: "Too many registration attempts, please try again later"},
		})
		return
	}

	// Parse request
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode register request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenericResponse{
			Error: ErrorResponse{Code: "INVALID_INPUT", Message: "Invalid request body"},
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		log.Debug().Err(err).Msg("Register validation failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenericResponse{
			Error: ErrorResponse{Code: "INVALID_INPUT", Message: "Validation failed"},
		})
		return
	}

	// Call service
	user, err := h.userService.Register(ctx, req.Username, req.Email, req.Password, req.Timezone)
	if err != nil {
		log.Error().Err(err).Str("username", req.Username).Msg("Registration failed")
		w.Header().Set("Content-Type", "application/json")

		// Handle specific errors
		if err.Error() == "username already exists" {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(GenericResponse{
				Error: ErrorResponse{Code: "USERNAME_ALREADY_EXISTS", Message: "Username is already in use"},
			})
			return
		}
		if err.Error() == "email already exists" {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(GenericResponse{
				Error: ErrorResponse{Code: "EMAIL_ALREADY_EXISTS", Message: "Email is already in use"},
			})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{
			Error: ErrorResponse{Code: "INTERNAL_ERROR", Message: "Registration failed"},
		})
		return
	}

	// Build response DTO
	timezone := ""
	if user.Timezone != nil {
		timezone = *user.Timezone
	}
	userDTO := &UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Timezone:  timezone,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	log.Info().
		Str("user_id", user.ID).
		Str("username", user.Username).
		Msg("User successfully registered")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(GenericResponse{Data: userDTO})
}
