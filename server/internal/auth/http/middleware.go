package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joanob/yourownboss/internal/auth/crypto"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware validates session tokens and handles automatic renewal with refresh tokens.
// It extracts the session_id and user_id from the JWT and stores them in the request context.
// If the session token is expired, it attempts to renew it using the refresh token.
func AuthMiddleware(jwtManager *crypto.JWTManager, sessionCache *cache.SessionCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract session token from Authorization header or cookie
			sessionToken := extractSessionToken(r)
			if sessionToken == "" {
				log.Debug().Msg("No session token found")
				// Continue without authentication (some endpoints are public)
				next.ServeHTTP(w, r)
				return
			}

			// Validate session token
			claims, err := jwtManager.ValidateSessionToken(sessionToken)
			if err == nil {
				// Token is valid, add to context
				ctx = context.WithValue(ctx, "user_id", claims.UserID)
				ctx = context.WithValue(ctx, "session_id", claims.SessionID)
				ctx = context.WithValue(ctx, "company_id", claims.CompanyID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Session token expired or invalid
			log.Debug().Err(err).Msg("Session token validation failed")

			// Try to renew with refresh token
			refreshToken := extractRefreshToken(r)
			if refreshToken == "" {
				log.Debug().Msg("No refresh token found, cannot renew session")
				next.ServeHTTP(w, r)
				return
			}

			// Validate refresh token
			refreshClaims, err := jwtManager.ValidateRefreshToken(refreshToken)
			if err != nil {
				log.Debug().Err(err).Msg("Refresh token validation failed")
				next.ServeHTTP(w, r)
				return
			}

			// Verify session is still active in cache/DB
			sessionData, exists := sessionCache.Get(refreshClaims.SessionID)
			if !exists {
				log.Debug().Str("session_id", refreshClaims.SessionID).Msg("Session not found in cache")
				next.ServeHTTP(w, r)
				return
			}

			// Generate new session token
			newSessionToken, expiresAt, err := jwtManager.GenerateSessionToken(
				refreshClaims.UserID,
				refreshClaims.SessionID,
				sessionData.CompanyID,
			)
			if err != nil {
				log.Error().Err(err).Str("user_id", refreshClaims.UserID).Msg("Failed to generate new session token")
				next.ServeHTTP(w, r)
				return
			}

			// Set new session token cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    newSessionToken,
				HttpOnly: true,
				Secure:   os.Getenv("ENVIRONMENT") == "production",
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
				MaxAge:   int(expiresAt.Sub(time.Now()).Seconds()),
			})

			log.Debug().Str("user_id", refreshClaims.UserID).Msg("Session token renewed automatically")

			// Add to context and continue
			ctx = context.WithValue(ctx, "user_id", refreshClaims.UserID)
			ctx = context.WithValue(ctx, "session_id", refreshClaims.SessionID)
			ctx = context.WithValue(ctx, "company_id", sessionData.CompanyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth is a middleware that ensures the user is authenticated.
// Returns 401 Unauthorized if no valid session is found.
func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID, ok := ctx.Value("user_id").(string)
			if !ok || userID == "" {
				log.Debug().Msg("Unauthorized request: no user_id in context")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(GenericResponse{
					Error: ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required"},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractSessionToken extracts the session token from the Authorization header or session_token cookie.
func extractSessionToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

// extractRefreshToken extracts the refresh token from the refresh_token cookie.
func extractRefreshToken(r *http.Request) string {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}
