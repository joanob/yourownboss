package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
)

// RequireAuth is a middleware that validates the access token from cookies
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get token from cookie first
		cookie, err := r.Cookie("access_token")
		var tokenString string

		if err == nil {
			tokenString = cookie.Value
		} else {
			// Fallback to Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenString = parts[1]
		}

		// Validate token
		claims, err := ValidateAccessToken(tokenString)
		if err != nil {
			if err == ErrExpiredToken {
				http.Error(w, "token expired", http.StatusUnauthorized)
				return
			}
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetUsernameFromContext retrieves the username from the request context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameKey).(string)
	return username, ok
}
