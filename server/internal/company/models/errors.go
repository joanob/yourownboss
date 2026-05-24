package models

import "errors"

// Error codes para dominio de Company
var (
	ErrInsufficientFunds     = errors.New("insufficient_funds")
	ErrInsufficientInventory = errors.New("insufficient_inventory")
	ErrInvalidAmount         = errors.New("invalid_amount")
	ErrInvalidQuantity       = errors.New("invalid_quantity")
	ErrCompanyNotFound       = errors.New("company_not_found")
	ErrCompanyAlreadyExists  = errors.New("company_already_exists")
	ErrCompanyDeleted        = errors.New("company_deleted")
)
