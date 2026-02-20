# Arquitectura del Backend

El backend sigue una arquitectura en capas con separación de responsabilidades y principios SOLID.

## Estructura de Capas

```
┌─────────────────────────────────────┐
│         HTTP Handler Layer          │  ← Controllers (HTTP)
│   (internal/http/*_handler.go)      │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│         Service Layer               │  ← Business Logic
│   (internal/service/*_service.go)   │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Repository Layer              │  ← Data Access
│ (internal/repository/*_repository.go)│
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│           Database                  │  ← SQLite
│        (internal/db/)                │
└─────────────────────────────────────┘
```

## Capas

### 1. Handler Layer (Controllers)

**Ubicación:** `internal/http/`

**Responsabilidad:**

- Manejo de requests HTTP
- Validación de entrada
- Serialización/Deserialización JSON
- Gestión de cookies
- Mapeo de errores del servicio a códigos HTTP

**Ejemplo:**

```go
type AuthHandler struct {
    authService service.AuthService
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    // 2. Validate input
    // 3. Call service
    // 4. Handle response/errors
    // 5. Set cookies
}
```

### 2. Service Layer

**Ubicación:** `internal/service/`

**Responsabilidad:**

- Lógica de negocio
- Validación de reglas de negocio
- Orquestación de operaciones
- Manejo de transacciones
- Generación de tokens

**Ejemplo:**

```go
type AuthService interface {
    Register(ctx context.Context, username, password string) (*AuthResult, error)
    Login(ctx context.Context, username, password string) (*AuthResult, error)
}

type authService struct {
    userRepo  repository.UserRepository
    tokenRepo repository.TokenRepository
}
```

### 3. Repository Layer

**Ubicación:** `internal/repository/`

**Responsabilidad:**

- Acceso a datos
- Queries SQL
- Mapeo de entidades
- Gestión de errores de BD

**Ejemplo:**

```go
type UserRepository interface {
    Create(ctx context.Context, username, passwordHash string) (*db.User, error)
    GetByUsername(ctx context.Context, username string) (*db.User, error)
}
```

### 4. Database Layer

**Ubicación:** `internal/db/`

**Responsabilidad:**

- Conexión a base de datos
- Esquema y migraciones
- Modelos de datos

## Inyección de Dependencias

El flujo de DI en `cmd/api/main.go`:

```go
// 1. Database
database, _ := db.Open(dbPath)

// 2. Repositories
userRepo := repository.NewUserRepository(database)
tokenRepo := repository.NewTokenRepository(database)

// 3. Services
authService := service.NewAuthService(userRepo, tokenRepo)

// 4. Handlers
authHandler := httpHandlers.NewAuthHandler(authService)
```

## Ventajas de esta Arquitectura

1. **Testabilidad**: Cada capa se puede testear independientemente con mocks
2. **Mantenibilidad**: Código organizado y fácil de entender
3. **Escalabilidad**: Fácil añadir nuevas features sin romper existentes
4. **Separación de Responsabilidades**: Cada capa tiene un propósito claro
5. **Reutilización**: Los servicios pueden ser usados desde diferentes handlers
6. **Cambio de BD**: Solo afecta la capa de repositorio

## Flujo de una Request

```
1. HTTP Request
   ↓
2. Handler (valida y parsea)
   ↓
3. Service (aplica lógica de negocio)
   ↓
4. Repository (accede a datos)
   ↓
5. Database
   ↓
6. Repository (retorna entidad)
   ↓
7. Service (procesa y retorna resultado)
   ↓
8. Handler (serializa y responde HTTP)
   ↓
9. HTTP Response
```

## Convenciones

### Naming

- Handlers: `*Handler` (ej: `AuthHandler`)
- Services: `*Service` (ej: `AuthService`)
- Repositories: `*Repository` (ej: `UserRepository`)

### Interfaces

- Todas las capas exponen interfaces
- Implementaciones son privadas
- Facilita testing y mocking

### Context

- Todas las operaciones aceptan `context.Context`
- Permite cancelación y timeouts
- Facilita tracing

### Errors

- Servicios retornan errores de negocio
- Handlers mapean a códigos HTTP
- Repositories retornan errores de datos

## Ejemplo Completo: Register Flow

```go
// 1. Handler recibe request
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. Llama al servicio
    result, err := h.authService.Register(r.Context(), req.Username, req.Password)

    // 3. Maneja response
    setAuthCookies(w, result.AccessToken, result.RefreshToken)
    respondJSON(w, toAuthResponse(result), http.StatusCreated)
}

// Service ejecuta lógica
func (s *authService) Register(ctx context.Context, username, password string) (*AuthResult, error) {
    // Valida password
    if len(password) < 4 {
        return nil, ErrWeakPassword
    }

    // Hash password
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    // Llama al repositorio
    user, err := s.userRepo.Create(ctx, username, string(hash))

    // Genera tokens
    tokens, _ := auth.GenerateTokenPair(user.ID, user.Username)

    // Guarda refresh token
    s.tokenRepo.Save(ctx, user.ID, tokens.RefreshToken, auth.GetRefreshTokenExpiry())

    return &AuthResult{User: user, ...}, nil
}

// Repository accede a BD
func (r *userRepository) Create(ctx context.Context, username, passwordHash string) (*db.User, error) {
    result, err := r.db.ExecContext(ctx,
        "INSERT INTO users (username, password_hash) VALUES (?, ?)",
        username, passwordHash)

    id, _ := result.LastInsertId()
    return &db.User{ID: id, Username: username, ...}, nil
}
```

## Testing

Con esta arquitectura, puedes testear cada capa:

```go
// Test Service (mock repository)
mockUserRepo := &MockUserRepository{}
mockTokenRepo := &MockTokenRepository{}
authService := service.NewAuthService(mockUserRepo, mockTokenRepo)

// Test Handler (mock service)
mockAuthService := &MockAuthService{}
authHandler := httpHandlers.NewAuthHandler(mockAuthService)
```
