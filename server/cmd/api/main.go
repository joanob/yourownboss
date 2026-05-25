package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	auditrepo "github.com/joanob/yourownboss/internal/audit/repository"
	authcrypto "github.com/joanob/yourownboss/internal/auth/crypto"
	authhttphandlers "github.com/joanob/yourownboss/internal/auth/http"
	authrepo "github.com/joanob/yourownboss/internal/auth/repository"
	authsvc "github.com/joanob/yourownboss/internal/auth/service"
	companyhttphandlers "github.com/joanob/yourownboss/internal/company/http"
	companyrepo "github.com/joanob/yourownboss/internal/company/repository"
	companysvc "github.com/joanob/yourownboss/internal/company/service"
	"github.com/joanob/yourownboss/internal/db"
	"github.com/joanob/yourownboss/internal/db/dbqueries"
	gameDataHttp "github.com/joanob/yourownboss/internal/gamedata/http"
	gameDataService "github.com/joanob/yourownboss/internal/gamedata/service"
	markethttphandlers "github.com/joanob/yourownboss/internal/market/http"
	marketsvc "github.com/joanob/yourownboss/internal/market/service"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	loggerutil "github.com/joanob/yourownboss/internal/pkg/logger"
	productionhttphandlers "github.com/joanob/yourownboss/internal/production/http"
	productionrepo "github.com/joanob/yourownboss/internal/production/repository"
	productionsvc "github.com/joanob/yourownboss/internal/production/service"
	resourcerepo "github.com/joanob/yourownboss/internal/resources/repository"
	salehttphandlers "github.com/joanob/yourownboss/internal/sale/http"
	salerepo "github.com/joanob/yourownboss/internal/sale/repository"
	salesvc "github.com/joanob/yourownboss/internal/sale/service"
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
		logger.Fatal().Msg("JWT_SECRET debe tener al menos 32 caracteres. Corrija la configuración y reinicie")
		os.Exit(1)
	}

	initialCompanyMoney := os.Getenv("INITIAL_COMPANY_MONEY")
	if initialCompanyMoney == "" {
		initialCompanyMoney = "1000"
	}
	initialCompanyMoneyInt, err := strconv.ParseInt(initialCompanyMoney, 10, 64)
	if err != nil || initialCompanyMoneyInt <= 0 {
		logger.Warn().Str("value", initialCompanyMoney).Msg("INITIAL_COMPANY_MONEY inválido. Usando 1000 por defecto")
		initialCompanyMoneyInt = 1000
	}

	// Root context for graceful shutdown (B-01)
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

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
	loginAttemptRepository := authrepo.NewLoginAttemptRepository(queries)
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

	// Sale company repositories (Phase 6)
	companySaleBuildingRepository := salerepo.NewCompanySaleBuildingRepository(queries)
	saleRunRepository := salerepo.NewSaleRunRepository(queries)

	// Audit repository (Phase 7)
	auditRepository := auditrepo.NewAuditRepository(queries)

	// Rate limiter (Phase 8 — market/production anti-abuse)
	rateLimiter := cache.NewRateLimiter()
	// Iniciar limpieza periódica de entradas antiguas del rate limiter (SEC-05)
	rateLimiter.StartCleanup(rootCtx, 5*time.Minute)

	// Crear Services
	userService := usersvc.NewUserService(userRepository, passwordManager)
	authService := authsvc.NewAuthService(userRepository, sessionRepository, loginAttemptRepository, passwordManager, jwtManager, sessionCache)
	companyService := companysvc.NewCompanyService(companyRepository, inventoryRepository)
	inventoryService := companysvc.NewInventoryService(inventoryRepository, companyRepository)

	// Market service (Phase 4)
	marketService := marketsvc.NewMarketService(companyRepository, inventoryRepository, gamedataCache, auditRepository)

	// Production service (Phase 5)
	productionService := productionsvc.NewProductionService(
		companyRepository,
		inventoryRepository,
		companyBuildingRepository,
		productionRunRepository,
		gamedataCache,
		auditRepository,
	)

	// Sale service (Phase 6)
	saleService := salesvc.NewSaleService(
		companyRepository,
		inventoryRepository,
		companySaleBuildingRepository,
		saleRunRepository,
		gamedataCache,
		auditRepository,
	)

	// C-01: Provide real db connection for atomic transactions
	marketService.SetDB(dbConn, queries)
	productionService.SetDB(dbConn, queries)
	saleService.SetDB(dbConn, queries)

	// Gamedata service (Phase 3)
	gamedataSvc := gameDataService.NewGamedataService(
		resourceRepository,
		productionBuildingRepository,
		productionProcessRepository,
		saleBuildingRepository,
		saleResourceRepository,
		gamedataCache,
	)

	// A-04: Periodic cleanup of old login_attempts records
	go startLoginAttemptCleanupRoutine(rootCtx, loginAttemptRepository, 24*time.Hour)

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
		// CODE-04: cancel explícito al salir del bloque, no defer (que se ejecutaría al final de main)
		adminCtx, adminCancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Verificar si admin ya existe
		adminUsername := os.Getenv("ADMIN_USERNAME")
		if adminUsername == "" {
			adminUsername = "admin"
		}
		adminUser, _ := userRepository.GetByUsername(adminCtx, adminUsername)
		if adminUser == nil {
			logger.Info().Str("username", adminUsername).Msg("Creando usuario admin...")
			_, err := userService.Register(adminCtx, adminUsername, "admin@yourownboss.local", adminPassword, "UTC")
			if err != nil {
				adminCancel()
				logger.Error().Err(err).Msg("Error al crear usuario admin")
				os.Exit(1)
			}
			logger.Info().Msg("Usuario admin creado exitosamente")
		} else {
			logger.Info().Msg("Usuario admin ya existe")
		}
		adminCancel()
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

	// M-01: Limit request body size to 1 MB to prevent memory exhaustion
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
			next.ServeHTTP(w, r)
		})
	})

	// Security headers (SEC-08): X-Content-Type-Options, X-Frame-Options, etc.
	r.Use(securityHeadersMiddleware())

	// CORS middleware (SEC-01): valida origen contra lista blanca
	r.Use(corsMiddleware())

	// Timeout middleware (30 segundos)
	r.Use(middleware.Timeout(30 * time.Second))

	// Rutas de salud
	r.Get("/api/v1/status", makeHealthHandler(dbConn))
	r.Get("/health", makeHealthHandler(dbConn))

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
	authhttphandlers.RegisterAuthRoutes(r, userService, authService, validate, rateLimiter)
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
		companyhttphandlers.RegisterCompanyRoutes(r, companyService, inventoryService, initialCompanyMoneyInt, sessionCache)
		markethttphandlers.RegisterMarketRoutes(r, marketService, rateLimiter)
		productionhttphandlers.RegisterProductionRoutes(r, productionService, rateLimiter)
		salehttphandlers.RegisterSaleRoutes(r, saleService)

		// Admin-only endpoints (require auth + admin role)
		r.Group(func(r chi.Router) {
			r.Use(authhttphandlers.RequireAuth())
			r.Use(authhttphandlers.RequireAdmin(userRepository))
			gameDataHttp.RegisterAdminGamedataRoutes(r, gamedataSvc)
		})
	})
	logger.Debug().Msg("Rutas de company, market, production y sale registradas")

	// Registrar rutas públicas de gamedata (sin autenticación)
	gameDataHttp.RegisterGamedataRoutes(r, gamedataSvc)
	logger.Debug().Msg("Rutas públicas de gamedata registradas")

	logger.Info().Msg("Rutas registradas exitosamente")

	// ============================================================================
	// INICIAR SERVIDOR HTTP
	// ============================================================================

	addr := fmt.Sprintf(":%s", port)
	logger.Info().
		Str("addr", addr).
		Str("version", appVersion).
		Msg("Servidor escuchando")

	// B-01: Graceful shutdown
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Error del servidor")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Señal de apagado recibida, cerrando servidor...")
	rootCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Error durante el apagado del servidor")
	}
	logger.Info().Msg("Servidor apagado correctamente")
}

