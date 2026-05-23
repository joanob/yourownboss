# Fase 1.3 - Configuración de JWT y Seguridad

## Objetivo
Implementar generación y validación de JWT, así como hash seguro de contraseñas con Argon2id.

## Archivos Creados

### 1. `internal/auth/jwt.go` ✅
Sistema completo de generación y validación de JWT.

**Estructura:**
```go
type JWTManager struct {
    secretKey     []byte
    sessionExpiry int64  // segundos (default 60)
    refreshExpiry int64  // segundos (default 25920000 = 300 días)
}
```

**Métodos principales:**

#### `NewJWTManager() (*JWTManager, error)`
- Carga o genera `JWT_SECRET` desde ENV
- Si no existe, genera una clave aleatoria de 32 bytes (con advertencia en logs)
- Valida longitud mínima de 32 caracteres
- Carga duraciones desde ENV con defaults

**Variables de entorno:**
- `JWT_SECRET`: Clave de firma (default: generada aleatoriamente, recomendado configurar en .env)
- `JWT_SESSION_EXPIRY`: Duración en segundos (default: 60 = 1 minuto)
- `JWT_REFRESH_EXPIRY`: Duración en segundos (default: 25920000 = 300 días)

#### `GenerateSessionToken(userID, sessionID string, companyID *string) (string, time.Time, error)`
- Genera JWT de sesión corta (1 minuto)
- Contiene: `user_id`, `session_id`, `company_id` (nullable)
- Devuelve: token string + expiry time
- Se valida **localmente sin acceso a BD** (solo firma)

**Claims:**
```go
type SessionTokenClaims struct {
    UserID    string
    SessionID string
    CompanyID *string  // nullable
    ExpiresAt int64    // Unix timestamp
    IssuedAt  int64    // Unix timestamp
}
```

#### `ValidateSessionToken(tokenString string) (*SessionTokenClaims, error)`
- Valida firma HS256
- Verifica expiración
- Devuelve claims si válido
- NO accede a BD

#### `GenerateRefreshToken(userID, sessionID string) (string, string, time.Time, error)`
- Genera JWT de refresco largo (300 días)
- Genera `verification_string` aleatorio de 32 bytes (base64)
- Devuelve: token string + verification_string + expiry time
- El `verification_string` se almacena como hash en BD

**Claims:**
```go
type RefreshTokenClaims struct {
    UserID             string
    SessionID          string
    VerificationString string  // hash para validar en BD/cache
    ExpiresAt          int64
    IssuedAt           int64
}
```

#### `ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error)`
- Valida firma HS256
- Verifica expiración
- **NO valida contra BD** (ese trabajo es del servicio/middleware)
- Devuelve claims si JWT válido

#### `GenerateSessionID() string`
- Genera UUID único para la sesión

**Flujo de validación:**
```
Request con Session Token
  ↓
ValidateSessionToken() ← No accede a BD, solo verifica firma
  ↓
Si expirado, usar Refresh Token
  ↓
ValidateRefreshToken() ← Verifica firma del JWT
  ↓
Service valida contra BD/cache (verification_string)
  ↓
Si válido, generar nuevo Session Token
```

### 2. `internal/auth/password.go` ✅
Hash y verificación segura de contraseñas.

**Estructura:**
```go
type PasswordManager struct {
    time      uint32  // iteraciones (default 3)
    memory    uint32  // memoria en KiB (default 65536 = 64MiB)
    threads   uint8   // paralelismo (default 4)
    keyLength uint32  // longitud del hash (default 32)
    saltSize  int     // tamaño del salt (default 16 bytes)
}
```

**Parámetros de Argon2id:**
- **Algoritmo**: Argon2id (más resistente a ataques GPU/ASIC que Argon2i)
- **Tiempo**: 3 iteraciones
- **Memoria**: 64 MiB (65536 KiB)
- **Threads**: 4 paralelismo
- **Salt**: 16 bytes aleatorio
- **Hash**: 32 bytes (256 bits)

**Métodos:**

#### `NewPasswordManager() *PasswordManager`
- Crea un nuevo gestor con parámetros seguros para 2026
- Sin ENV requeridas (parámetros fijos)

