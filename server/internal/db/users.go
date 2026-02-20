package db

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("username already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

// User represents a user in the database
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
