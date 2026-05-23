# Phase 1.8: HTTP Handlers for Auth and Users

## Overview

**Phase 1.8** implements all HTTP request/response handlers for authentication and user management endpoints. This phase creates:

1. **Auth Handlers** - RegisterHandler, LoginHandler, LogoutHandler
2. **User Handlers** - GetMeHandler, UpdateMeHandler
3. **Request/Response DTOs** - Type-safe request validation and response formatting
4. **Route Registration** - Chi router setup for all auth and user endpoints

The architecture follows the handler pattern with dependency injection, input validation, and consistent error responses.

## Architecture

### Handler Flow

```
HTTP Request
    ↓
    ├─→ Handler receives request (w http.ResponseWriter, r *http.Request)
    ├─→ Parse JSON body into Request struct
    ├─→ Validate using go-playground/validator
    ├─→ Call service layer (injected)
    ├─→ Build Response DTO from service result
    ├─→ Set HTTP status + headers
    ├─→ Encode response as JSON
    └─→ Return to client

Error Flow:
    ├─→ Catch service error
    ├─→ Map to error code (INVALID_CREDENTIALS, USER_NOT_FOUND, etc.)
    ├─→ Return appropriate HTTP status (400, 401, 404, 409, 500)
    ├─→ Include ErrorResponse with code and message
    └─→ Log error at appropriate level (Debug for expected, Error for failures)
```

### Handler Structure

Each handler implements the same pattern:

```go
type MyHandler struct {
    service  SomeService
    validator *validator.Validate  // optional, if needed
}

func NewMyHandler(service SomeService, validator *validator.Validate) *MyHandler {
    return &MyHandler{service, validator}
}

func (h *MyHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Extract context
    // 2. Parse + validate request
    // 3. Call service
    // 4. Handle errors
    // 5. Build + return response
}
```

## DTOs and Request/Response Types

### Location: `internal/auth/http/dto.go`

```go
// RegisterRequest - POST /api/v1/auth/register
type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=128"`
    Timezone string `json:"timezone" validate:"required"`
}

// LoginRequest - POST /api/v1/auth/login
type LoginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
}

// UserDTO - Used in all responses
type UserDTO struct {
    ID        string `json:"id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    Timezone  string `json:"timezone"`
    CreatedAt string `json:"created_at"`
}

// LoginResponse - Response body for successful login
type LoginResponse struct {
    User             *UserDTO `json:"user"`
    SessionToken     string   `json:"session_token"`
    RefreshToken     string   `json:"refresh_token"`
    SessionExpiresAt int64    `json:"session_expires_at"`
}

// GenericResponse - Wraps all API responses
type GenericResponse struct {
    Data  interface{}   `json:"data"`
    Error interface{}   `json:"error,omitempty"`
}

// ErrorResponse - Error details in response
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### Location: `internal/users/http/dto.go`

```go
// UpdateMeRequest - PUT /api/v1/users/me
type UpdateMeRequest struct {
    Timezone string `json:"timezone" validate:"required"`
}
```

## Auth Handlers

### RegisterHandler

**Location:** `internal/auth/http/register_handler.go`

**Endpoint:** `POST /api/v1/auth/register`

**Purpose:** Register a new user

**Request:**
```json
{
  "username": "player1",
  "email": "player@example.com",
  "password": "securepassword123",
  "timezone": "Europe/Madrid"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": "user-uuid",
    "username": "player1",
    "email": "player@example.com",
    "timezone": "Europe/Madrid",
    "created_at": "2026-05-24T12:34:56Z"
  }
}
```

**Implementation:**

```go
type RegisterHandler struct {
    userService service.UserService
    validator   *validator.Validate
}

