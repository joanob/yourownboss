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
		buildingsFile = flag.String("production-buildings", "data/production_buildings.json", "Production buildings JSON file")
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
	productionBuildingRepo := repository.NewProductionBuildingRepository(database)
	productionProcessRepo := repository.NewProductionProcessRepository(database)
	processResourceRepo := repository.NewProductionProcessResourceRepository(database)

	if err := loadResourcesFromFile(context.Background(), resourceRepo, *resourcesFile); err != nil {
		log.Printf("Warning: failed to load resources: %v", err)
	}

	if err := loadProductionBuildingsFromFile(
		context.Background(),
		productionBuildingRepo,
		productionProcessRepo,
		processResourceRepo,
		resourceRepo,
		*buildingsFile,
	); err != nil {
		log.Printf("Warning: failed to load production buildings: %v", err)
	}

	// Service layer
	authService := service.NewAuthService(userRepo, tokenRepo)
	companyService := service.NewCompanyService(companyRepo, initialMoney)
	inventoryService := service.NewInventoryService(resourceRepo, inventoryRepo)
	marketService := service.NewMarketService(resourceRepo, companyRepo, inventoryRepo)
	productionService := service.NewProductionService(
		productionBuildingRepo,
		productionProcessRepo,
		processResourceRepo,
		resourceRepo,
	)

	// Handler/Controller layer
	authHandler := httpHandlers.NewAuthHandler(authService)
	companyHandler := httpHandlers.NewCompanyHandler(companyService)
	inventoryHandler := httpHandlers.NewInventoryHandler(inventoryService, companyRepo)
	marketHandler := httpHandlers.NewMarketHandler(marketService, companyRepo)
	productionHandler := httpHandlers.NewProductionHandler(productionService)

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
		r.Get("/production-buildings", productionHandler.GetProductionBuildings)

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

type productionBuildingSeed struct {
	ID        int64                   `json:"id"`
	Name      string                  `json:"name"`
	Cost      int64                   `json:"cost"`
	Processes []productionProcessSeed `json:"processes"`
}

type productionProcessSeed struct {
	ID               int64                     `json:"id"`
	Name             string                    `json:"name"`
	ProcessingTimeMs int64                     `json:"processing_time_ms"`
	TimeWindow       *productionTimeWindowSeed `json:"time_window"`
	Resources        []processResourceSeed     `json:"resources"`
}

type processResourceSeed struct {
	ResourceID int64  `json:"resource_id"`
	Direction  string `json:"direction"`
	Quantity   int64  `json:"quantity"`
}

type productionTimeWindowSeed struct {
	StartHour int64 `json:"start_hour"`
	EndHour   int64 `json:"end_hour"`
}

