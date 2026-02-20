package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"yourownboss/internal/auth"
	"yourownboss/internal/db"
	httpHandlers "yourownboss/internal/http"
	"yourownboss/internal/repository"
	"yourownboss/internal/service"
)

func main() {
	// Parse flags
	var (
		port      = flag.String("port", "8080", "Server port")
		dbPath    = flag.String("db", "yourownboss.db", "Database file path")
		jwtSecret = flag.String("jwt-secret", "", "JWT secret key (if empty, uses default)")
		staticDir = flag.String("static", "../public", "Static files directory")
	)
	flag.Parse()

	// Set JWT secret if provided
	if *jwtSecret != "" {
		auth.SetJWTSecret(*jwtSecret)
	} else {
		log.Println("WARNING: Using default JWT secret. Set -jwt-secret in production!")
	}

	// Open database
	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	log.Printf("Database initialized: %s", *dbPath)

	// Initialize layers (Dependency Injection)
	// Repository layer
	userRepo := repository.NewUserRepository(database)
	tokenRepo := repository.NewTokenRepository(database)

	// Service layer
	authService := service.NewAuthService(userRepo, tokenRepo)

	// Handler/Controller layer
	authHandler := httpHandlers.NewAuthHandler(authService)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes (public)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)
		r.Post("/auth/logout", authHandler.Logout)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth)
			r.Get("/auth/me", authHandler.Me)
		})
	})

	// Serve static files (frontend)
	staticPath, err := filepath.Abs(*staticDir)
	if err != nil {
		log.Printf("Warning: Could not resolve static path: %v", err)
	} else if _, err := os.Stat(staticPath); err == nil {
		log.Printf("Serving static files from: %s", staticPath)
		fileServer := http.FileServer(http.Dir(staticPath))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Check if file exists
			path := filepath.Join(staticPath, r.URL.Path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// Serve index.html for client-side routing
				http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	} else {
		log.Printf("Warning: Static directory not found: %s", staticPath)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("API server running. Frontend not built yet."))
		})
	}

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
