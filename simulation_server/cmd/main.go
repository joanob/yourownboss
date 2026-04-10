package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"yourownboss/simulation/internal/auth"
	"yourownboss/simulation/internal/migrations"
	"yourownboss/simulation/internal/production"
	"yourownboss/simulation/internal/resources"
	"yourownboss/simulation/internal/simulation"

	_ "modernc.org/sqlite"
)

func main() {
	// Lod environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found or failed to load, continuing with env vars")
	}

	// Database connection

	dsn := os.Getenv("DB_PATH")
	if dsn == "" {
		// default to a local file-based SQLite DB
		dsn = "file:yob_simulation.db?cache=shared&mode=rwc"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("database connected")

	// Run SQL migrations from ./migrations
	migrationsDir := "./migrations"
	if err := migrations.Migrate(db, migrationsDir); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("migrations applied")

	// Router and middleware

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Auth middleware (all routes require Authorization: Bearer <AUTH_TOKEN>)
	r.Use(auth.AuthMiddleware)

	// Simple health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Register resources routes
	resources.RegisterRoutes(r, db)

	// Register production routes
	production.RegisterRoutes(r, db)

	// Register simulation routes
	simulation.RegisterRoutes(r, db)

	// Server config
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	// Start server
	go func() {
		log.Printf("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
