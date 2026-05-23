package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// ResponseWriter is a custom response writer that tracks status code and response size.
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader captures the HTTP status code.
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// LoggingMiddleware logs all HTTP requests with method, path, user ID, status code, and duration.
// Format: method=GET path=/api/v1/users/me user_id=abc123 status=200 duration_ms=45
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Serve the request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)
			durationMs := duration.Milliseconds()

			// Extract user ID from context if available
			userID := "anonymous"
			if uid, ok := r.Context().Value("user_id").(string); ok && uid != "" {
				userID = uid
			}

			// Log the request
			log.Info().
				Str("method", r.Method).
				Str("path", r.RequestURI).
				Str("user_id", userID).
				Int("status", wrapped.statusCode).
				Int64("duration_ms", durationMs).
				Int("response_size", wrapped.size).
				Msg("HTTP request")
		})
	}
}

// RecoveryMiddleware recovers from panics and logs them as errors.
// Returns 500 Internal Server Error if a panic occurs.
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					log.Error().
						Interface("panic", err).
						Str("method", r.Method).
						Str("path", r.RequestURI).
						Msg("Request panic recovered")

					// Return 500 error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error":{"code":"INTERNAL_ERROR","message":"An unexpected error occurred"}}`)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware configures CORS headers for cross-origin requests.
// Allows requests from the specified frontend origin.
func CORSMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestTimeoutMiddleware sets a timeout for all requests.
// Default timeout is 30 seconds as per specification.
func RequestTimeoutMiddleware(timeoutSeconds int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds)*time.Second)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
