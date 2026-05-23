# Fase 1.6: Service de Usuarios

## Objetivo
Implementar la capa de servicio para operaciones de usuario: registro, obtención, actualización y eliminación. El service valida datos de negocio y orquesta operaciones entre el repositorio y otras capas.

## Arquitectura

```
HTTP Handler (Fase 1.8)
        ↓
   UserService
   /    |    \    \
  /     |     \    \
Validate Register GetUser UpdateUser DeleteUser
  |     |     |     |     |
  ↓     ↓     ↓     ↓     ↓
UserRepository (Fase 1.4)
  |
  ↓
Database (via sqlc)
```

## Archivos Creados

### `internal/users/service/user_service.go`
- **Interfaz `UserService`** - Define contrato para operaciones de usuario
- **Struct `userService`** - Implementación con inyección de dependencias
- **Struct `User`** - Modelo de usuario con lógica de negocio (DBO → Model)
- **Struct `UserUpdateRequest`** - Request para actualización de usuario

## Interfaz UserService

```go
type UserService interface {
	Register(ctx context.Context, username, email, password, timezone string) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, userID string, updates *UserUpdateRequest) (*User, error)
	DeleteUser(ctx context.Context, userID string) error
}
```

## Métodos

### 1. Register(ctx, username, email, password, timezone)

**Responsabilidades:**
- Validar que `username` no existe (caso insensible)
- Validar que `email` no existe (caso insensible)
- Hashear contraseña con Argon2id (PasswordManager)
- Crear usuario en BD con ID generado (UUID v4)
- Role por defecto: `"P"` (Player, no Admin)
- Retornar usuario creado o error

**Validaciones:**
- Username ya existe → `username_already_exists`
- Email ya existe → `email_already_exists`
- Fallo al hashear contraseña → error interno

**Flujo:**
```
1. Check username uniqueness (repository.ExistsByUsername)
2. Check email uniqueness (repository.ExistsByEmail)
3. Hash password (passwordManager.HashPassword)
4. Generate UUID for user ID
5. Create user in DB (repository.CreateUser)
6. Log success
7. Return user model
```

**Ejemplo de uso:**
```go
userService := service.NewUserService(userRepo, passwordMgr)
user, err := userService.Register(ctx, "john_doe", "john@example.com", "secret123", "Europe/Madrid")
if err != nil {
    // Handle: username_already_exists, email_already_exists, internal errors
}
// user.ID, user.Username, user.CreatedAt, user.Timezone
```

### 2. GetUser(ctx, userID)

**Responsabilidades:**
- Obtener usuario por ID desde repositorio
- Retornar nil si no existe (no error)
- Convertir DBO a Model (dbUserToModel)

**Flujo:**
```
1. Get user from repository (repository.GetByID)
2. Return nil if not found (normal case)
3. Log debug message if not found
4. Convert DBO to Model
5. Return Model user
```

**Ejemplo de uso:**
```go
user, err := userService.GetUser(ctx, "user-uuid")
if err != nil {
    // Handle DB errors
}
if user == nil {
    // User not found (not an error, expected case)
}
```

### 3. UpdateUser(ctx, userID, updates)

**Responsabilidades:**
- Obtener usuario actual para validaciones
- Validar restricción de 30 días para cambio de timezone
- Actualizar usuario en BD
- Retornar usuario actualizado o error

**Validaciones:**
- Usuario no existe → `user_not_found`
- Cambio de timezone menos de 30 días → `timezone_recently_changed`
- Fallo al actualizar → error interno

**Restricción de Timezone:**
- Primera vez: Sin restricción (no hay `LastTimezoneModificationAt`)
- Cambios posteriores: Mínimo 30 días desde el último cambio
- Se verifica usando `LastTimezoneModificationAt`
- Se actualiza el timestamp en cada cambio

**Flujo:**
```
1. Get current user (repository.GetByID)
2. If timezone update requested:
   a. Check if user has timezone set
   b. Check if 30 days passed since last modification
   c. Return error if < 30 days
3. Prepare UpdateUserParams (ID, timezone, now timestamp)
4. Update user in DB (repository.UpdateUser)
5. Log success
6. Return updated user model
```

