package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	authcrypto "github.com/joanob/yourownboss/internal/auth/crypto"
	authhttphandlers "github.com/joanob/yourownboss/internal/auth/http"
	authrepo "github.com/joanob/yourownboss/internal/auth/repository"
	authsvc "github.com/joanob/yourownboss/internal/auth/service"
	companyhttphandlers "github.com/joanob/yourownboss/internal/company/http"
	companyrepo "github.com/joanob/yourownboss/internal/company/repository"
	companysvc "github.com/joanob/yourownboss/internal/company/service"
	"github.com/joanob/yourownboss/internal/db"
	"github.com/joanob/yourownboss/internal/db/dbqueries"
	gameDataService "github.com/joanob/yourownboss/internal/gamedata/service"
	markethttphandlers "github.com/joanob/yourownboss/internal/market/http"
	marketsvc "github.com/joanob/yourownboss/internal/market/service"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	loggerutil "github.com/joanob/yourownboss/internal/pkg/logger"
	productionhttphandlers "github.com/joanob/yourownboss/internal/production/http"
	productionrepo "github.com/joanob/yourownboss/internal/production/repository"
	productionsvc "github.com/joanob/yourownboss/internal/production/service"
	resourcerepo "github.com/joanob/yourownboss/internal/resources/repository"
	salerepo "github.com/joanob/yourownboss/internal/sale/repository"
	userhttphandlers "github.com/joanob/yourownboss/internal/users/http"
	userrepo "github.com/joanob/yourownboss/internal/users/repository"
	usersvc "github.com/joanob/yourownboss/internal/users/service"
)

