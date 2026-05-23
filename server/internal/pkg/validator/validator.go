package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wrapper sobre go-playground/validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator crea una nueva instancia del validador
func NewValidator() *CustomValidator {
	v := validator.New()

	// Registrar validadores personalizados
	v.RegisterValidation("multiple_of", validateMultipleOf)
	v.RegisterValidation("positive", validatePositive)

	return &CustomValidator{
		validator: v,
	}
}

// Validate valida un struct
func (cv *CustomValidator) Validate(data interface{}) error {
	return cv.validator.Struct(data)
}

// ValidateVar valida una variable individual
func (cv *CustomValidator) ValidateVar(data interface{}, tag string) error {
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
