package errors

import "fmt"

// ErrorCode representa un código de error estándar
type ErrorCode string

// Códigos de error 400 (Bad Request)
const (
	InvalidInputError       ErrorCode = "INVALID_INPUT"
	InvalidQuantityMultiple ErrorCode = "INVALID_QUANTITY_MULTIPLE"
	InvalidTimezone         ErrorCode = "INVALID_TIMEZONE"
	TimezoneRecentlyChanged ErrorCode = "TIMEZONE_RECENTLY_CHANGED"
	ProcessNotAvailableNow  ErrorCode = "PROCESS_NOT_AVAILABLE_NOW"
)

// Códigos de error 401 (Unauthorized)
const (
	UnauthorizedError  ErrorCode = "UNAUTHORIZED"
	InvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	SessionExpired     ErrorCode = "SESSION_EXPIRED"
)

// Códigos de error 403 (Forbidden)
const (
	InsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
	CompanyNotFoundError    ErrorCode = "COMPANY_NOT_FOUND"
)

// Códigos de error 404 (Not Found)
const (
	UserNotFoundError     ErrorCode = "USER_NOT_FOUND"
	ResourceNotFoundError ErrorCode = "RESOURCE_NOT_FOUND"
	BuildingNotFoundError ErrorCode = "BUILDING_NOT_FOUND"
	ProcessNotFoundError  ErrorCode = "PROCESS_NOT_FOUND"
)

// Códigos de error 409 (Conflict)
const (
	InsufficientFunds         ErrorCode = "INSUFFICIENT_FUNDS"
	InsufficientInventory     ErrorCode = "INSUFFICIENT_INVENTORY"
	BuildingNotIdleError      ErrorCode = "BUILDING_NOT_IDLE"
	BuildingNotProducingError ErrorCode = "BUILDING_NOT_PRODUCING"
	CompanyAlreadyExists      ErrorCode = "COMPANY_ALREADY_EXISTS"
	UsernameAlreadyExists     ErrorCode = "USERNAME_ALREADY_EXISTS"
	EmailAlreadyExists        ErrorCode = "EMAIL_ALREADY_EXISTS"
)

// AppError representa un error de aplicación con código y mensaje
type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// NewAppError crea un nuevo error de aplicación
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   nil,
	}
}

// NewAppErrorWithCause crea un nuevo error con causa
func NewAppErrorWithCause(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Error implementa la interfaz error
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap devuelve el error subyacente
func (e *AppError) Unwrap() error {
	return e.Cause
}

// GetHTTPStatusCode devuelve el código HTTP correspondiente al código de error
func (e *AppError) GetHTTPStatusCode() int {
	switch e.Code {
	// 400 Bad Request
	case InvalidInputError, InvalidQuantityMultiple, InvalidTimezone, TimezoneRecentlyChanged, ProcessNotAvailableNow:
		return 400

	// 401 Unauthorized
	case UnauthorizedError, InvalidCredentials, SessionExpired:
		return 401

	// 403 Forbidden
	case InsufficientPermissions, CompanyNotFoundError:
		return 403

	// 404 Not Found
	case UserNotFoundError, ResourceNotFoundError, BuildingNotFoundError, ProcessNotFoundError:
		return 404

	// 409 Conflict
	case InsufficientFunds, InsufficientInventory, BuildingNotIdleError, BuildingNotProducingError,
		CompanyAlreadyExists, UsernameAlreadyExists, EmailAlreadyExists:
		return 409

	// Default
	default:
		return 500
	}
}

// ErrorResponse estructura para respuestas de error HTTP
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse crea una respuesta de error
func NewErrorResponse(appErr *AppError) ErrorResponse {
	return ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Message,
	}
}
