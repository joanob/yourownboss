package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger configura zerolog con nivel configurable desde ENV
// Parámetros de rotación: máximo 100 archivos de 10MB cada uno (1GB total)
func InitLogger() {
	// Crear directorio de logs si no existe
	os.MkdirAll("./logs", 0755)

	// Configurar rotación de archivos
	logFile := &lumberjack.Logger{
		Filename:   "./logs/server.log",
		MaxSize:    10, // MB
		MaxBackups: 100,
		MaxAge:     30, // días (aproximado, pero no es crítico con MaxBackups)
		Compress:   true,
	}

	// Obtener nivel de log desde ENV (default: info)
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info"
	}

	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Configurar formato de timestamp a RFC3339 (UTC ISO8601)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Multi-writer: stdout + archivo
	multiWriter := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout},
		logFile,
	)

	log.Logger = zerolog.New(multiWriter).
		Level(level).
		With().
		Timestamp().
		Logger()

	log.Info().Str("log_level", levelStr).Msg("Logger inicializado")
}

// GetLogger devuelve la instancia global del logger
func GetLogger() zerolog.Logger {
	return log.Logger
}
