package http

import (
	"encoding/json"
	"net/http"

	"yourownboss/internal/auth"
	"yourownboss/internal/service"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Request/Response DTOs
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		respondError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// Call service
	result, err := h.authService.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			respondError(w, "username already exists", http.StatusConflict)
		case service.ErrWeakPassword:
			respondError(w, err.Error(), http.StatusBadRequest)
		default:
			respondError(w, "failed to register user", http.StatusInternalServerError)
		}
		return
	}

	// Set cookies
	setAuthCookies(w, result.AccessToken, result.RefreshToken)

	// Send response
	respondJSON(w, toAuthResponse(result), http.StatusCreated)
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		respondError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// Call service
	result, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			respondError(w, "invalid username or password", http.StatusUnauthorized)
		} else {
			respondError(w, "failed to login", http.StatusInternalServerError)
		}
		return
	}

	// Set cookies
	setAuthCookies(w, result.AccessToken, result.RefreshToken)

	// Send response
	respondJSON(w, toAuthResponse(result), http.StatusOK)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie(auth.RefreshTokenCookieName)
	if err == nil {
		// Call service to revoke token
		_ = h.authService.Logout(r.Context(), cookie.Value)
	}

	// Clear cookies
	clearAuthCookies(w)

	respondJSON(w, map[string]string{"message": "logged out successfully"}, http.StatusOK)
}

// Me returns the current user's information
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, "user not found", http.StatusNotFound)
		return
	}

	var resp AuthResponse
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	respondJSON(w, resp, http.StatusOK)
}

// Helper functions

func toAuthResponse(result *service.AuthResult) AuthResponse {
	var resp AuthResponse
	resp.User.ID = result.User.ID
	resp.User.Username = result.User.Username
	return resp
}

func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	// Set access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     auth.AccessTokenCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.AccessTokenDuration.Seconds()),
	})

	// Set refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     auth.RefreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.RefreshTokenDuration.Seconds()),
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.AccessTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     auth.RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, status int) {
	respondJSON(w, ErrorResponse{Error: message}, status)
}
