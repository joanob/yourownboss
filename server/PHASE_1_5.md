# Fase 1.5 - Repository de Sesiones de Usuario

## Objetivo
Implementar la capa de acceso a datos (Repository) para operaciones CRUD de sesiones de usuario usando sqlc.

## Archivos Creados

### 1. `internal/auth/queries.sql` ✅
Queries SQL para operaciones de sesiones, procesadas por sqlc.

**Queries implementadas:**
- `CreateSession` - Insertar nueva sesión
- `GetSessionByID` - Obtener por ID (no soft-deleted)
- `GetSessionBySessionID` - Obtener por session_id (no soft-deleted)
- `GetSessionByTokenHash` - Obtener por token_hash (no soft-deleted)
- `RevokeSession` - Marcar como revocada (set revoked_at)
- `DeleteExpiredSessions` - Soft-delete de sesiones expiradas o revocadas
- `CountActiveSessionsByUserID` - Contar sesiones activas de usuario
- `GetUserSessionsByUserID` - Obtener todas las sesiones de un usuario

**Características:**
- Todos los GETs excluyen sesiones soft-deleted (`is_deleted = 0`)
- El CREATE inicia `created_at` y `is_deleted` automáticamente en la BD
- El REVOKE actualiza `revoked_at` sin marcar como borrado (soft-delete manual)
- DeleteExpiredSessions busca `expires_at < NOW()` OR `revoked_at IS NOT NULL`
- CountActiveSessions cuenta solo sesiones no revocadas y no expiradas
- GetUserSessions devuelve array de sesiones (`:many`)

### 2. `internal/db/gen/queries.sql.go` ✅ (Generado por sqlc)
Código Go tipado generado automáticamente por sqlc.

**Tipos generados:**
- `UserSession` - DBO (Data Base Object) con todos los campos de la tabla user_sessions
- `CreateSessionParams` - Parámetros para CreateSession
- Otros tipos para queries `:many`

**Funciones generadas:**
```go
func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) error
func (q *Queries) GetSessionByID(ctx context.Context, id string) (UserSession, error)
func (q *Queries) GetSessionBySessionID(ctx context.Context, sessionID string) (UserSession, error)
func (q *Queries) GetSessionByTokenHash(ctx context.Context, tokenHash string) (UserSession, error)
func (q *Queries) RevokeSession(ctx context.Context, sessionID string) error
func (q *Queries) DeleteExpiredSessions(ctx context.Context) error
func (q *Queries) CountActiveSessionsByUserID(ctx context.Context, userID string) (int64, error)
func (q *Queries) GetUserSessionsByUserID(ctx context.Context, userID string) ([]UserSession, error)
```

### 3. `internal/auth/repository/user_session_repo.go` ✅
Repository que implementa la interfaz UserSessionRepository y encapsula acceso a BD.

**Interfaz UserSessionRepository:**
```go
type UserSessionRepository interface {
	CreateSession(ctx context.Context, id, userID, sessionID, verificationString, tokenHash string, expiresAt time.Time) error
	GetByID(ctx context.Context, id string) (*gen.UserSession, error)
	GetBySessionID(ctx context.Context, sessionID string) (*gen.UserSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*gen.UserSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
	DeleteExpiredSessions(ctx context.Context) error
	CountActiveSessions(ctx context.Context, userID string) (int64, error)
	GetUserSessions(ctx context.Context, userID string) ([]gen.UserSession, error)
}
```

**Métodos implementados:**

#### CreateSession
- Parámetros: id, userID, sessionID, verificationString, tokenHash, expiresAt
- Logging: Info al éxito, Error si falla
- Retorna: error si falla
- Nota: ID es generado por el servicio (UUID)

#### GetByID / GetBySessionID / GetByTokenHash
- Retorna: *gen.UserSession (DBO), nil (no existe), error (error BD)
- Logging: Debug si no existe, Error si falla en BD
- Notas: Excluye soft-deleted automáticamente (BD lo filtra)

#### RevokeSession
- Parámetros: sessionID
- Realiza: UPDATE set revoked_at = CURRENT_TIMESTAMP
- Logging: Info al éxito, Error si falla
- Nota: NO es un soft-delete, solo marca como revocada

#### DeleteExpiredSessions
- Parámetros: ninguno
- Realiza: Soft-delete de sesiones expiradas O revocadas
- Logging: Debug al éxito, Error si falla
- Nota: Se ejecuta por un cleaner cada 6 horas (Fase posterior)