#### `HashPassword(password string) (string, error)`
- Genera salt aleatorio de 16 bytes
- Computa Argon2id hash
- Devuelve string en formato PHC: `$argon2id$v=19$m=65536,t=3,p=4$<hex-salt>$<hex-hash>`

**Ejemplo:**
```
$argon2id$v=19$m=65536,t=3,p=4$7a8c9e2f3b4d5a6c7e8f9a2b3c4d5e6f$9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9
```

#### `VerifyPassword(password, hash string) bool`
- Decodifica hash PHC
- Recomputa Argon2id con contraseña ingresada
- Compara con timing-safe comparison (evita timing attacks)
- Devuelve true/false

**Seguridad:**
- Resistente a ataques GPU/ASIC (alto uso de memoria)
- Timing-safe comparison (previene timing attacks)
- Salt aleatorio (previene rainbow tables)
- Parámetros configurables en código (no en ENV para no exponerlos)

## Uso en Servicios

### Generar tokens en login
```go
jwtManager, err := auth.NewJWTManager()
if err != nil {
    // error
}

// Generar Session Token
sessionToken, sessionExpiry, err := jwtManager.GenerateSessionToken(
    userID,
    auth.GenerateSessionID(),
    &companyID,
)

// Generar Refresh Token
refreshToken, verificationString, refreshExpiry, err := jwtManager.GenerateRefreshToken(
    userID,
    sessionID,
)

// Guardar en BD: session_id, verification_string (hash), refreshToken (hash)
```

### Validar tokens en middleware
```go
// Validar Session Token (sin BD)
claims, err := jwtManager.ValidateSessionToken(sessionTokenFromCookie)
if err != nil {
    // Token inválido o expirado, intentar con Refresh Token
}

// Si Session Token expirado:
refreshClaims, err := jwtManager.ValidateRefreshToken(refreshTokenFromCookie)
if err != nil {
    // Refresh Token inválido, usuario debe login de nuevo
}

// Validar contra BD/cache
session, err := sessionRepo.GetByID(refreshClaims.SessionID)
if session.VerificationString != refreshClaims.VerificationString {
    // Fraud attempt
}
if session.IsRevoked() {
    // Session revocada
}

// Generar nuevo Session Token
newSessionToken, _, err := jwtManager.GenerateSessionToken(...)
```

### Hash de contraseña en registro
```go
pm := auth.NewPasswordManager()

// Hash en register
passwordHash, err := pm.HashPassword(password)
if err != nil {
    // error
}
// Guardar passwordHash en BD

// Verificación en login
if !pm.VerifyPassword(passwordIngresada, passwordHashDeBD) {
    // Contraseña incorrecta
}
```

## Configuración en .env (Recomendado)

```bash
# JWT
JWT_SECRET=tu-clave-secreta-de-32-caracteres-minimo-muy-segura
JWT_SESSION_EXPIRY=60                # 1 minuto
JWT_REFRESH_EXPIRY=25920000          # 300 días

# Otras configuraciones
PORT=8080
DATABASE_URL=yourownboss.db
GAMEDATA_FILE=./config/gamedata.json
LOG_LEVEL=info
```

## Seguridad

✅ **Algoritmo Argon2id**
- Resistente a ataques GPU/ASIC (2026+)
- Alto uso de memoria = caro para ataques fuerza bruta
- Iteraciones configurables = adaptable a futuros ataques

✅ **Timing-safe comparison**
- Previene timing attacks en verificación de contraseñas
- Compara bit a bit en tiempo constante

✅ **JWT con firma HS256**
- Validación local sin BD (Session Token)
- Imposible falsificar sin JWT_SECRET

✅ **Session ID aleatorio (UUID)**
- Generado por servidor
- No es predecible
- Almacenado en BD

✅ **Verification String aleatorio**
- 32 bytes aleatorios (256 bits)
- Base64 codificado
- Almacenado en BD para validar Refresh Token

## Próximos Pasos (Fase 1.4)

**Fase 1.4: Repository de Usuarios**
- Crear `internal/users/repository/user_repo.go`
- Métodos: CreateUser, GetByID, GetByUsername, UpdateUser, SoftDeleteUser
- Usar sqlc para generar queries tipadas
