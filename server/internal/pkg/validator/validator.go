package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validator envuelve go-playground/validator y permite omitir la validación
// cuando skip es true (p. ej. en el entorno de desarrollo).
type Validator struct {
	validator *validator.Validate
	skip      bool
}

// New crea un validador. Si skip es true, Struct/Validate/ValidateVar no aplican
// ninguna comprobación y siempre devuelven nil.
func New(skip bool) *Validator {
	v := validator.New()

	// Registrar validadores personalizados
	v.RegisterValidation("multiple_of", validateMultipleOf)
	v.RegisterValidation("positive", validatePositive)

	return &Validator{
		validator: v,
		skip:      skip,
	}
}

// NewValidator crea un validador con la validación siempre activa.
// Conservado por compatibilidad; prefiere New(skip).
func NewValidator() *Validator {
	return New(false)
}

// Struct valida un struct. Devuelve nil si la validación está desactivada.
func (cv *Validator) Struct(data interface{}) error {
	if cv.skip {
		return nil
	}
	return cv.validator.Struct(data)
}

// Validate valida un struct (alias de Struct).
func (cv *Validator) Validate(data interface{}) error {
	return cv.Struct(data)
}

// ValidateVar valida una variable individual. Devuelve nil si está desactivada.
func (cv *Validator) ValidateVar(data interface{}, tag string) error {
	if cv.skip {
		return nil
	}
	return cv.validator.Var(data, tag)
}

// validateMultipleOf valida que un número es múltiplo de otro
// Uso: `validate:"multiple_of=3"`
func validateMultipleOf(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()
	param := fl.Param()

	fieldInt, ok := field.(int)
	if !ok {
		fieldInt64, ok := field.(int64)
		if !ok {
			return false
		}
		fieldInt = int(fieldInt64)
	}

	paramInt := 0
	_, err := fmt.Sscanf(param, "%d", &paramInt)
	if err != nil || paramInt == 0 {
		return false
	}

	return fieldInt%paramInt == 0
}

// validatePositive valida que un número es positivo (> 0)
// Uso: `validate:"positive"`
func validatePositive(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()

	switch v := field.(type) {
	case int:
		return v > 0
	case int64:
		return v > 0
	case float64:
		return v > 0
	default:
		return false
	}
}
