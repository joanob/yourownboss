# Fase 1.4 - Repository de Usuarios

## Objetivo
Implementar la capa de acceso a datos (Repository) para operaciones CRUD de usuarios usando sqlc.

## Archivos Creados

### 1. `internal/users/queries.sql` ✅
Queries SQL para operaciones de usuarios, procesadas por sqlc.

**Queries implementadas:**
- `CreateUser` - Insertar nuevo usuario
- `GetUserByID` - Obtener por ID (no soft-deleted)
- `GetUserByUsername` - Obtener por username (no soft-deleted)
- `GetUserByEmail` - Obtener por email (no soft-deleted)
- `UpdateUser` - Actualizar timezone
- `SoftDeleteUser` - Marcar como borrado
- `CountUserByUsername` - Verificar existencia de username
- `CountUserByEmail` - Verificar existencia de email

**Características:**
- Todos los GETs excluyen usuarios soft-deleted (`is_deleted = 0`)
- El CREATE inicia `created_at` y `is_deleted` automáticamente en la BD
- El UPDATE modifica `timezone` y `last_timezone_modification_at`
- El DELETE es soft-delete (marca `is_deleted = 1, deleted_at = CURRENT_TIMESTAMP`)

### 2. `internal/queries.sql.go` ✅ (Generado por sqlc)
Código Go tipado generado automáticamente por sqlc.

**Tipos generados:**
- `User` - DBO (Data Base Object) con todos los campos de la tabla users
- `CreateUserParams` - Parámetros para CreateUser
- `UpdateUserParams` - Parámetros para UpdateUser

**Funciones generadas:**
```go
// Queries interface - método que genera sqlc en Queries struct
type Queries struct {
    db *sql.DB
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error
func (q *Queries) GetUserByID(ctx context.Context, id string) (User, error)
func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error)
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error)
func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error
func (q *Queries) SoftDeleteUser(ctx context.Context, id string) error
func (q *Queries) CountUserByUsername(ctx context.Context, username string) (int64, error)
func (q *Queries) CountUserByEmail(ctx context.Context, email string) (int64, error)
```

### 3. `internal/users/repository/user_repo.go` ✅
Repository que implementa la interfaz UserRepository y encapsula acceso a BD.

**Interfaz UserRepository:**
```go
type UserRepository interface {
	CreateUser(ctx context.Context, id, username, email, passwordHash, role string, timezone *string) error
	GetByID(ctx context.Context, id string) (*internal.User, error)
	GetByUsername(ctx context.Context, username string) (*internal.User, error)
	GetByEmail(ctx context.Context, email string) (*internal.User, error)
	UpdateUser(ctx context.Context, id string, timezone *string, lastTimezoneModAt *time.Time) error
	SoftDeleteUser(ctx context.Context, id string) error
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
```

**Métodos implementados:**

#### CreateUser
- Parámetros: id, username, email, passwordHash, role, timezone (nullable)
- Validación: Ninguna en repository (hecha en service)
- Logging: Info al éxito, Error si falla
- Retorna: error si falla

#### GetByID / GetByUsername / GetByEmail
- Retorna: *User (DBO), nil (no existe), error (error BD)
- Logging: Debug si no existe, Error si falla en BD
- Notas: Excluye soft-deleted automáticamente (BD lo filtra)

#### UpdateUser
- Parámetros: id, timezone (nullable), lastTimezoneModAt (nullable)
- Actualiza: timezone y last_timezone_modification_at
- Logging: Info al éxito, Error si falla
- Nota: Solo actualiza campos de timezone

#### SoftDeleteUser
- Parámetros: id
- Realiza: UPDATE set is_deleted=1, deleted_at=CURRENT_TIMESTAMP
- Logging: Info al éxito, Error si falla
- Nota: Transacción NO automática, puede ser llamada dentro de transacción mayor

#### ExistsByUsername / ExistsByEmail
- Parámetros: username / email
- Retorna: bool (existe), error (error BD)
- Logging: Error si falla
- Nota: Usa COUNT query para verificar existencia rápidamente