func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Parse RegisterRequest
    // 2. Validate with validator.Struct()
    // 3. Call userService.Register(ctx, username, email, password, timezone)
    // 4. Handle errors:
    //    - "username already exists" → 409 CONFLICT, code: USERNAME_ALREADY_EXISTS
    //    - "email already exists" → 409 CONFLICT, code: EMAIL_ALREADY_EXISTS
    //    - Other errors → 500 INTERNAL_SERVER_ERROR
    // 5. Return UserDTO wrapped in GenericResponse
    // 6. Log: Info on success, Error on failure
}
```

**Error Codes:**
- `INVALID_INPUT` (400) - Validation failed
- `USERNAME_ALREADY_EXISTS` (409) - Username taken
- `EMAIL_ALREADY_EXISTS` (409) - Email taken
- `INTERNAL_ERROR` (500) - Service error

### LoginHandler

**Location:** `internal/auth/http/login_handler.go`

**Endpoint:** `POST /api/v1/auth/login`

**Purpose:** Authenticate user and issue session/refresh tokens

**Request:**
```json
{
  "username": "player1",
  "password": "securepassword123"
}
```

**Response (201 Created, Sets Cookies):**
```json
{
  "data": {
    "user": {
      "id": "user-uuid",
      "username": "player1",
      "email": "player@example.com",
      "timezone": "Europe/Madrid",
      "created_at": "2026-05-24T12:34:56Z"
    },
    "session_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "session_expires_at": 1234567890
  }
}
```

**Cookies Set:**
```
Set-Cookie: session_token=eyJhbGc...; HttpOnly; Secure; SameSite=Strict; Max-Age=60; Path=/
Set-Cookie: refresh_token=eyJhbGc...; HttpOnly; Secure; SameSite=Strict; Max-Age=25920000; Path=/
```

**Implementation:**

```go
type LoginHandler struct {
    authService service.AuthService
    validator   *validator.Validate
}

func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Parse LoginRequest
    // 2. Validate
    // 3. Call authService.Login(ctx, username, password)
    // 4. Handle errors:
    //    - Invalid credentials → 401 UNAUTHORIZED, code: INVALID_CREDENTIALS
    //    - Other errors → 500 INTERNAL_SERVER_ERROR
    // 5. Set two httpOnly cookies (session_token, refresh_token)
    // 6. Return LoginResponse with user DTO + tokens + expiration
    // 7. Log: Info on success, Debug on failed credentials
}
```

**Cookies:**
- `session_token` - 1 minute TTL, httpOnly, Secure, SameSite=Strict
- `refresh_token` - 300 days TTL, httpOnly, Secure, SameSite=Strict

**Error Codes:**
- `INVALID_INPUT` (400) - Validation failed
- `INVALID_CREDENTIALS` (401) - Wrong username/password
- `INTERNAL_ERROR` (500) - Service error

### LogoutHandler

**Location:** `internal/auth/http/logout_handler.go`

**Endpoint:** `POST /api/v1/auth/logout`

**Purpose:** Revoke user session and clear cookies

**Request:**
```json
{}
```

**Response (200 OK, Clears Cookies):**
```json
{
  "data": "logged out"
}
```

**Cookies Cleared:**
```
Set-Cookie: session_token=; HttpOnly; Secure; SameSite=Strict; Max-Age=-1; Path=/
Set-Cookie: refresh_token=; HttpOnly; Secure; SameSite=Strict; Max-Age=-1; Path=/
```

**Implementation:**

```go
type LogoutHandler struct {
    authService service.AuthService
}

func (h *LogoutHandler) Handle(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. Extract session_id from context (set by auth middleware)
    // 2. Verify session exists
    // 3. Call authService.Logout(ctx, sessionID)
    // 4. Clear both cookies (MaxAge=-1)
    // 5. Return success message
    // 6. Log: Info on success, Error on failure
}
```

**Error Codes:**
- `UNAUTHORIZED` (401) - No valid session in context
- `INTERNAL_ERROR` (500) - Service error

## User Handlers

### GetMeHandler

**Location:** `internal/users/http/get_me_handler.go`

**Endpoint:** `GET /api/v1/users/me`

**Purpose:** Get authenticated user profile

**Request:** None (requires valid session token)

**Response (200 OK):**
```json
{
  "data": {
    "id": "user-uuid",
    "username": "player1",
    "email": "player@example.com",
    "timezone": "Europe/Madrid",
    "created_at": "2026-05-24T12:34:56Z"
  }
}
```

**Implementation:**

```go
type GetMeHandler struct {
    userService service.UserService
}

