package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

// InitDatabase inicializa la base de datos SQLite
// Crea el archivo de BD si no existe y ejecuta el schema
func InitDatabase(dbPath string) (*sql.DB, error) {
	logger := log.Logger

	// Crear directorio si no existe
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error al crear directorio de BD: %w", err)
	}

	// Abrir conexión a SQLite
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir BD: %w", err)
	}

	// Verificar que la conexión funciona
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al conectar a BD: %w", err)
	}

	// Habilitar foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("error al habilitar foreign keys: %w", err)
	}

	// Leer y ejecutar schema.sql
	schemaPath := filepath.Join(filepath.Dir(dbPath), "migrations", "schema.sql")
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Si no encuentra en migrations, intentar en ruta relativa común
		schemaPath = "internal/db/migrations/schema.sql"
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("error al leer schema.sql en %s: %w", schemaPath, err)
	}

	// Ejecutar cada sentencia SQL del schema
	if _, err := db.Exec(string(schema)); err != nil {
		return nil, fmt.Errorf("error al ejecutar schema: %w", err)
	}

	logger.Info().
		Str("path", dbPath).
		Msg("Base de datos inicializada exitosamente")

	return db, nil
}

// GetDatabase abre una conexión existente a la BD
// No crea la BD ni ejecuta migraciones, solo abre la conexión
func GetDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir BD: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al conectar a BD: %w", err)
	}

	// Habilitar foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("error al habilitar foreign keys: %w", err)
	}

	return db, nil
}

// ResetDatabase elimina la BD y la reinicializa (útil para desarrollo)
func ResetDatabase(dbPath string) (*sql.DB, error) {
	logger := log.Logger

	// Eliminar archivos de BD
	_ = os.Remove(dbPath)
	_ = os.Remove(dbPath + "-shm")
	_ = os.Remove(dbPath + "-wal")

	logger.Info().
		Str("path", dbPath).
		Msg("Base de datos anterior eliminada")

	// Reinicializar
	return InitDatabase(dbPath)
}
