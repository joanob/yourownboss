package main

import (
	"context"
	"crypto/rand"
	"database/sql"
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
	"github.com/rs/zerolog/log"

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

// appConfig contiene todos los valores de configuración validados desde variables de entorno.
type appConfig struct {
	AppVersion             string
	LogLevel               string
	DatabaseURL            string
	JWTSecret              string
	Port                   string
	InitialCompanyMoney    int64
	AdminUsername          string
	AdminPassword          string
	CORSAllowedOrigins     string
	SessionCleanupInterval time.Duration
}

// appCaches agrupa todas las cachés en memoria utilizadas por la aplicación.
type appCaches struct {
	gamedata    *cache.GamedataCache
	session     *cache.SessionCache
	rateLimiter *cache.RateLimiter
}

// routerDeps agrupa todos los servicios y componentes necesarios para registrar las rutas HTTP.
type routerDeps struct {
	validate          *validator.Validate
	jwtManager        *authcrypto.JWTManager
	userService       usersvc.UserService
	authService       authsvc.AuthService
	companyService    companysvc.CompanyService
	inventoryService  companysvc.InventoryService
	marketService     *marketsvc.MarketService
	productionService *productionsvc.ProductionService
	saleService       *salesvc.SaleService
	gamedataSvc       *gameDataService.GamedataService
	userRepository    userrepo.UserRepository
}

func main() {
	loadEnv()

	logger := initLogger()

	cfg := loadConfig(logger)

	dbConn := initDatabase(cfg, logger)
	defer dbConn.Close()

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	caches := initCaches(rootCtx, cfg, logger)

	deps := buildDependencies(dbConn, caches, cfg, rootCtx, logger)

	router := buildRouter(deps, caches, cfg, dbConn)

	runServer(rootCtx, rootCancel, cfg, router, logger)
}

// loadEnv carga las variables de entorno desde el fichero .env si existe.
func loadEnv() {
	godotenv.Load()
}

// initLogger inicializa el logger global y lo devuelve.
func initLogger() zerolog.Logger {
	loggerutil.InitLogger()
	return loggerutil.GetLogger()
}

// loadConfig lee, valida y devuelve la configuración de la aplicación desde las variables de entorno.
// Termina el proceso si alguna variable obligatoria falta o tiene un valor inválido.
func loadConfig(logger zerolog.Logger) *appConfig {
	cfg := &appConfig{}

	cfg.AppVersion = getEnvOrDefault("APP_VERSION", "0.1.0")
	cfg.LogLevel = getEnvOrDefault("LOG_LEVEL", "info")
	cfg.DatabaseURL = getEnvOrDefault("DATABASE_URL", "./data/game.db")
	cfg.Port = getEnvOrDefault("PORT", "8080")
	cfg.CORSAllowedOrigins = getEnvOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	cfg.AdminUsername = getEnvOrDefault("ADMIN_USERNAME", "admin")

	cfg.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	if cfg.AdminPassword == "" {
		logger.Warn().Msg("ADMIN_PASSWORD no configurado. Admin no será creado automáticamente")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		logger.Warn().Msg("JWT_SECRET no configurado. Generando secreto aleatorio para desarrollo")
		cfg.JWTSecret = generateRandomSecret(32)
	}
	if len(cfg.JWTSecret) < 32 {
		logger.Fatal().Msg("JWT_SECRET debe tener al menos 32 caracteres. Corrige la configuración y reinicia")
	}

	moneyStr := getEnvOrDefault("INITIAL_COMPANY_MONEY", "1000")
	money, err := strconv.ParseInt(moneyStr, 10, 64)
	if err != nil || money <= 0 {
		logger.Warn().Str("value", moneyStr).Msg("INITIAL_COMPANY_MONEY inválido. Usando 1000 por defecto")
		money = 1000
	}
	cfg.InitialCompanyMoney = money

	intervalStr := getEnvOrDefault("SESSION_CACHE_CLEANUP_INTERVAL", "21600")
	intervalSecs, err := strconv.ParseInt(intervalStr, 10, 64)
	if err != nil || intervalSecs <= 0 {
		intervalSecs = 21600
	}
	cfg.SessionCleanupInterval = time.Duration(intervalSecs) * time.Second

	logger.Info().
		Str("port", cfg.Port).
		Str("log_level", cfg.LogLevel).
		Str("version", cfg.AppVersion).
		Msg("Configuración cargada correctamente")

	return cfg
}

// initDatabase abre y valida la conexión a la base de datos.
// Termina el proceso si la inicialización falla.
func initDatabase(cfg *appConfig, logger zerolog.Logger) *sql.DB {
	logger.Info().Str("path", cfg.DatabaseURL).Msg("Inicializando base de datos...")
	dbConn, err := db.InitDatabase(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error al inicializar la base de datos")
	}
	logger.Info().Msg("Base de datos inicializada correctamente")
	return dbConn
}

// initCaches crea todas las cachés en memoria e inicia sus rutinas de limpieza en segundo plano.
func initCaches(ctx context.Context, cfg *appConfig, logger zerolog.Logger) *appCaches {
	gamedataCache := cache.NewGamedataCache()
	sessionCache := cache.NewSessionCache()
	rateLimiter := cache.NewRateLimiter()

	go startSessionCleanupRoutine(sessionCache, cfg.SessionCleanupInterval)
	rateLimiter.StartCleanup(ctx, 5*time.Minute)

	logger.Info().Msg("Cachés inicializadas y rutinas de limpieza iniciadas")

	return &appCaches{
		gamedata:    gamedataCache,
		session:     sessionCache,
		rateLimiter: rateLimiter,
	}
}

// buildDependencies construye todos los repositorios y servicios, carga los datos maestros
// en caché y crea el usuario admin si no existe.
func buildDependencies(dbConn *sql.DB, caches *appCaches, cfg *appConfig, ctx context.Context, logger zerolog.Logger) *routerDeps {
	logger.Info().Msg("Inicializando dependencias...")

	validate := validator.New()
	queries := dbqueries.New(dbConn)

	jwtManager, err := authcrypto.NewJWTManager()
	if err != nil {
		logger.Fatal().Err(err).Msg("Error al crear JWT Manager")
	}
	passwordManager := authcrypto.NewPasswordManager()

	// Repositorios
	userRepository := userrepo.NewUserRepository(queries)
	sessionRepository := authrepo.NewUserSessionRepository(queries)
	loginAttemptRepository := authrepo.NewLoginAttemptRepository(queries)
	companyRepository := companyrepo.NewCompanyRepository(queries)
	inventoryRepository := companyrepo.NewInventoryRepository(queries)

	resourceRepository := resourcerepo.NewResourceRepository(queries)
	productionBuildingRepository := productionrepo.NewProductionBuildingRepository(queries)
	productionProcessRepository := productionrepo.NewProductionProcessRepository(queries)
	saleBuildingRepository := salerepo.NewSaleBuildingRepository(queries)
	saleResourceRepository := salerepo.NewSaleResourceRepository(queries)

	companyBuildingRepository := productionrepo.NewCompanyBuildingRepository(queries)
	productionRunRepository := productionrepo.NewProductionRunRepository(queries)

	companySaleBuildingRepository := salerepo.NewCompanySaleBuildingRepository(queries)
	saleRunRepository := salerepo.NewSaleRunRepository(queries)

	auditRepository := auditrepo.NewAuditRepository(queries)

	// Servicios
	userService := usersvc.NewUserService(userRepository, passwordManager)
	authService := authsvc.NewAuthService(userRepository, sessionRepository, loginAttemptRepository, passwordManager, jwtManager, caches.session)
	companyService := companysvc.NewCompanyService(companyRepository, inventoryRepository)
	inventoryService := companysvc.NewInventoryService(inventoryRepository, companyRepository)

	marketService := marketsvc.NewMarketService(companyRepository, inventoryRepository, caches.gamedata, auditRepository)

	productionService := productionsvc.NewProductionService(
		companyRepository,
		inventoryRepository,
		companyBuildingRepository,
		productionRunRepository,
		caches.gamedata,
		auditRepository,
	)

	saleService := salesvc.NewSaleService(
		companyRepository,
		inventoryRepository,
		companySaleBuildingRepository,
		saleRunRepository,
		caches.gamedata,
		auditRepository,
	)

	// C-01: conexión real para transacciones atómicas
	marketService.SetDB(dbConn, queries)
	productionService.SetDB(dbConn, queries)
	saleService.SetDB(dbConn, queries)

	gamedataSvc := gameDataService.NewGamedataService(
		resourceRepository,
		productionBuildingRepository,
		productionProcessRepository,
		saleBuildingRepository,
		saleResourceRepository,
		caches.gamedata,
	)

	// A-04: limpieza periódica de registros antiguos de login_attempts
	go startLoginAttemptCleanupRoutine(ctx, loginAttemptRepository, 24*time.Hour)

	// Cargar datos maestros en caché
	cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cacheCancel()
	if err := gamedataSvc.RefreshCache(cacheCtx); err != nil {
		logger.Fatal().Err(err).Msg("Error al cargar datos maestros en caché")
	}
	logger.Info().Msg("Datos maestros cargados en caché")

	// Crear usuario admin si no existe
	if cfg.AdminPassword != "" {
		adminCtx, adminCancel := context.WithTimeout(context.Background(), 5*time.Second)
		adminUser, _ := userRepository.GetByUsername(adminCtx, cfg.AdminUsername)
		if adminUser == nil {
			logger.Info().Str("username", cfg.AdminUsername).Msg("Creando usuario admin...")
			_, err := userService.Register(adminCtx, cfg.AdminUsername, "admin@yourownboss.local", cfg.AdminPassword, "UTC")
			if err != nil {
				adminCancel()
				logger.Fatal().Err(err).Msg("Error al crear usuario admin")
			}
			logger.Info().Msg("Usuario admin creado exitosamente")
		} else {
			logger.Info().Msg("Usuario admin ya existe")
		}
		adminCancel()
	}

	logger.Info().Msg("Dependencias inicializadas correctamente")

	return &routerDeps{
		validate:          validate,
		jwtManager:        jwtManager,
		userService:       userService,
		authService:       authService,
		companyService:    companyService,
		inventoryService:  inventoryService,
		marketService:     marketService,
		productionService: productionService,
		saleService:       saleService,
		gamedataSvc:       gamedataSvc,
		userRepository:    userRepository,
	}
}

// buildRouter crea el router HTTP y registra todos los middlewares y rutas.
func buildRouter(deps *routerDeps, caches *appCaches, cfg *appConfig, dbConn *sql.DB) http.Handler {
	logger := loggerutil.GetLogger()

	r := chi.NewRouter()

	// Middlewares globales
	r.Use(middleware.RequestID)
	r.Use(clientIPMiddleware())
	r.Use(requestLoggingMiddleware())
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// M-01: limitar el tamaño del cuerpo de la petición a 1 MB
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			next.ServeHTTP(w, r)
		})
	})
	// SEC-08: cabeceras de seguridad HTTP
	r.Use(securityHeadersMiddleware())
	// SEC-01: CORS con lista blanca de orígenes
	r.Use(corsMiddleware())
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", makeHealthHandler(dbConn))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Your Own Boss API",
			"version": cfg.AppVersion,
			"status":  "ok",
		})
	})

	authhttphandlers.RegisterAuthRoutes(r, deps.userService, deps.authService, deps.validate, caches.rateLimiter)
	logger.Debug().Msg("Rutas de autenticación registradas")

	r.Route("/api/v1", func(r chi.Router) {
		gameDataHttp.RegisterGamedataRoutes(r, deps.gamedataSvc)
		logger.Debug().Msg("Rutas públicas de gamedata registradas")

		r.Group(func(r chi.Router) {
			r.Use(authhttphandlers.AuthMiddleware(deps.jwtManager, caches.session))

			r.Route("/users", func(r chi.Router) {
				userhttphandlers.RegisterUsersRoutes(r, deps.userService, deps.validate)
			})
			logger.Debug().Msg("Rutas de usuarios registradas")

			r.Get("/status", makeHealthHandler(dbConn))
			companyhttphandlers.RegisterCompanyRoutes(r, deps.companyService, deps.inventoryService, cfg.InitialCompanyMoney, caches.session)
			markethttphandlers.RegisterMarketRoutes(r, deps.marketService, caches.rateLimiter)
			productionhttphandlers.RegisterProductionRoutes(r, deps.productionService, caches.rateLimiter)
			salehttphandlers.RegisterSaleRoutes(r, deps.saleService)

			// Endpoints exclusivos de admin
			r.Group(func(r chi.Router) {
				r.Use(authhttphandlers.RequireAuth())
				r.Use(authhttphandlers.RequireAdmin(deps.userRepository))
				gameDataHttp.RegisterAdminGamedataRoutes(r, deps.gamedataSvc)
			})
		})
	})
	logger.Debug().Msg("Rutas de company, market, production, users y gamedata registradas")

	logger.Info().Msg("Rutas registradas exitosamente")

	return r
}