func (h *GetMeHandler) Handle(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. Extract user_id from context (set by auth middleware)
    // 2. Call userService.GetUser(ctx, userID)
    // 3. Handle errors:
    //    - No user_id in context → 401 UNAUTHORIZED
    //    - User not found → 404 NOT_FOUND
    //    - Other errors → 500 INTERNAL_SERVER_ERROR
    // 4. Return UserDTO
    // 5. Handle *string Timezone by dereferencing safely
}
```

**Error Codes:**
- `UNAUTHORIZED` (401) - No session
- `USER_NOT_FOUND` (404) - User doesn't exist
- `INTERNAL_ERROR` (500) - Service error

### UpdateMeHandler

**Location:** `internal/users/http/update_me_handler.go`

**Endpoint:** `PUT /api/v1/users/me`

**Purpose:** Update user profile (timezone)

**Request:**
```json
{
  "timezone": "Europe/Paris"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": "user-uuid",
    "username": "player1",
    "email": "player@example.com",
    "timezone": "Europe/Paris",
    "created_at": "2026-05-24T12:34:56Z"
  }
}
```

**Implementation:**

```go
type UpdateMeHandler struct {
    userService service.UserService
    validator   *validator.Validate
}

func (h *UpdateMeHandler) Handle(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. Extract user_id from context
    // 2. Parse UpdateMeRequest
    // 3. Validate timezone field
    // 4. Create UserUpdateRequest with timezone pointer
    // 5. Call userService.UpdateUser(ctx, userID, updateReq)
    // 6. Handle errors:
    //    - "timezone_recently_changed" → 400 BAD_REQUEST, code: TIMEZONE_RECENTLY_CHANGED
    //    - Other errors → 500 INTERNAL_SERVER_ERROR
    // 7. Return updated UserDTO
    // 8. Log: Info on success, Warn on timezone conflict
}
```

**Error Codes:**
- `INVALID_INPUT` (400) - Validation failed
- `UNAUTHORIZED` (401) - No session
- `TIMEZONE_RECENTLY_CHANGED` (400) - Changed timezone within 30 days
- `INTERNAL_ERROR` (500) - Service error

## Route Registration

### AuthRoutes

**Location:** `internal/auth/http/routes.go`

```go
// RegisterAuthRoutes registers all auth-related routes
func RegisterAuthRoutes(
    router chi.Router,
    userService service.UserService,
    authService service.AuthService,
    validator *validator.Validate,
) {
    registerHandler := NewRegisterHandler(userService, validator)
    loginHandler := NewLoginHandler(authService, validator)
    logoutHandler := NewLogoutHandler(authService)
    
    router.Post("/api/v1/auth/register", registerHandler.Handle)
    router.Post("/api/v1/auth/login", loginHandler.Handle)
    router.Post("/api/v1/auth/logout", logoutHandler.Handle)
}
```

### UsersRoutes

**Location:** `internal/users/http/routes.go`

```go
// RegisterUsersRoutes registers all user-related routes
func RegisterUsersRoutes(
    router chi.Router,
    userService service.UserService,
    validator *validator.Validate,
) {
    getMeHandler := NewGetMeHandler(userService)
    updateMeHandler := NewUpdateMeHandler(userService, validator)
    
    router.Get("/api/v1/users/me", getMeHandler.Handle)
    router.Put("/api/v1/users/me", updateMeHandler.Handle)
}
```

## Dependency Injection

### In main.go (Phase 1.11)

```go
// Setup validator
validate := validator.New()

// Setup services (from Phase 1.7)
userService := service.NewUserService(userRepo, passwordManager)
authService := service.NewAuthService(
    userRepo,
    sessionRepo,
    passwordManager,
    jwtManager,
    sessionCache,
)

