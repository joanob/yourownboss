package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/joanob/yourownboss/internal/db"
	"github.com/joanob/yourownboss/internal/gamedata/service"
	"github.com/joanob/yourownboss/internal/pkg/cache"
	"github.com/joanob/yourownboss/internal/pkg/logger"
)

func main() {
	// Cargar variables de entorno
	godotenv.Load()

	// Inicializar logger
	logger.InitLogger()
	logger := logger.GetLogger()

	// Obtener configuración
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Validar JWT_SECRET
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Error().Msg("JWT_SECRET no configurado. Usar variable de entorno JWT_SECRET")
		os.Exit(1)
	}
	if len(jwtSecret) < 32 {
		logger.Warn().Msg("JWT_SECRET tiene menos de 32 caracteres. Se recomienda usar al menos 32 caracteres")
	}

	logger.Info().
		Str("port", port).
		Str("env", env).
		Msg("Iniciando Your Own Boss Backend")

	// Inicializar base de datos
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "./yourownboss.db"
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

	// Cargar datos maestros (gamedata)
	gamedataFilePath := os.Getenv("GAMEDATA_FILE")
	if gamedataFilePath == "" {
		gamedataFilePath = "./config/gamedata.json"
	}

	gamedataSvc := service.NewGamedataService(gamedataFilePath, dbConn)

	// 1. Asegurar que BD tiene datos (importa desde JSON si está vacía)
	if err := gamedataSvc.Load(); err != nil {
		logger.Error().Err(err).Msg("Error al cargar datos maestros en BD")
		os.Exit(1)
	}

	// 2. Sincronizar cache desde BD
	if err := gamedataSvc.RefreshCache(gamedataCache); err != nil {
		logger.Error().Err(err).Msg("Error al sincronizar cache desde BD")
		os.Exit(1)
	}

	// Iniciar limpiador de sesiones (cada 6 horas)
	sessionCleanupInterval := os.Getenv("SESSION_CACHE_CLEANUP_INTERVAL")
	if sessionCleanupInterval == "" {
		sessionCleanupInterval = "21600" // 6 horas
	}

	cleanupIntervalSecs, _ := strconv.ParseInt(sessionCleanupInterval, 10, 64)
	go startSessionCleanupRoutine(sessionCache, time.Duration(cleanupIntervalSecs)*time.Second)

	// Crear router
	r := chi.NewRouter()

	// Middleware global
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS middleware (basic)
	r.Use(corsMiddleware())

	// Rutas de salud
	r.Get("/api/v1/status", healthHandler)
	r.Get("/health", healthHandler)

	// Placeholder para rutas futuras
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Your Own Boss API",
			"version": "0.1.0",
			"status":  "ok",
		})
	})

	// Servidor HTTP
	addr := fmt.Sprintf(":%s", port)
	logger.Info().Str("addr", addr).Msg("Servidor escuchando")

	if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Error al iniciar servidor")
	}
}

// healthHandler devuelve el estado del servidor
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"status":       "ok",
			"version":      "0.1.0",
			"timestamp":    time.Now().UTC(),
			"db_connected": true, // TODO: implementar verificación real de BD
		},
	})
}

// corsMiddleware agrega headers CORS básicos
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

	logger := log.Logger

	for range ticker.C {
		deleted := sc.CleanupExpired()
		logger.Info().
			Int("deleted_sessions", deleted).
			Int("remaining_sessions", sc.Count()).
			Msg("Limpieza de sesiones completada")
	}
}
