package http

import (
	"encoding/json"
	nativehttp "net/http"

	authdto "github.com/joanob/yourownboss/internal/auth/http"
	"github.com/joanob/yourownboss/internal/users/service"
	"github.com/rs/zerolog/log"
)

// GetMeHandler handles getting user profile.
type GetMeHandler struct {
	userService service.UserService
}

// NewGetMeHandler creates a new get me handler.
func NewGetMeHandler(userService service.UserService) *GetMeHandler {
	return &GetMeHandler{
		userService: userService,
	}
}

// Handle processes get me requests.
func (h *GetMeHandler) Handle(w nativehttp.ResponseWriter, r *nativehttp.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		log.Debug().Msg("GetMe request without valid user_id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nativehttp.StatusUnauthorized)
		json.NewEncoder(w).Encode(authdto.GenericResponse{
			Error: authdto.ErrorResponse{Code: "UNAUTHORIZED", Message: "Not authenticated"},
		})
		return
	}

	// Get user from service
	user, err := h.userService.GetUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nativehttp.StatusInternalServerError)
		json.NewEncoder(w).Encode(authdto.GenericResponse{
			Error: authdto.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to get user"},
		})
		return
	}

	if user == nil {
		log.Debug().Str("user_id", userID).Msg("User not found")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nativehttp.StatusNotFound)
		json.NewEncoder(w).Encode(authdto.GenericResponse{
			Error: authdto.ErrorResponse{Code: "USER_NOT_FOUND", Message: "User not found"},
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nativehttp.StatusOK)
	json.NewEncoder(w).Encode(authdto.GenericResponse{Data: userDTO})
}
