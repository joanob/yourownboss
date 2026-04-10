package auth

import (
    "net/http"
    "os"
)

// AuthMiddleware enforces Authorization: Bearer <AUTH_TOKEN> header on all requests.
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := os.Getenv("AUTH_TOKEN")
        // If no token configured on server, deny access explicitly
        if token == "" {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        auth := r.Header.Get("Authorization")
        if auth != "Bearer "+token {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
