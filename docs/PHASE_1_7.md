# Phase 1.7: Session Cache and Authentication Service

## Overview

**Phase 1.7** implements the authentication service and session caching layer. This phase adds:

1. **Session Cache** (`internal/auth/session_cache.go`) - In-memory cache for fast session lookups during auth operations
2. **Auth Service** (`internal/auth/service/auth_service.go`) - Business logic for user authentication (Login, Logout)
3. **Repository Interface** (`internal/auth/repository.go`) - Interface definition for session persistence

The architecture follows a **dual-layer persistence model**:
- **Database Layer**: Sessions persisted to SQLite for durability across restarts
- **Cache Layer**: Active sessions stored in memory for fast validation and refresh token verification

## Architecture

### Session Flow Diagram

```
Login Request (username, password)
    ↓
    ├─→ Verify credentials (password check with Argon2id)
    ├─→ Generate sessionID (UUID)
    ├─→ Generate verificationString (UUID)
    ├─→ Compute tokenHash (SHA256)
    ├─→ Create session in Database (via SessionRepository)
    ├─→ Cache session in Memory (SessionCache) for fast lookups
    ├─→ Generate Session Token (JWT, 1 minute expiry)
    ├─→ Generate Refresh Token (JWT, 300 days expiry)
    └─→ Return LoginResponse with both tokens

Subsequent API Call (with SessionToken)
    ↓
    ├─→ Validate JWT signature and expiration
    ├─→ Extract sessionID from claims
    ├─→ Lookup session in Cache (O(1) operation)
    └─→ Allow request or redirect to login

Token Refresh (with RefreshToken)
    ↓
    ├─→ Validate JWT signature
    ├─→ Extract sessionID and verificationString from claims
    ├─→ Lookup session in Cache
    ├─→ Verify tokenHash matches SHA256(verificationString)
    └─→ Generate new Session Token

Logout Request (with SessionToken)
    ↓
    ├─→ Extract sessionID from token
    ├─→ Revoke in Database (sets revoked_at timestamp)
    ├─→ Revoke in Cache (mark RevokedAt)
    └─→ Return success
```

### Component Responsibilities

| Component | Responsibility | Layer |
|-----------|-----------------|-------|
| `SessionCache` | In-memory session storage with thread-safe Get/Store/Revoke operations | Cache |
| `AuthService` | Login/Logout business logic, credential validation, token generation | Service |
| `UserSessionRepository` | Persist sessions to SQLite, retrieve, revoke, count active sessions | Repository |
| `JWTManager` | Generate and validate JWT tokens (session + refresh) | Auth |
| `PasswordManager` | Hash and verify passwords using Argon2id | Auth |

## SessionData Structure

```go
// SessionData is defined in internal/auth/models.go
type SessionData struct {
    SessionID          string       // UUID, unique session identifier
    UserID             string       // UUID of the authenticated user
    CompanyID          *string      // UUID of associated company (nullable)
    VerificationString string       // UUID used to generate refresh token
    TokenHash          string       // SHA256 hash of verificationString
    ExpiresAt          time.Time    // When session expires (300 days from creation)
    RevokedAt          *time.Time   // When session was revoked (nil if active)
    CreatedAt          time.Time    // When session was created
}
```

### Session Validity Rules

A session is **valid** when:
- `RevokedAt == nil` (not revoked)
- `time.Now().Before(ExpiresAt)` (not expired)

A session is **invalid** when:
- Explicitly revoked (`RevokedAt != nil`)
- Expired (`time.Now().After(ExpiresAt)`)
- Not found in database or cache

## SessionCache Implementation

### Purpose
Thread-safe in-memory cache for active sessions with O(1) lookup performance.

### Key Methods

#### `NewSessionCache() *SessionCache`
Creates a new session cache with empty sessions map.

```go
cache := auth.NewSessionCache()
```

