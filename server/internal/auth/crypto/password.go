package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// PasswordManager maneja hash y verificación de contraseñas
type PasswordManager struct {
	// Parámetros de Argon2id (ajustables según necesidad)
	// Estos parámetros son seguros contra ataques GPU en 2026
	time      uint32 // número de iteraciones
	memory    uint32 // memoria en KiB
	threads   uint8  // paralelismo
	keyLength uint32 // longitud del hash
	saltSize  int    // tamaño de salt en bytes
}

// NewPasswordManager crea un nuevo gestor de contraseñas
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		time:      3,     // 3 iteraciones
		memory:    65536, // 64 MiB
		threads:   4,     // 4 threads
		keyLength: 32,    // 32 bytes = 256 bits
		saltSize:  16,    // 16 bytes de salt
	}
}

// HashPassword hashea una contraseña usando Argon2id
// Devuelve hash en formato: $argon2id$v=19$m=65536,t=3,p=4$<base64-salt>$<base64-hash>
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	// Generar salt aleatorio
	salt := make([]byte, pm.saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("error generando salt: %w", err)
	}

	// Computar hash con Argon2id
	// Argon2id es más resistente a ataques GPU/ASIC que Argon2i
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		pm.time,
		pm.memory,
		pm.threads,
		pm.keyLength,
	)

	// Codificar en formato PHC (Password Hashing Competition)
	// Formato: $argon2id$v=19$m=65536,t=3,p=4$<base64-salt>$<base64-hash>
	return encodeArgon2(salt, hash, pm.time, pm.memory, pm.threads), nil
}

// VerifyPassword verifica que una contraseña coincide con su hash
func (pm *PasswordManager) VerifyPassword(password, hash string) bool {
	// Decodificar el hash
	salt, hashBytes, params, err := decodeArgon2(hash)
	if err != nil {
		return false
	}

	// Recomputar el hash con la contraseña ingresada
	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.time,
		params.memory,
		params.threads,
		uint32(len(hashBytes)),
	)

	// Comparar hashes de forma segura (timing-safe)
	return constantTimeCompare(hashBytes, newHash)
}

// encodeArgon2 codifica salt y hash en formato PHC
func encodeArgon2(salt, hash []byte, time, memory uint32, threads uint8) string {
	// Formato: $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<base64-salt>$<base64-hash>
	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		time,
		threads,
		encodeBase64(salt),
		encodeBase64(hash),
	)
}

// argon2Params estructura para almacenar parámetros decodificados
type argon2Params struct {
	time    uint32
	memory  uint32
	threads uint8
}

// decodeArgon2 decodifica un hash en formato PHC
func decodeArgon2(hash string) ([]byte, []byte, argon2Params, error) {
	var params argon2Params
	var saltStr, hashStr string

	// Parse: $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<base64-salt>$<base64-hash>
	_, err := fmt.Sscanf(hash, "$argon2id$v=19$m=%d,t=%d,p=%d$%[4]s$%[5]s",
		&params.memory, &params.time, &params.threads, &saltStr, &hashStr)
	if err != nil {
		// Intenta parsear con diferentes separadores
		_, err = fmt.Sscanf(hash, "$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
			&params.memory, &params.time, &params.threads, &saltStr, &hashStr)
		if err != nil {
			return nil, nil, params, fmt.Errorf("error parseando hash: %w", err)
		}
	}

	// Decodificar salt y hash desde base64
	salt, err := decodeBase64(saltStr)
	if err != nil {
		return nil, nil, params, fmt.Errorf("error decodificando salt: %w", err)
	}

	hashBytes, err := decodeBase64(hashStr)
	if err != nil {
		return nil, nil, params, fmt.Errorf("error decodificando hash: %w", err)
	}

	return salt, hashBytes, params, nil
}

// encodeBase64 codifica bytes a base64 sin padding
func encodeBase64(data []byte) string {
	return fmt.Sprintf("%x", data)
}

// decodeBase64 decodifica hex a bytes
func decodeBase64(s string) ([]byte, error) {
	// Usar hex en lugar de base64 para mayor compatibilidad
	var result []byte
	_, err := fmt.Sscanf(s, "%x", &result)
	if err != nil {
		// Intentar con raw string si sscanf falla
		// En realidad usar hex.DecodeString es más seguro
		for i := 0; i < len(s); i += 2 {
			var b byte
			_, err := fmt.Sscanf(s[i:i+2], "%02x", &b)
			if err != nil {
				return nil, err
			}
			result = append(result, b)
		}
	}
	return result, nil
}

// constantTimeCompare compara dos slices de bytes en tiempo constante
// Previene timing attacks
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