**Ejemplo de uso:**
```go
updates := &service.UserUpdateRequest{
    Timezone: pointer.String("Europe/Paris"),
}
user, err := userService.UpdateUser(ctx, "user-uuid", updates)
if err != nil {
    // Handle: user_not_found, timezone_recently_changed
}
// user with updated timezone
```

### 4. DeleteUser(ctx, userID)

**Responsabilidades:**
- Soft-delete usuario (marca `is_deleted=1`, `deleted_at`)
- Cascade delete aplicado por triggers en BD
- Retornar error solo si BD falla

**Cascada de eliminación:**
- user_sessions (todos asociados)
- companies (todas asociadas)
- company_inventory
- company_production_buildings
- company_sale_buildings
- production_runs
- sale_runs

**Flujo:**
```
1. Call repository.SoftDeleteUser(userID)
2. Log success or error
3. Return error if operation fails
```

**Ejemplo de uso:**
```go
err := userService.DeleteUser(ctx, "user-uuid")
if err != nil {
    // Handle DB errors
}
// User is now soft-deleted with all related data
```

## Inyección de Dependencias

```go
// En main.go (Fase 1.11)
queries := gen.New(db)
userRepo := repository.NewUserRepository(queries)
passwordMgr := auth.NewPasswordManager()
userService := service.NewUserService(userRepo, passwordMgr)

// Pasar userService a handlers (Fase 1.8)
userHandler := http.NewUserHandler(userService)
```

## Logging

**Info Level:**
- `Registering new user` - Inicio de registro
- `User successfully registered` - Usuario creado exitosamente
- `Updating user` - Inicio de actualización
- `User successfully updated` - Usuario actualizado exitosamente
- `Deleting user` - Inicio de eliminación
- `User successfully deleted` - Usuario eliminado exitosamente

**Debug Level:**
- `User not found` - En GetUser cuando usuario no existe

**Warning Level:**
- `Username already exists` - Intento de registro con username duplicado
- `Email already exists` - Intento de registro con email duplicado
- `User not found for update` - Intento de actualizar usuario que no existe
- `Timezone change attempted too soon` - Violación de restricción de 30 días

**Error Level:**
- `Failed to check username uniqueness` - Error en BD
- `Failed to check email uniqueness` - Error en BD
- `Failed to hash password` - Error en hashing (raramente ocurre)
- `Failed to create user in database` - Error en creación de usuario
- `Failed to get current user` - Error al obtener usuario en UpdateUser
- `Failed to update user` - Error al actualizar en BD
- `Failed to delete user` - Error al eliminar en BD

## Modelo User (Model Layer)

```go
type User struct {
	ID                         string     // UUID
	Username                   string     // Único, username para login
	Email                      string     // Único, email del usuario
	Role                       string     // "P" (Player) o "A" (Admin)
	Timezone                   *string    // Ej: "Europe/Madrid", nullable
	LastTimezoneModificationAt *time.Time // Timestamp del último cambio de timezone
	CreatedAt                  time.Time  // Timestamp de creación (UTC)
}
```

## Conversión DBO → Model

**Función:** `dbUserToModel(dbo *gen.User) *User`

Convierte el DBO (DB Object desde sqlc) al Model (lógica de negocio):

```go
DBO gen.User:
{
  ID: "uuid",
  Username: "john",
  Email: "john@example.com",
  PasswordHash: "$argon2id$...",
  Role: "P",
  Timezone: pointer("Europe/Madrid"),
  LastTimezoneModificationAt: pointer(time.Time),
  CreatedAt: time.Time,
  IsDeleted: 0,
  DeletedAt: nil
}
      ↓
Model User: (excluye PasswordHash, IsDeleted, DeletedAt)
{
  ID: "uuid",
  Username: "john",
  Email: "john@example.com",
  Role: "P",
  Timezone: pointer("Europe/Madrid"),
  LastTimezoneModificationAt: pointer(time.Time),
  CreatedAt: time.Time
}
```

## Constructor