#### `Store(sessionID string, data SessionData)`
Adds or updates a session in the cache. Thread-safe with RWMutex.

```go
cache.Store(sessionID, SessionData{
    UserID:             "user-123",
    SessionID:          "session-456",
    VerificationString: "verify-789",
    TokenHash:          sha256Hash,
    ExpiresAt:          time.Now().Add(300 * 24 * time.Hour),
    CreatedAt:          time.Now().UTC(),
})
```

#### `Get(sessionID string) *SessionData`
Retrieves a session by ID. Returns nil if:
- Session not found
- Session expired (`time.Now().After(ExpiresAt)`)
- Session revoked (`RevokedAt != nil`)

```go
session := cache.Get("session-456")
if session == nil {
    // Session invalid or not found
}
```

#### `Revoke(sessionID string)`
Marks a session as revoked by setting `RevokedAt = now`. Subsequent `Get()` calls return nil.

```go
cache.Revoke("session-456")
```

#### `Delete(sessionID string)`
Removes a session entirely from cache (hard delete).

```go
cache.Delete("session-456")
```

#### `DeleteExpired() int64`
Removes all expired or revoked sessions. Returns count deleted.

```go
deleted := cache.DeleteExpired()
log.Info().Int64("count", deleted).Msg("Cleaned expired sessions from cache")
```

#### `Count() int64`
Returns count of active (non-expired, non-revoked) sessions in cache.

```go
activeCount := cache.Count()
```

### Thread Safety

SessionCache uses `sync.RWMutex` for thread-safe operations:
- **Read operations** (Get, Count): Acquire RLock
- **Write operations** (Store, Revoke, Delete, DeleteExpired): Acquire Lock

All operations are atomic with respect to the mutex.

## AuthService Implementation

### Purpose
Implements `AuthService` interface with business logic for Login and Logout operations.

### AuthService Interface

```go
type AuthService interface {
    Login(ctx context.Context, username, password string) (*LoginResponse, error)
    Logout(ctx context.Context, sessionID string) error
}
```

### Constructor: NewAuthService

```go
func NewAuthService(
    userRepo repository.UserRepository,
    sessionRepo auth.UserSessionRepository,
    passwordManager *auth.PasswordManager,
    jwtManager *auth.JWTManager,
    sessionCache *auth.SessionCache,
) AuthService
```

**Dependencies:**
- `userRepo`: Retrieve user by username
- `sessionRepo`: Persist sessions to database
- `passwordManager`: Verify password against hash
- `jwtManager`: Generate session and refresh tokens
- `sessionCache`: Store sessions in memory

**Injection Pattern:**
Uses constructor injection with interfaces for testability. All dependencies are required.

```go
authService := service.NewAuthService(
    userRepo,           // *repository.userRepository
    sessionRepo,        // *repository.userSessionRepository
    passwordManager,    // *auth.PasswordManager
    jwtManager,         // *auth.JWTManager
    sessionCache,       // *auth.SessionCache
)
```

### LoginResponse Structure

```go
type LoginResponse struct {
    User         *models.User // User model (DTO)
    SessionToken string       // JWT session token (1 min expiry)
    RefreshToken string       // JWT refresh token (300 days expiry)
    ExpiresAt    int64        // Session token expiration as Unix timestamp
}
```

### Method: Login

**Signature:**
```go
func (s *authService) Login(ctx context.Context, username, password string) (*LoginResponse, error)
```

**Process:**

1. **Get user by username** from repository
   - If not found: return error "invalid_credentials"
   - Log debug if not found

2. **Verify password** with timing-safe comparison
   - Use PasswordManager.VerifyPassword()
   - If mismatch: return error "invalid_credentials"
   - Log debug if mismatch

3. **Generate session identifiers**
   - `sessionID = uuid.New().String()` (unique session ID)
   - `verificationString = uuid.New().String()` (for refresh token)
   - `tokenHash = SHA256(verificationString)` (for token validation)