**Inyección de dependencias:**
```go
// Constructor
func NewUserRepository(queries *internal.Queries) UserRepository {
    return &userRepository{
        queries: queries,
    }
}

// Uso en main.go (después)
db := sql.Open("sqlite", databasePath)
queries := internal.New(db)
userRepo := repository.NewUserRepository(queries)
```

## Arquitectura DBO/Model/DTO

**DBO (Data Base Object)** = `internal.User` (generado por sqlc)
- Mapeo exacto de tabla users
- Todos los campos: ID, Username, Email, PasswordHash, Role, Timezone, LastTimezoneModificationAt, CreatedAt, IsDeleted, DeletedAt
- Devuelto por Repository (capa de BD)

**Model** = `User` en `internal/users/models/user.go` (ya existe desde Fase 1.2)
- Estructura interna con reglas de negocio
- Métodos: FromDBO(), ToDBO(), CanChangeTimezone(), TimezoneChangeBlockedUntil()
- Usado por Service

**DTO** = `UserDTO` en `internal/users/models/user.go` (ya existe)
- Para respuestas API: id, username, email, role, timezone, createdAt
- SIN password_hash, is_deleted, deleted_at

**Flujo:**
```
DB (tabla users)
  ↓
DBO (internal.User) ← Repository devuelve esto
  ↓
Model (User con lógica) ← Service convierte con FromDBO()
  ↓
DTO (UserDTO) ← Handler convierte con ToDBO() para respuesta
  ↓
API Response JSON
```

## Configuración de sqlc

**Archivo `sqlc.yaml`:**
```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "./internal/**/*.sql"
    schema: "./internal/db/migrations/schema.sql"
    gen:
      go:
        package: "sql"
        out: "./internal"
        emit_json_tags: true
        emit_pointers_for_null_types: true
```

**Cambios necesarios:**
- Removido campo `project.name` (no soportado en v2)
- Actualizados paths a rutas relativas (`./`)
- Removido `query_parameter_limit: 1`

**Schema.sql (ajustes):**
- Reemplazados UNIQUE constraints con WHERE por CREATE UNIQUE INDEX
- Línea 131: `UNIQUE (company_building_id, is_collected) WHERE is_collected = 0` → `CREATE UNIQUE INDEX idx_production_runs_active`
- Línea 158: `UNIQUE (company_sale_building_id, is_collected) WHERE is_collected = 0` → `CREATE UNIQUE INDEX idx_sale_runs_active`

## Uso en el Servicio (Fase 1.6)

El servicio usará el repository así:

```go
// En service/user_service.go
type UserService struct {
    userRepo repository.UserRepository
}

func (s *UserService) Register(ctx context.Context, username, email, password, timezone string) (*User, error) {
    // Validar que username no existe
    exists, err := s.userRepo.ExistsByUsername(ctx, username)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("USERNAME_ALREADY_EXISTS")
    }

    // Hash password
    pm := auth.NewPasswordManager()
    hash, err := pm.HashPassword(password)
    if err != nil {
        return nil, err
    }

    // Crear usuario
    userID := uuid.New().String()
    err = s.userRepo.CreateUser(ctx, userID, username, email, hash, "P", &timezone)
    if err != nil {
        return nil, err
    }

    // Devolver usuario creado
    userDBO, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Convertir DBO → Model → DTO
    return userDBO.ToModel(), nil
}
```

## Testing (Próximo)

En Fase 1.10, crear tests para:
- CreateUser con validaciones (username único, email único)
- GetByID / GetByUsername / GetByEmail con casos de éxito y no existe
- UpdateUser con timezone válido
- SoftDeleteUser verifica soft-delete
- ExistsByUsername / ExistsByEmail
- Todas las funciones con transacciones y rollbacks

## Próximos Pasos (Fase 1.5)

**Fase 1.5: Repository de Sesiones de Usuario**
- Crear `internal/auth/repository/user_session_repo.go`
- sqlc queries para session CRUD
- Métodos: CreateSession, GetByID, RevokeSession, DeleteExpiredSessions