// runServer inicia el servidor HTTP y bloquea hasta recibir una señal de apagado.
// Realiza un graceful shutdown con un timeout de 30 segundos.
func runServer(ctx context.Context, cancel context.CancelFunc, cfg *appConfig, router http.Handler, logger zerolog.Logger) {
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info().
		Str("addr", addr).
		Str("version", cfg.AppVersion).
		Msg("Servidor escuchando")

	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Error del servidor")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Señal de apagado recibida, cerrando servidor...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Error durante el apagado del servidor")
	}
	logger.Info().Msg("Servidor apagado correctamente")
}

// getEnvOrDefault devuelve el valor de la variable de entorno key, o defaultVal si está vacía.
func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
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

func clientIPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			ctx := context.WithValue(r.Context(), "client_ip", ip)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	// 1. X-Forwarded-For (si está detrás de un proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[len(ips)-1])
		}
	}

	// 2. X-Real-IP (usado por algunos proxies)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 3. True-Client-IP (Cloudflare)
	if tci := r.Header.Get("True-Client-IP"); tci != "" {
		return strings.TrimSpace(tci)
	}

	// 4. RemoteAddr como fallback
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

func requestLoggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Context().Value("client_ip")
			reqID := r.Context().Value(middleware.RequestIDKey)
			log.Info().
				Str("ip", ip.(string)).
				Str("method", r.Method).
				Str("path", r.RequestURI).
				Str("request_id", reqID.(string)).
				Msg("Incoming request")
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
