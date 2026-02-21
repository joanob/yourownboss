package auth

import (
	"context"
	"net/http"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
)

const (
	AccessTokenCookieName  = "yourownboss_access"
	RefreshTokenCookieName = "yourownboss_refresh"
)

// AuthService interface for dependency injection
type AuthService interface {
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
}

// RequireAuth is a middleware that validates the access token from cookies
// and automatically refreshes it if expired/invalid and a valid refresh token exists
func RequireAuth(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get access token from cookie first
			cookie, err := r.Cookie(AccessTokenCookieName)
			var tokenString string

			if err == nil {
				tokenString = cookie.Value
			}

			// Validate access token if present
			var claims *Claims
			var tokenErr error
			if tokenString != "" {
				claims, tokenErr = ValidateAccessToken(tokenString)
			} else {
				tokenErr = ErrInvalidToken
			}

			// If access token is invalid/expired/missing
			if tokenErr != nil {
				refreshCookie, refreshErr := r.Cookie(RefreshTokenCookieName)
				if refreshErr == nil {
					// Try to refresh using the refresh token
					newAccessToken, refreshErr := authService.RefreshAccessToken(r.Context(), refreshCookie.Value)
					if refreshErr == nil {
						// Update the access token cookie
						http.SetCookie(w, &http.Cookie{
							Name:     AccessTokenCookieName,
							Value:    newAccessToken,
							Path:     "/",
							HttpOnly: true,
							Secure:   true,
							SameSite: http.SameSiteLaxMode,
						})

						// Parse the new token to get claims
						claims, tokenErr = ValidateAccessToken(newAccessToken)
						if tokenErr != nil {
							http.Error(w, "failed to validate refreshed token", http.StatusUnauthorized)
							return
						}

						// Continue with the new token
						ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
						ctx = context.WithValue(ctx, UsernameKey, claims.Username)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// If we still don't have valid claims, return unauthorized
			if tokenErr != nil {
				if tokenErr == ErrExpiredToken {
					http.Error(w, "token expired", http.StatusUnauthorized)
				} else {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
				}
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
