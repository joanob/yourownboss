package http

import (
	"encoding/json"
	nativehttp "net/http"

	"github.com/go-playground/validator/v10"
	authdto "github.com/joanob/yourownboss/internal/auth/http"
	"github.com/joanob/yourownboss/internal/users/service"
	"github.com/rs/zerolog/log"
)

// UpdateMeHandler handles updating user profile.
type UpdateMeHandler struct {
	userService service.UserService
	validator   *validator.Validate
}

// NewUpdateMeHandler creates a new update me handler.
func NewUpdateMeHandler(userService service.UserService, validator *validator.Validate) *UpdateMeHandler {
	return &UpdateMeHandler{
		userService: userService,
		validator:   validator,
	}
}

// Handle processes update me requests.
func (h *UpdateMeHandler) Handle(w nativehttp.ResponseWriter, r *nativehttp.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		log.Debug().Msg("UpdateMe request without valid user_id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nativehttp.StatusUnauthorized)
		json.NewEncoder(w).Encode(authdto.GenericResponse{
			Error: authdto.ErrorResponse{Code: "UNAUTHORIZED", Message: "Not authenticated"},
		})
		return
	}

	// Parse request
	var req UpdateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode update me request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nativehttp.StatusBadRequest)
		json.NewEncoder(w).Encode(authdto.GenericResponse{
			Error: authdto.ErrorResponse{Code: "INVALID_INPUT", Message: "Invalid request body"},
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		log.Debug().Err(err).Str("user_id", userID).Msg("UpdateMe validation failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nativehttp.StatusBadRequest)
		json.NewEncoder(w).Encode(authdto.GenericResponse{
			Error: authdto.ErrorResponse{Code: "INVALID_INPUT", Message: "Validation failed"},
		})
		return
	}

	// Create update request
	timezonePtr := req.Timezone
	updateReq := &service.UserUpdateRequest{
		Timezone: &timezonePtr,
	}

	// Call service to update user
	user, err := h.userService.UpdateUser(ctx, userID, updateReq)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to update user")
		w.Header().Set("Content-Type", "application/json")

		// Handle specific errors
		if err.Error() == "timezone_recently_changed" {
			w.WriteHeader(nativehttp.StatusBadRequest)
			json.NewEncoder(w).Encode(authdto.GenericResponse{
				Error: authdto.ErrorResponse{Code: "TIMEZONE_RECENTLY_CHANGED", Message: "Cannot change timezone more than once every 30 days"},
			})
			return
		}

		w.WriteHeader(nativehttp.StatusInternalServerError)
		json.NewEncoder(w).Encode(authdto.GenericResponse{
			Error: authdto.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to update user"},
		})
		return
	}

	// Build response DTO
	timezone := ""
	if user.Timezone != nil {
		timezone = *user.Timezone
	}
	userDTO := &authdto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Timezone:  timezone,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	log.Info().
		Str("user_id", user.ID).
		Str("new_timezone", req.Timezone).
		Msg("User timezone updated")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nativehttp.StatusOK)
	json.NewEncoder(w).Encode(authdto.GenericResponse{Data: userDTO})
}