4. **Create session in database**
   - Call SessionRepository.CreateSession() with CreateSessionParams
   - Set ExpiresAt to 300 days from now
   - If error: return wrapped error

5. **Store session in cache**
   - Call SessionCache.Store() with SessionData
   - Enables fast lookups during API requests
   - No TTL needed; Get() checks ExpiresAt

6. **Generate tokens**
   - Session Token: 1 minute expiry, contains userID + sessionID
   - Refresh Token: 300 days expiry, contains userID + sessionID + verificationString

7. **Return LoginResponse**
   - User model (DTO format)
   - Both tokens
   - Session token expiration timestamp

**Error Handling:**
- `"invalid_credentials"` - User not found or password mismatch
- `"failed to create session"` - Database persistence error
- `"failed to generate session token"` - JWT generation error
- `"failed to generate refresh token"` - JWT generation error

**Logging:**
```go
// Success case
log.Info().
    Str("user_id", user.ID).
    Str("username", username).
    Str("session_id", sessionID).
    Msg("User successfully logged in")

// Failure cases
log.Debug().Str("username", username).Msg("User not found during login")
log.Debug().Str("username", username).Msg("Password verification failed")
log.Error().Err(err).Str("user_id", user.ID).Msg("Failed to create session in database")
```

### Method: Logout

**Signature:**
```go
func (s *authService) Logout(ctx context.Context, sessionID string) error
```

**Process:**

1. **Verify session exists in database**
   - Call SessionRepository.GetBySessionID()
   - If not found: return error "session_not_found"

2. **Revoke session in database**
   - Call SessionRepository.RevokeSession()
   - Sets `revoked_at` timestamp in database
   - If error: return wrapped error

3. **Revoke session in cache**
   - Call SessionCache.Revoke()
   - Marks session as revoked in memory
   - Subsequent Get() calls return nil

**Error Handling:**
- `"session_not_found"` - Session not in database
- `"failed to revoke session"` - Database operation error

**Logging:**
```go
// Attempt
log.Info().Str("session_id", sessionID).Msg("Logout attempt")

// Success
log.Info().Str("session_id", sessionID).Msg("Session successfully revoked")

// Failure
log.Debug().Str("session_id", sessionID).Msg("Session not found during logout")
log.Error().Err(err).Str("session_id", sessionID).Msg("Failed to revoke session")
```

## UserSessionRepository Interface

**Location:** `internal/auth/repository.go`

Defines database operations for session persistence.

### Methods

#### `CreateSession(ctx, params *gen.CreateSessionParams) (*gen.UserSession, error)`
Inserts a new session into database and returns the created session.

#### `GetByID(ctx, id string) (*gen.UserSession, error)`
Retrieves session by primary key ID. Returns nil if not found.

#### `GetBySessionID(ctx, sessionID string) (*gen.UserSession, error)`
Retrieves session by unique `session_id` field. Returns nil if not found.

#### `GetByTokenHash(ctx, tokenHash string) (*gen.UserSession, error)`
Retrieves session by `token_hash`. Used for refresh token validation. Returns nil if not found.

#### `RevokeSession(ctx, sessionID string) error`
Marks session as revoked by setting `revoked_at = now`. Error if operation fails.

#### `DeleteExpiredSessions(ctx) error`
Soft-deletes all expired or revoked sessions. Scheduled cleanup task.

#### `CountActiveSessions(ctx, userID string) (int64, error)`
Returns count of active (non-revoked, non-expired) sessions for a user.

#### `GetUserSessions(ctx, userID string) ([]*gen.UserSession, error)`
Returns slice of all non-deleted sessions for a user.

## Dependency Injection Pattern

### Service Initialization