// Register routes
auth_http.RegisterAuthRoutes(router, userService, authService, validate)
users_http.RegisterUsersRoutes(router, userService, validate)
```

## Middleware Integration

### Auth Middleware (Phase 1.9)

GetMe and UpdateMe handlers require auth middleware to extract `user_id` from JWT:

```go
// In auth middleware:
sessionToken := r.Header.Get("Authorization") // or from cookie
claims := validateSessionToken(sessionToken)
ctx = context.WithValue(ctx, "user_id", claims.UserID)
ctx = context.WithValue(ctx, "session_id", claims.SessionID)
```

Logout handler extracts `session_id` from context to revoke.

## Error Handling Patterns

### Validation Errors

```go
if err := validator.Struct(request); err != nil {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(GenericResponse{
        Error: ErrorResponse{
            Code:    "INVALID_INPUT",
            Message: "Validation failed",
        },
    })
    return
}
```

### Service Errors

```go
user, err := service.Register(...)
if err != nil {
    if err.Error() == "username already exists" {
        w.WriteHeader(http.StatusConflict)
        json.NewEncoder(w).Encode(GenericResponse{
            Error: ErrorResponse{
                Code:    "USERNAME_ALREADY_EXISTS",
                Message: "Username is already in use",
            },
        })
        return
    }
    // Generic error handling
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(GenericResponse{
        Error: ErrorResponse{
            Code:    "INTERNAL_ERROR",
            Message: "Operation failed",
        },
    })
}
```

## Type Pointer Handling

**Important:** User model has `Timezone *string`. Handlers must dereference safely:

```go
// Building UserDTO from service User
timezone := ""
if user.Timezone != nil {
    timezone = *user.Timezone
}
userDTO := &UserDTO{
    Timezone: timezone,
    // ... other fields
}
```

## Logging Levels

- **Info**: Successful operations (register, login, logout, update)
- **Debug**: Expected failures (not found, validation failed, wrong credentials)
- **Error**: Unexpected failures (database errors, service errors)
- **Warn**: Unusual but handled situations (timezone conflict, rate limiting)

```go
// Success
log.Info().Str("user_id", user.ID).Msg("User successfully registered")

// Debug
log.Debug().Str("username", username).Msg("User not found during login")

// Error
log.Error().Err(err).Str("user_id", userID).Msg("Failed to update user")
```

## Testing Considerations

### Unit Tests (Phase 1.10)

Test each handler with:
- Valid input → Success response
- Invalid input → 400 with INVALID_INPUT
- Service errors → 500 with INTERNAL_ERROR
- Service-specific errors (username exists) → Correct HTTP status and code
- Context extraction failures → 401 UNAUTHORIZED

Mock dependencies:
- UserService
- AuthService
- Validator

### Integration Tests (Phase 1.10)

Test complete flows:
- Register → Login → GetMe → UpdateMe → Logout
- Register with duplicate username
- Login with wrong credentials
- GetMe without auth

## Implementation Checklist

- [x] DTOs defined in auth/http/dto.go and users/http/dto.go
- [x] RegisterHandler created with validation and error handling
- [x] LoginHandler created with cookie setting and token generation
- [x] LogoutHandler created with session revocation and cookie clearing
- [x] GetMeHandler created with context extraction
- [x] UpdateMeHandler created with timezone update logic
- [x] routes.go for auth endpoints
- [x] routes.go for user endpoints
- [x] All handlers compile without errors
- [ ] HTTP middleware with auth support (Phase 1.9)
- [ ] Unit tests for all handlers (Phase 1.10)
- [ ] Integration in main.go (Phase 1.11)

## Related Phases

- **Phase 1.7**: Auth Service and Session Cache (provides Login/Logout logic)
- **Phase 1.6**: User Service (provides Register/GetUser/UpdateUser)
- **Phase 1.9**: Auth Middleware (validates tokens, extracts user_id)
- **Phase 1.10**: Unit Tests (tests all handler paths)
- **Phase 1.11**: Main Integration (wires everything together)

## HTTP Status Codes Used

| Status | Code | Meaning |
|--------|------|---------|
| 200 | OK | GetMe, Logout, UpdateMe success |
| 201 | Created | Register, Login success |
| 400 | Bad Request | Validation failed, TIMEZONE_RECENTLY_CHANGED |
| 401 | Unauthorized | Invalid credentials, no session, invalid token |
| 404 | Not Found | User not found |
| 409 | Conflict | Username/Email already exists |
| 500 | Internal Server Error | Service failure |

---

**Documentation Version:** 1.0  
**Last Updated:** Phase 1.8 Completion  
**Author:** Development Team