func loadProductionBuildingsFromFile(
	ctx context.Context,
	buildingRepo repository.ProductionBuildingRepository,
	processRepo repository.ProductionProcessRepository,
	processResourceRepo repository.ProductionProcessResourceRepository,
	resourceRepo repository.ResourceRepository,
	path string,
) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var seeds []productionBuildingSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		return err
	}

	created := 0
	updated := 0
	processesCreated := 0
	processesUpdated := 0
	processResourcesCreated := 0
	processResourcesUpdated := 0
	processResourcesDeleted := 0
	for _, seed := range seeds {
		if seed.ID <= 0 || seed.Name == "" {
			continue
		}

		existing, err := buildingRepo.GetByID(ctx, seed.ID)
		if err != nil {
			if err == repository.ErrProductionBuildingNotFound {
				if _, err := buildingRepo.Create(ctx, seed.ID, seed.Name, seed.Cost); err != nil {
					return err
				}
				created++
			} else {
				return err
			}
		} else {
			if _, err := buildingRepo.Update(ctx, existing.ID, seed.Name, seed.Cost); err != nil {
				return err
			}
			updated++
		}

		for _, processSeed := range seed.Processes {
			if processSeed.ID <= 0 || processSeed.Name == "" || processSeed.ProcessingTimeMs <= 0 {
				continue
			}

			var windowStartHour *int64
			var windowEndHour *int64
			if processSeed.TimeWindow != nil {
				if processSeed.TimeWindow.StartHour < 0 || processSeed.TimeWindow.StartHour > 23 {
					continue
				}
				if processSeed.TimeWindow.EndHour < 0 || processSeed.TimeWindow.EndHour > 23 {
					continue
				}
				if processSeed.TimeWindow.StartHour >= processSeed.TimeWindow.EndHour {
					continue
				}
				windowStartHour = &processSeed.TimeWindow.StartHour
				windowEndHour = &processSeed.TimeWindow.EndHour
			}

			existingProcess, err := processRepo.GetByID(ctx, processSeed.ID)
			if err != nil {
				if err == repository.ErrProductionProcessNotFound {
					if _, err := processRepo.Create(
						ctx,
						processSeed.ID,
						processSeed.Name,
						processSeed.ProcessingTimeMs,
						seed.ID,
						windowStartHour,
						windowEndHour,
					); err != nil {
						return err
					}
					processesCreated++
				} else {
					return err
				}
			} else {
				if _, err := processRepo.Update(
					ctx,
					existingProcess.ID,
					processSeed.Name,
					processSeed.ProcessingTimeMs,
					seed.ID,
					windowStartHour,
					windowEndHour,
				); err != nil {
					return err
				}
				processesUpdated++
			}

			existingResources, err := processResourceRepo.GetAllByProcess(ctx, processSeed.ID)
			if err != nil {
				return err
			}

			existingByKey := make(map[string]db.ProductionProcessResource, len(existingResources))
			for _, existingResource := range existingResources {
				key := fmt.Sprintf("%d|%s", existingResource.ResourceID, existingResource.Direction)
				existingByKey[key] = existingResource
			}

			seen := make(map[string]struct{}, len(processSeed.Resources))
			for _, resourceSeed := range processSeed.Resources {
				if resourceSeed.ResourceID <= 0 || resourceSeed.Quantity <= 0 {
					continue
				}
				if resourceSeed.Direction != "input" && resourceSeed.Direction != "output" {
					continue
				}

				if _, err := resourceRepo.GetByID(ctx, resourceSeed.ResourceID); err != nil {
					if err == repository.ErrResourceNotFound {
						continue
					}
					return err
				}

				key := fmt.Sprintf("%d|%s", resourceSeed.ResourceID, resourceSeed.Direction)
				if existing, ok := existingByKey[key]; ok {
					if existing.Quantity != resourceSeed.Quantity {
						processResourcesUpdated++
					}
				} else {
					processResourcesCreated++
				}

				if err := processResourceRepo.Upsert(
					ctx,
					processSeed.ID,
					resourceSeed.ResourceID,
					resourceSeed.Direction,
					resourceSeed.Quantity,
				); err != nil {
					return err
				}
				seen[key] = struct{}{}
			}

			for _, existingResource := range existingResources {
				key := fmt.Sprintf("%d|%s", existingResource.ResourceID, existingResource.Direction)
				if _, ok := seen[key]; ok {
					continue
				}
				if err := processResourceRepo.Delete(
					ctx,
					existingResource.ProcessID,
					existingResource.ResourceID,
					existingResource.Direction,
				); err != nil {
					return err
				}
				processResourcesDeleted++
			}

		}
	}

	if created > 0 {
		log.Printf("Production buildings loaded: %d", created)
	}
	if updated > 0 {
		log.Printf("Production buildings updated: %d", updated)
	}
	if processesCreated > 0 {
		log.Printf("Production processes loaded: %d", processesCreated)
	}
	if processesUpdated > 0 {
		log.Printf("Production processes updated: %d", processesUpdated)
	}
	if processResourcesCreated > 0 {
		log.Printf("Production process resources loaded: %d", processResourcesCreated)
	}
	if processResourcesUpdated > 0 {
		log.Printf("Production process resources updated: %d", processResourcesUpdated)
	}
	if processResourcesDeleted > 0 {
		log.Printf("Production process resources removed: %d", processResourcesDeleted)
	}

	return nil
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