```go
// Create individual components
passwordManager := auth.NewPasswordManager()
jwtManager := auth.NewJWTManager()
sessionCache := auth.NewSessionCache()

// Create repositories
queries := gen.New(db)
userRepo := repository.NewUserRepository(queries)
sessionRepo := repository.NewUserSessionRepository(queries)

// Create service
authService := service.NewAuthService(
    userRepo,
    sessionRepo,
    passwordManager,
    jwtManager,
    sessionCache,
)
```

### Testing Considerations

All dependencies are injected as interfaces, enabling:
- Mock repositories for unit tests
- Fake JWT managers to control token generation
- Mock password managers for testing invalid credentials
- Test session caches with controlled state

Example mock:
```go
type mockUserRepository struct{}
func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
    return &models.User{ID: "test-user", Username: "test"}, nil
}
```

## Token Management

### Session Token (Short-lived)
- **Expiry**: 1 minute (configurable via `JWT_SESSION_EXPIRY` env var)
- **Contents**: userID, sessionID, companyID (nullable)
- **Storage**: httpOnly cookie (automatic renewal in middleware)
- **Validation**: Signature + expiration check in auth middleware

### Refresh Token (Long-lived)
- **Expiry**: 300 days (configurable via `JWT_REFRESH_EXPIRY` env var)
- **Contents**: userID, sessionID, verificationString
- **Storage**: httpOnly cookie
- **Validation**: 
  - Signature check
  - tokenHash verification (SHA256 of verificationString)
  - Session not revoked

### Refresh Flow

When session token expires:
1. Client sends refresh token
2. Middleware extracts verificationString from token
3. SessionCache.Get() retrieves session
4. Verify tokenHash matches SHA256(verificationString)
5. Generate new session token
6. Set new token in cookie

## Implementation Checklist

- [x] SessionCache created with thread-safe operations
- [x] SessionData structure defined in models.go
- [x] AuthService interface defined
- [x] AuthService.Login() implemented with full flow
- [x] AuthService.Logout() implemented with DB + cache revocation
- [x] UserSessionRepository interface extracted
- [x] Compilation verified - no errors
- [ ] PHASE_1_7.md documentation (THIS FILE)
- [ ] Unit tests for Login/Logout
- [ ] HTTP handlers using AuthService (Phase 1.8)
- [ ] Auth middleware for token validation (Phase 1.9)

## Related Phases

- **Phase 1.6**: User Service (Register, GetUser, UpdateUser, DeleteUser)
- **Phase 1.5**: User Session Repository (database persistence)
- **Phase 1.3**: JWT and Password managers
- **Phase 1.8**: HTTP Handlers (RegisterHandler, LoginHandler, LogoutHandler)
- **Phase 1.9**: Auth Middleware (SessionToken validation)

## Error Codes

| Code | Description | Recovery |
|------|-------------|----------|
| `invalid_credentials` | User not found or wrong password | Show login form again |
| `session_not_found` | Logout attempted on non-existent session | Already logged out |
| `failed to create session` | Database error during session creation | Retry login |
| `failed to generate tokens` | JWT generation error | Retry login |
| `session_expired` | Session token expired | Use refresh token |
| `token_hash_mismatch` | Refresh token tampered | Force re-login |

## Performance Characteristics

| Operation | Time Complexity | Notes |
|-----------|-----------------|-------|
| SessionCache.Get() | O(1) | Hash map lookup |
| SessionCache.Store() | O(1) | Hash map insert/update |
| SessionCache.Revoke() | O(1) | Hash map update (RevokedAt) |
| SessionCache.Delete() | O(1) | Hash map delete |
| SessionCache.Count() | O(n) | Iterates all sessions |
| SessionCache.DeleteExpired() | O(n) | Iterates all sessions |
| AuthService.Login() | O(1) + Argon2id | Password hashing dominant |
| AuthService.Logout() | O(1) | Cache revoke is O(1) |

Session cache is highly performant for lookups. Expired session cleanup should be scheduled periodically (not on every request).

---

**Documentation Version:** 1.0  
**Last Updated:** Phase 1.7 Completion  
**Author:** Development Team
