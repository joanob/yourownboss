package repository

import (
	"context"
	"database/sql"
	"time"

	"yourownboss/internal/db"
)

// UserRepository handles user data access
type UserRepository interface {
	Create(ctx context.Context, username, passwordHash string) (*db.User, error)
	GetByUsername(ctx context.Context, username string) (*db.User, error)
	GetByID(ctx context.Context, id int64) (*db.User, error)
}

type userRepository struct {
	db *db.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(database *db.DB) UserRepository {
	return &userRepository{db: database}
}

func (r *userRepository) Create(ctx context.Context, username, passwordHash string) (*db.User, error) {
	result, err := r.db.ExecContext(
		ctx,
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: users.username" {
			return nil, db.ErrUserAlreadyExists
		}
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &db.User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	var user db.User
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, db.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*db.User, error) {
	var user db.User
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, username, password_hash, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, db.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
