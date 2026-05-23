# Fase 1.2 - Modelos y Tipos Base

## Objetivo
Crear estructuras de datos (DBO, Model, DTO) para usuarios y sesiones, así como tipos de error personalizados.

## Archivos Creados

### 1. `internal/pkg/errors.go` ✅
Sistema de errores centralizado con códigos HTTP correspondientes.

**Contenido:**
- **ErrorCode**: tipo para códigos de error estándar
- **Códigos 400**: INVALID_INPUT, INVALID_QUANTITY_MULTIPLE, INVALID_TIMEZONE, TIMEZONE_RECENTLY_CHANGED, PROCESS_NOT_AVAILABLE_NOW
- **Códigos 401**: UNAUTHORIZED, INVALID_CREDENTIALS, SESSION_EXPIRED
- **Códigos 403**: INSUFFICIENT_PERMISSIONS, COMPANY_NOT_FOUND
- **Códigos 404**: USER_NOT_FOUND, RESOURCE_NOT_FOUND, BUILDING_NOT_FOUND, PROCESS_NOT_FOUND
- **Códigos 409**: INSUFFICIENT_FUNDS, INSUFFICIENT_INVENTORY, BUILDING_NOT_IDLE, COMPANY_ALREADY_EXISTS, USERNAME_ALREADY_EXISTS, EMAIL_ALREADY_EXISTS
- **AppError**: estructura para errores de aplicación con código, mensaje y causa
- **GetHTTPStatusCode()**: método para obtener código HTTP desde ErrorCode

**Uso:**
```go
import "github.com/joanob/yourownboss/internal/pkg/errors"

// Crear error
err := errors.NewAppError(errors.InvalidCredentials, "Username o contraseña incorrectos")

// Obtener código HTTP
statusCode := err.GetHTTPStatusCode() // 401

// Respuesta JSON
response := errors.NewErrorResponse(err)
```

### 2. `internal/auth/models.go` ✅
Modelos de JWT y sesiones.

**Contenido:**
- **SessionTokenClaims**: JWT de sesión corta (1 minuto)
  - `user_id`, `session_id`, `company_id` (nullable)
  - Se valida localmente sin acceso a BD
  
- **RefreshTokenClaims**: JWT de refresco (300 días)
  - `user_id`, `session_id`, `verification_string`
  - Se valida en cache/BD
  
- **SessionData**: Estructura para almacenar sesiones en cache
  - Contiene información para validación
  - Métodos: `IsValid()`, `IsRevoked()`, `IsExpired()`

### 3. `internal/users/models/user.go` ✅
Modelos completos de usuario (DBO, Model, DTO).

**DBO (Database Object):**
```go
type UserDBO struct {
	ID                           string
	Username                     string
	Email                        string
	PasswordHash                 string
	Role                         Role
	Timezone                     *string
	LastTimezoneModificationAt   *time.Time
	CreatedAt                    time.Time
	IsDeleted                    int
	DeletedAt                    *time.Time
}
```

**Model (Lógica de negocio):**
```go
type User struct {
	// ... mismos campos que DBO ...
}

// Métodos:
- FromDBO(dbo *UserDBO)         // Convertir desde DBO
- ToDBO() *UserDBO               // Convertir a DBO
- CanChangeTimezone() bool       // Valida regla de 30 días
- TimezoneChangeBlockedUntil() *time.Time  // Fecha hasta que está bloqueado
```

**DTO (API Response):**
```go
type UserDTO struct {
	ID        string
	Username  string
	Email     string
	Role      string
	Timezone  *string
	CreatedAt string  // ISO 8601
}
```

**Request DTOs:**
- **RegisterRequest**: `{username, email, password, timezone}` + validación
- **LoginRequest**: `{username, password}` + validación
- **UpdateUserRequest**: `{timezone?}` + validación

**Response DTOs:**
- **AuthResponse**: Para login/register con session_expires_at, refresh_expires_at
- **SessionResponse**: Para respuestas de sesión

**Validaciones:**
- Username: 3-50 caracteres, alfanumérico
- Email: formato válido
- Password: 8-128 caracteres
- Timezone: formato válido IANA

### 4. `internal/users/models/session.go` ✅
Modelos completos de sesión de usuario (DBO, Model, DTO).

**DBO (Database Object):**
```go
type UserSessionDBO struct {
	ID                  string
	UserID              string
	SessionID           string
	VerificationString  string
	TokenHash           string
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	CreatedAt           time.Time
	IsDeleted           int
	DeletedAt           *time.Time
}
```

**Model:**
```go
type UserSession struct {
	// ... mismos campos que DBO ...
}

// Métodos:
- FromDBO(dbo *UserSessionDBO)
- ToDBO() *UserSessionDBO
- IsValid() bool      // No revocada + No expirada + No eliminada
- IsRevoked() bool
- IsExpired() bool
```

**DTO:**
```go
type UserSessionDTO struct {
	SessionID string
	UserID    string
	CreatedAt string  // ISO 8601
	ExpiresAt string  // ISO 8601
	IsActive  bool
}
```

## Arquitectura DBO / Model / DTO

```
Database (BD)
    ↓
  [DBO] (UserDBO, UserSessionDBO)
    ↓
 FromDBO() / ToDBO()
    ↓
  [Model] (User, UserSession)
    ↓
 Reglas de negocio
    ↓
 ToDTO()
    ↓
  [DTO] (UserDTO, AuthResponse)
    ↓
JSON Response → Cliente
```

## Validaciones Implementadas

### Usuario
- **Username**: 3-50 caracteres, solo letras y números
- **Email**: Formato válido
- **Password**: 8-128 caracteres
- **Timezone**: Formato IANA válido (ej: Europe/Madrid, America/New_York)
- **Cambio de Timezone**: Solo cada 30 días

### Sesión
- Validación de expiración
- Validación de revocación
- Validación de soft-delete

## Patrones de Uso

### Convertir DBO a Model
```go
userDBO := repo.GetByID(ctx, id)
user := &models.User{}
user.FromDBO(userDBO)

// Ahora puedo usar métodos de negocio
if user.CanChangeTimezone() {
    // permitir cambio
}
```

### Responder con DTO
```go
dto := user.ToDTO()
json.Marshal(dto)  // Enviar al cliente
```

### Manejar errores
```go
import "github.com/joanob/yourownboss/internal/pkg/errors"

if err != nil {
    appErr := errors.NewAppError(
        errors.InvalidCredentials,
        "Username o contraseña incorrectos",
    )
    statusCode := appErr.GetHTTPStatusCode()  // 401
    response := errors.NewErrorResponse(appErr)
    // Enviar response
}
```

## Próximos Pasos (Fase 1.3)

Una vez completados los modelos:

1. ✅ Modelos y tipos base creados
2. → **Fase 1.3: Configuración de JWT y Seguridad**
   - `internal/auth/jwt.go` - Generación y validación de JWT
   - `internal/auth/password.go` - Hash y verificación de contraseñas