```go
func NewUserService(
    userRepo repository.UserRepository,
    passwordManager *auth.PasswordManager,
) UserService {
    return &userService{
        userRepo:        userRepo,
        passwordManager: passwordManager,
    }
}
```

## Dependencias del Service

- **`repository.UserRepository`** (Fase 1.4)
  - CreateUser(ctx, params) - Crear usuario
  - GetByID(ctx, id) - Obtener por ID
  - ExistsByUsername(ctx, username) - Verificar existencia
  - ExistsByEmail(ctx, email) - Verificar existencia
  - UpdateUser(ctx, params) - Actualizar datos
  - SoftDeleteUser(ctx, id) - Eliminar (soft-delete)

- **`auth.PasswordManager`** (Fase 1.3)
  - HashPassword(password) - Hashear con Argon2id

- **`google/uuid`**
  - uuid.New().String() - Generar ID de usuario

## Próximas Fases

### Fase 1.7: Service de Autenticación
- AuthService con Login y Logout
- Validación de credenciales
- Generación de tokens de sesión
- Gestión de sesiones en cache

### Fase 1.8: Handlers HTTP
- Registrar handlers para endpoints de autenticación y usuarios
- Validación HTTP con go-playground/validator
- Conversión Model → DTO para respuestas

### Fase 1.9: Middleware de Autenticación
- Validación de session tokens
- Manejo de refresh tokens
- Inyección de usuario en contexto

## Notas de Implementación

### Atomic Operations
- CreateUser es atomic (BD se encarga via triggers)
- UpdateUser es atomic (single SQL statement)
- DeleteUser es atomic (triggers en BD)

### Error Handling
- Los errores de BD se wrappean con `fmt.Errorf`
- Se loguean a nivel Error
- Se retorna el error específico (ej: `username_already_exists`) cuando es validación de negocio

### Timezone Logic
- Timezone es optional (nullable en BD)
- Primera vez sin restricción
- Cambios posteriores: mínimo 30 días
- Se almacena UTC timestamp de último cambio
- El cliente envía timezone en formato IANA (ej: "Europe/Madrid")

### Security
- Contraseñas nunca se exponen en modelos
- Se hashean inmediatamente en Register
- PasswordHash removido en conversión DBO → Model
- No se loguean contraseñas ni dados sensibles

### Concurrency
- Múltiples goroutines pueden llamar Register simultáneamente
- Uniqueness garantizada por índices UNIQUE en BD
- Logs con contexto para auditoría

## Ejemplo Completo: Flujo de Registro

```go
// 1. User hace POST /api/v1/auth/register
// 2. Handler valida input HTTP
// 3. Handler llama userService.Register()
userService.Register(ctx, "alice", "alice@example.com", "secret123", "America/NewYork")
//   a. Check username no existe ✓
//   b. Check email no existe ✓
//   c. Hash password: "$argon2id$v=19$m=65536,t=3,p=4$..." ✓
//   d. Create DB user con ID="<uuid>" ✓
//   e. Retrieve created user ✓
//   f. Return User model ✓
// 4. Handler convierte Model → DTO
// 5. HTTP response 201 Created con user DTO

// Resultado:
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "alice",
    "email": "alice@example.com",
    "role": "P",
    "timezone": "America/NewYork",
    "last_timezone_modification_at": null,
    "created_at": "2026-05-23T10:30:00Z"
  }
}
```

## Testing

En Fase 1.10, se necesitan tests:

**Unit Tests:**
```go
TestRegister_Success
TestRegister_UsernameExists
TestRegister_EmailExists
TestRegister_HashPasswordFails
TestGetUser_Found
TestGetUser_NotFound
TestUpdateUser_Success
TestUpdateUser_TimezoneRestriction
TestUpdateUser_UserNotFound
TestDeleteUser_Success
```

**Mocks Necesarios:**
- Mock de UserRepository (para testear service sin BD)
- Mock de PasswordManager (para testear registro sin hashing real)

## Estado

✅ **Fase 1.6 Completada** - Service de Usuarios implementado

**Próximo paso:** Fase 1.7 - Service de Autenticación (Login, Logout)