func main() {
	// Cargar variables de entorno
	godotenv.Load()

	// Inicializar logger
	loggerutil.InitLogger()
	logger := loggerutil.GetLogger()

	// Obtener configuración
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "0.1.0"
	}

	// Ajustar nivel de logging
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logger.Info().
		Str("port", port).
		Str("log_level", logLevel).
		Str("version", appVersion).
		Msg("Iniciando Your Own Boss Backend - Fase 1.11")

	// Validar JWT_SECRET
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Warn().Msg("JWT_SECRET no configurado. Generando secreto aleatorio para desarrollo")
		jwtSecret = generateRandomSecret(32)
	}
	if len(jwtSecret) < 32 {
		logger.Warn().Msg("JWT_SECRET tiene menos de 32 caracteres. Se recomienda usar al menos 32 caracteres")
	}

	initialCompanyMoney := os.Getenv("INITIAL_COMPANY_MONEY")
	if initialCompanyMoney == "" {
		initialCompanyMoney = "1000"
	}

	// Inicializar base de datos
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "./data/game.db"
	}

	logger.Info().Str("path", dbPath).Msg("Inicializando base de datos...")
	dbConn, err := db.InitDatabase(dbPath)
	if err != nil {
		logger.Error().Err(err).Msg("Error al inicializar base de datos")
		os.Exit(1)
	}
	defer dbConn.Close()

	logger.Info().Msg("Base de datos inicializada correctamente")

	// Inicializar caches
	gamedataCache := cache.NewGamedataCache()
	sessionCache := cache.NewSessionCache()

	// Iniciar limpiador de sesiones (cada 6 horas)
	sessionCleanupInterval := os.Getenv("SESSION_CACHE_CLEANUP_INTERVAL")
	if sessionCleanupInterval == "" {
		sessionCleanupInterval = "21600" // 6 horas
	}

	cleanupIntervalSecs, _ := strconv.ParseInt(sessionCleanupInterval, 10, 64)
	go startSessionCleanupRoutine(sessionCache, time.Duration(cleanupIntervalSecs)*time.Second)

	logger.Info().Msg("Limpiador de sesiones iniciado")

	// ============================================================================
	// INYECCIÓN DE DEPENDENCIAS - FASE 1.11
	// ============================================================================

	logger.Info().Msg("Inicializando dependencias...")

	// Crear validator
	validate := validator.New()

	// Crear queries
	queries := dbqueries.New(dbConn)

	// Crear JWT Manager
	jwtManager, err := authcrypto.NewJWTManager()
	if err != nil {
		logger.Error().Err(err).Msg("Error al crear JWT Manager")
		os.Exit(1)
	}

	// Crear Password Manager
	passwordManager := authcrypto.NewPasswordManager()

	// Crear Repositories
	userRepository := userrepo.NewUserRepository(queries)
	sessionRepository := authrepo.NewUserSessionRepository(queries)
	companyRepository := companyrepo.NewCompanyRepository(queries)
	inventoryRepository := companyrepo.NewInventoryRepository(queries)

	// Gamedata repositories (Phase 3)
	resourceRepository := resourcerepo.NewResourceRepository(queries)
	productionBuildingRepository := productionrepo.NewProductionBuildingRepository(queries)
	productionProcessRepository := productionrepo.NewProductionProcessRepository(queries)
	saleBuildingRepository := salerepo.NewSaleBuildingRepository(queries)
	saleResourceRepository := salerepo.NewSaleResourceRepository(queries)

	// Production company repositories (Phase 5)
	companyBuildingRepository := productionrepo.NewCompanyBuildingRepository(queries)
	productionRunRepository := productionrepo.NewProductionRunRepository(queries)

	// Crear Services
	userService := usersvc.NewUserService(userRepository, passwordManager)
	authService := authsvc.NewAuthService(userRepository, sessionRepository, passwordManager, jwtManager, sessionCache)
	companyService := companysvc.NewCompanyService(companyRepository, inventoryRepository)
	inventoryService := companysvc.NewInventoryService(inventoryRepository, companyRepository)

	// Market service (Phase 4)
	marketService := marketsvc.NewMarketService(companyRepository, inventoryRepository, gamedataCache)

	// Production service (Phase 5)
	productionService := productionsvc.NewProductionService(
		companyRepository,
		inventoryRepository,
		companyBuildingRepository,
		productionRunRepository,
		gamedataCache,
	)

	// Gamedata service (Phase 3)
	gamedataSvc := gameDataService.NewGamedataService(
		resourceRepository,
		productionBuildingRepository,
		productionProcessRepository,
		saleBuildingRepository,
		saleResourceRepository,
		gamedataCache,
	)

	// Load and cache gamedata
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := gamedataSvc.RefreshCache(ctx); err != nil {
		logger.Error().Err(err).Msg("Error al cargar datos maestros en cache")
		os.Exit(1)
	}
	logger.Info().Msg("Datos maestros cargados en cache")

	// ============================================================================
	// CREAR ADMIN SI NO EXISTE
	// ============================================================================

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		logger.Warn().Msg("ADMIN_PASSWORD no configurado. Admin no será creado automáticamente")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Verificar si admin ya existe
		adminUser, _ := userRepository.GetByUsername(ctx, "admin")
		if adminUser == nil {
			logger.Info().Msg("Creando usuario admin...")
			_, err := userService.Register(ctx, "admin", "admin@yourownboss.local", adminPassword, "UTC")
			if err != nil {
				logger.Error().Err(err).Msg("Error al crear usuario admin")
				os.Exit(1)
			}
			logger.Info().Msg("Usuario admin creado exitosamente")
		} else {
			logger.Info().Msg("Usuario admin ya existe")
		}
	}

	// ============================================================================
	// CREAR ROUTER Y REGISTRAR RUTAS
	// ============================================================================

	logger.Info().Msg("Registrando rutas...")

	r := chi.NewRouter()

	// Middleware global
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS middleware
	r.Use(corsMiddleware())

	// Timeout middleware (30 segundos)
	r.Use(middleware.Timeout(30 * time.Second))

	// Rutas de salud
	r.Get("/api/v1/status", healthHandler)
	r.Get("/health", healthHandler)

	// Placeholder para root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Your Own Boss API",
			"version": "0.1.0",
			"status":  "ok",
		})
	})

	// Registrar rutas de autenticación
	authhttphandlers.RegisterAuthRoutes(r, userService, authService, validate)
	logger.Debug().Msg("Rutas de autenticación registradas")

	// Registrar rutas de usuarios (requiere autenticación)
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Use(authhttphandlers.AuthMiddleware(jwtManager, sessionCache))
		userhttphandlers.RegisterUsersRoutes(r, userService, validate)
	})
	logger.Debug().Msg("Rutas de usuarios registradas")

	// Registrar rutas de company, market y production (requieren autenticación)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authhttphandlers.AuthMiddleware(jwtManager, sessionCache))
		companyhttphandlers.RegisterCompanyRoutes(r, companyService, inventoryService)
		markethttphandlers.RegisterMarketRoutes(r, marketService)
		productionhttphandlers.RegisterProductionRoutes(r, productionService)
	})
	logger.Debug().Msg("Rutas de company, market y production registradas")

	logger.Info().Msg("Rutas registradas exitosamente")

	// ============================================================================
	// INICIAR SERVIDOR HTTP
	// ============================================================================

	addr := fmt.Sprintf(":%s", port)
	logger.Info().
		Str("addr", addr).
		Str("version", appVersion).
		Msg("Servidor escuchando")

	if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Error al iniciar servidor")
	}
}

// generateRandomSecret generates a random secret of the specified length
func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// healthHandler devuelve el estado del servidor
func healthHandler(w http.ResponseWriter, r *http.Request) {
	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "0.1.0"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"status":    "ok",
			"version":   appVersion,
			"timestamp": time.Now().UTC(),
		},
	})
}

// corsMiddleware agrega headers CORS
func corsMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
			if allowedOrigins == "" {
				allowedOrigins = "http://localhost:3000"
			}

			origin := r.Header.Get("Origin")
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// startSessionCleanupRoutine inicia la limpieza periódica de sesiones expiradas
func startSessionCleanupRoutine(sc *cache.SessionCache, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger := loggerutil.GetLogger()

	for range ticker.C {
		deleted := sc.CleanupExpired()
		logger.Info().
			Int("deleted_sessions", deleted).
			Int("remaining_sessions", sc.Count()).
			Msg("Limpieza de sesiones completada")
	}
}
