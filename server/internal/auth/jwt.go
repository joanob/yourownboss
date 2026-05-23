package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// JWTManager maneja la generación y validación de JWT
type JWTManager struct {
	secretKey     []byte
	sessionExpiry int64 // segundos
	refreshExpiry int64 // segundos
}

// NewJWTManager crea un nuevo gestor de JWT
func NewJWTManager() (*JWTManager, error) {
	logger := log.Logger

	// Cargar o generar JWT_SECRET desde ENV
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		// Generar clave aleatoria de 32 bytes si no existe
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			return nil, fmt.Errorf("error generando JWT_SECRET aleatorio: %w", err)
		}
		secretKey = base64.StdEncoding.EncodeToString(randomBytes)
		logger.Warn().Msg("JWT_SECRET no configurado, generado uno aleatorio. CONFIGURA JWT_SECRET en .env para producción")
	}

	// Validar que la clave tenga longitud mínima (32 caracteres)
	if len(secretKey) < 32 {
		return nil, fmt.Errorf("JWT_SECRET debe tener al menos 32 caracteres, actual: %d", len(secretKey))
	}

	// Cargar duración de sesión desde ENV (default 60 segundos)
	sessionExpiryStr := os.Getenv("JWT_SESSION_EXPIRY")
	if sessionExpiryStr == "" {
		sessionExpiryStr = "60"
	}
	sessionExpiry, err := strconv.ParseInt(sessionExpiryStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("JWT_SESSION_EXPIRY inválido: %w", err)
	}

	// Cargar duración de refresh desde ENV (default 300 días)
	refreshExpiryStr := os.Getenv("JWT_REFRESH_EXPIRY")
	if refreshExpiryStr == "" {
		refreshExpiryStr = "25920000" // 300 días en segundos
	}
	refreshExpiry, err := strconv.ParseInt(refreshExpiryStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("JWT_REFRESH_EXPIRY inválido: %w", err)
	}

	logger.Info().
		Int64("session_expiry_seconds", sessionExpiry).
		Int64("refresh_expiry_seconds", refreshExpiry).
		Msg("JWT Manager inicializado")

	return &JWTManager{
		secretKey:     []byte(secretKey),
		sessionExpiry: sessionExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// GenerateSessionToken genera un JWT de sesión corta
// Contiene user_id, session_id, company_id (nullable)
// Se valida localmente sin acceso a BD
func (m *JWTManager) GenerateSessionToken(userID, sessionID string, companyID *string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(m.sessionExpiry) * time.Second)

	claims := SessionTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		CompanyID: companyID,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error firmando session token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateSessionToken valida y parsea un JWT de sesión
// NO accede a BD, solo verifica firma y expiración
func (m *JWTManager) ValidateSessionToken(tokenString string) (*SessionTokenClaims, error) {
	claims := &SessionTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el algoritmo es HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parseando session token: %w", err)
	}

	// Verificar que el token es válido
	if !token.Valid {
		return nil, fmt.Errorf("session token inválido")
	}

	// Verificar que no ha expirado
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("session token expirado")
	}

	return claims, nil
}

// GenerateRefreshToken genera un JWT de refresco largo (300 días)
// Contiene user_id, session_id, verification_string
// Se almacena como hash en BD y se valida buscándolo
func (m *JWTManager) GenerateRefreshToken(userID, sessionID string) (string, string, time.Time, error) {
	// Generar verification_string aleatorio
	verificationBytes := make([]byte, 32)
	if _, err := rand.Read(verificationBytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("error generando verification_string: %w", err)
	}
	verificationString := base64.StdEncoding.EncodeToString(verificationBytes)

	now := time.Now()
	expiresAt := now.Add(time.Duration(m.refreshExpiry) * time.Second)

	claims := RefreshTokenClaims{
		UserID:             userID,
		SessionID:          sessionID,
		VerificationString: verificationString,
		IssuedAt:           now.Unix(),
		ExpiresAt:          expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("error firmando refresh token: %w", err)
	}

	return tokenString, verificationString, expiresAt, nil
}

// ValidateRefreshToken valida y parsea un JWT de refresco
// Solo verifica firma y expiración, NO valida contra BD
// La validación contra BD se hace en el servicio/middleware
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el algoritmo es HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parseando refresh token: %w", err)
	}

	// Verificar que el token es válido
	if !token.Valid {
		return nil, fmt.Errorf("refresh token inválido")
	}

	// Verificar que no ha expirado
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("refresh token expirado")
	}

	return claims, nil
}

// GenerateSessionID genera un UUID único para la sesión
func GenerateSessionID() string {
	return uuid.New().String()
}