// generateRandomSecret generates a cryptographically random secret encoded as base64.
// The length parameter controls the number of random bytes (not the output length).
// SEC-10: no truncation — the full base64 encoding preserves all entropy.
func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// makeHealthHandler returns a handler that includes db connectivity in the status response
func makeHealthHandler(dbConn interface{ Ping() error }) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appVersion := os.Getenv("APP_VERSION")
		if appVersion == "" {
			appVersion = "0.1.0"
		}

		dbConnected := dbConn.Ping() == nil

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"status":       "ok",
				"version":      appVersion,
				"timestamp":    time.Now().UTC(),
				"db_connected": dbConnected,
			},
		})
	}
}

// securityHeadersMiddleware añade cabeceras de seguridad HTTP estándar (SEC-08).
func securityHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'none'")
			next.ServeHTTP(w, r)
		})
	}
}

// corsMiddleware valida el origen contra la lista blanca antes de establecer
// Access-Control-Allow-Origin (SEC-01). Un origen no listado no recibe la cabecera
// y el navegador bloqueará la petición cross-origin.
func corsMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
			if allowedOrigins == "" {
				allowedOrigins = "http://localhost:3000"
			}

			// Construir conjunto de orígenes permitidos
			allowed := make(map[string]bool)
			for _, o := range strings.Split(allowedOrigins, ",") {
				if trimmed := strings.TrimSpace(o); trimmed != "" {
					allowed[trimmed] = true
				}
			}

			origin := r.Header.Get("Origin")
			if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}
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

// startLoginAttemptCleanupRoutine runs a periodic job to delete login_attempts older than 7 days (A-04).
func startLoginAttemptCleanupRoutine(ctx context.Context, repo authrepo.LoginAttemptRepository, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger := loggerutil.GetLogger()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := repo.DeleteOldAttempts(ctx); err != nil {
				logger.Error().Err(err).Msg("Error durante limpieza de login_attempts")
			} else {
				logger.Info().Msg("Limpieza de login_attempts completada")
			}
		}
	}
}
