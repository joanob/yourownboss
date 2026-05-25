package http

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Timezone string `json:"timezone" validate:"required"`
}

// LoginRequest is the request body for user login.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UserDTO represents a user in API responses.
type UserDTO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Timezone  string `json:"timezone"`
	CreatedAt string `json:"created_at"`
}

// LoginResponse is the response body for successful login.
// SEC-02: tokens are delivered only via httpOnly cookies, NOT in the body.
type LoginResponse struct {
	User             *UserDTO `json:"user"`
	SessionExpiresAt int64    `json:"session_expires_at"`
}

// GenericResponse wraps the response data.
type GenericResponse struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error,omitempty"`
}

// ErrorResponse represents an error in API responses.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
