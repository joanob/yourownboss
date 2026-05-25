package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

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
	// Formato: $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<hex-salt>$<hex-hash>
	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		time,
		threads,
		encodePasswordBytes(salt),
		encodePasswordBytes(hash),
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

	// Formato: $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<hex-salt>$<hex-hash>
	// Split produce: ["", "argon2id", "v=19", "m=...,t=...,p=...", "<salt>", "<hash>"]
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, params, fmt.Errorf("formato de hash inválido")
	}

	// Parsear parámetros: "m=65536,t=3,p=4"
	for _, kv := range strings.Split(parts[3], ",") {
		pair := strings.SplitN(kv, "=", 2)
		if len(pair) != 2 {
			return nil, nil, params, fmt.Errorf("parámetro de hash malformado: %s", kv)
		}
		val, err := strconv.ParseUint(pair[1], 10, 64)
		if err != nil {
			return nil, nil, params, fmt.Errorf("valor de parámetro inválido %s: %w", kv, err)
		}
		switch pair[0] {
		case "m":
			params.memory = uint32(val)
		case "t":
			params.time = uint32(val)
		case "p":
			params.threads = uint8(val)
		}
	}

	// Decodificar salt y hash desde hex
	salt, err := hex.DecodeString(parts[4])
	if err != nil {
		return nil, nil, params, fmt.Errorf("error decodificando salt: %w", err)
	}

	hashBytes, err := hex.DecodeString(parts[5])
	if err != nil {
		return nil, nil, params, fmt.Errorf("error decodificando hash: %w", err)
	}

	return salt, hashBytes, params, nil
}

// encodePasswordBytes codifica bytes a hex
func encodePasswordBytes(data []byte) string {
	return hex.EncodeToString(data)
}

// constantTimeCompare compara dos slices de bytes en tiempo constante
// Previene timing attacks
func constantTimeCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