#### CountActiveSessions
- Parámetros: userID
- Retorna: int64 (count), error (error BD)
- Query: Cuenta sesiones con revoked_at IS NULL AND expires_at > NOW()
- Logging: Error si falla

#### GetUserSessions
- Parámetros: userID
- Retorna: []gen.UserSession, error (error BD)
- Logging: Error si falla
- Nota: Devuelve todas las sesiones no soft-deleted del usuario

**Inyección de dependencias:**
```go
// Constructor
func NewUserSessionRepository(queries *gen.Queries) UserSessionRepository {
    return &userSessionRepository{
        queries: queries,
    }
}

// Uso en main.go (después)
db := sql.Open("sqlite", databasePath)
queries := gen.New(db)
sessionRepo := repository.NewUserSessionRepository(queries)
```

## Arquitectura DBO/Model/DTO

**DBO (Data Base Object)** = `gen.UserSession` (generado por sqlc)
- Mapeo exacto de tabla user_sessions
- Todos los campos: ID, UserID, SessionID, VerificationString, TokenHash, ExpiresAt, RevokedAt, CreatedAt, IsDeleted, DeletedAt
- Devuelto por Repository (capa de BD)

**Model** = (Crear en Fase posterior)
- Estructura interna con reglas de negocio
- Métodos: IsValid(), IsExpired(), IsRevoked()
- Usado por Service

**DTO** = (No exponer directamente a API)
- Las sesiones NO se devuelven al cliente
- El cliente recibe solo session_token y refresh_token como cookies

**Flujo:**
```
BD (tabla user_sessions)
  ↓
DBO (gen.UserSession) ← Repository devuelve esto
  ↓
Model (SessionData con lógica) ← Service convierte con lógica
  ↓
Cache de sesiones ← AuthService almacena aquí
  ↓
Validación en middleware
```

## Uso en el Servicio (Fase 1.7)

El servicio de autenticación usará el repository así:

```go
// En service/auth_service.go
type AuthService struct {
    sessionRepo repository.UserSessionRepository
    jwtManager  *auth.JWTManager
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, error) {
    // Validar credenciales
    user, err := s.userRepo.GetByUsername(ctx, username)
    if user == nil || !pm.VerifyPassword(password, user.PasswordHash) {
        return "", "", errors.New("INVALID_CREDENTIALS")
    }

    // Generar tokens
    sessionToken, _, err := s.jwtManager.GenerateSessionToken(user.ID, gen.GenerateSessionID(), nil)
    refreshToken, verificationString, _, err := s.jwtManager.GenerateRefreshToken(user.ID, sessionID)

    // Crear sesión en BD
    sessionID := gen.GenerateSessionID()
    err = s.sessionRepo.CreateSession(ctx, uuid.New().String(), user.ID, sessionID, verificationString, hashToken(refreshToken), expiresAt)
    if err != nil {
        return "", "", err
    }

    // Guardar en cache de sesiones
    s.sessionCache.Set(sessionID, SessionData{...})

    return sessionToken, refreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
    // Revocar en BD
    err := s.sessionRepo.RevokeSession(ctx, sessionID)
    if err != nil {
        return err
    }

    // Invalidar en cache
    s.sessionCache.Delete(sessionID)

    return nil
}
```

## Limpieza de Sesiones Expiradas

Las sesiones expiradas o revocadas se limpian automáticamente con:

```go
// En main.go o en un cleaner
ticker := time.NewTicker(6 * time.Hour)
go func() {
    for range ticker.C {
        err := sessionRepo.DeleteExpiredSessions(context.Background())
        if err != nil {
            log.Error().Err(err).Msg("Error limpiando sesiones expiradas")
        }
    }
}()
```

## Testing (Próximo)

En Fase 1.10, crear tests para:
- CreateSession con parámetros válidos
- GetByID / GetBySessionID / GetByTokenHash con casos de éxito y no existe
- RevokeSession verifica que revoked_at se actualiza (no soft-delete)
- DeleteExpiredSessions encuentra y marca como borradas sesiones expiradas
- CountActiveSessions cuenta correctamente sesiones activas
- GetUserSessions retorna todas las sesiones del usuario
- Todas las funciones con transacciones y rollbacks

## Próximos Pasos (Fase 1.6)

**Fase 1.6: Service de Usuarios**
- internal/users/service/user_service.go
- Métodos: Register, GetUser, UpdateUser, DeleteUser
- Usa UserRepository del usuario
