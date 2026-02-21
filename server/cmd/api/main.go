package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"yourownboss/internal/auth"
	"yourownboss/internal/db"
	httpHandlers "yourownboss/internal/http"
	"yourownboss/internal/repository"
	"yourownboss/internal/service"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults and flags")
	}

	// Parse flags
	var (
		port          = flag.String("port", "8080", "Server port")
		dbPath        = flag.String("db", "yourownboss.db", "Database file path")
		jwtSecret     = flag.String("jwt-secret", "", "JWT secret key (if empty, uses default)")
		staticDir     = flag.String("static", "../public", "Static files directory")
		resourcesFile = flag.String("resources", "data/resources.json", "Resources JSON file")
	)
	flag.Parse()

	// Set JWT secret if provided
	if *jwtSecret != "" {
		auth.SetJWTSecret(*jwtSecret)
	} else if envJwtSecret := os.Getenv("JWT_SECRET"); envJwtSecret != "" {
		auth.SetJWTSecret(envJwtSecret)
	} else {
		log.Println("WARNING: Using default JWT secret. Set -jwt-secret in production!")
	}

	// Get initial company money from environment
	initialMoney := int64(0)
	if envMoney := os.Getenv("INITIAL_COMPANY_MONEY"); envMoney != "" {
		if parsed, err := strconv.ParseInt(envMoney, 10, 64); err == nil {
			initialMoney = parsed
			log.Printf("Initial company money set to: %d (from .env)", initialMoney)
		} else {
			log.Printf("WARNING: Invalid INITIAL_COMPANY_MONEY value, using default: %d", initialMoney)
		}
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
	companyRepo := repository.NewCompanyRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	inventoryRepo := repository.NewInventoryRepository(database)

	if err := loadResourcesFromFile(context.Background(), resourceRepo, *resourcesFile); err != nil {
		log.Printf("Warning: failed to load resources: %v", err)
	}

	// Service layer
	authService := service.NewAuthService(userRepo, tokenRepo)
	companyService := service.NewCompanyService(companyRepo, initialMoney)
	inventoryService := service.NewInventoryService(resourceRepo, inventoryRepo)
	marketService := service.NewMarketService(resourceRepo, companyRepo, inventoryRepo)

	// Handler/Controller layer
	authHandler := httpHandlers.NewAuthHandler(authService)
	companyHandler := httpHandlers.NewCompanyHandler(companyService)
	inventoryHandler := httpHandlers.NewInventoryHandler(inventoryService, companyRepo)
	marketHandler := httpHandlers.NewMarketHandler(marketService, companyRepo)

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
		r.Post("/auth/logout", authHandler.Logout)

		// Public inventory routes
		r.Get("/resources", inventoryHandler.GetResources)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(authService))

			r.Get("/auth/me", authHandler.Me)

			// Company routes
			r.Post("/companies", companyHandler.CreateCompany)
			r.Get("/companies/me", companyHandler.GetMyCompany)

			// Inventory routes
			r.Get("/inventory", inventoryHandler.GetInventory)

			// Market routes
			r.Post("/market/buy", marketHandler.BuyResource)
			r.Post("/market/sell", marketHandler.SellResource)
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

type resourceSeed struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	PackSize int64  `json:"pack_size"`
}

func loadResourcesFromFile(ctx context.Context, repo repository.ResourceRepository, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var seeds []resourceSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		return err
	}

	created := 0
	updated := 0
	for _, seed := range seeds {
		if seed.ID <= 0 || seed.Name == "" {
			continue
		}
		if seed.PackSize <= 0 {
			seed.PackSize = 1
		}

		existing, err := repo.GetByID(ctx, seed.ID)
		if err != nil {
			if err == repository.ErrResourceNotFound {
				if _, err := repo.Create(ctx, seed.ID, seed.Name, seed.Price, seed.PackSize); err != nil {
					return err
				}
				created++
				continue
			}
			return err
		}

		if _, err := repo.Update(ctx, existing.ID, seed.Name, seed.Price, seed.PackSize); err != nil {
			return err
		}
		updated++
	}

	if created > 0 {
		log.Printf("Resources loaded: %d", created)
	}
	if updated > 0 {
		log.Printf("Resources updated: %d", updated)
	}

	return nil
}
