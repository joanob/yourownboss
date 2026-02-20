package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"yourownboss/internal/db"
)

// TokenRepository handles refresh token data access
type TokenRepository interface {
	Save(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	Validate(ctx context.Context, token string) (int64, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllForUser(ctx context.Context, userID int64) error
	CleanupExpired(ctx context.Context) error
}

type tokenRepository struct {
	db *db.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(database *db.DB) TokenRepository {
	return &tokenRepository{db: database}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (r *tokenRepository) Save(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	tokenHash := hashToken(token)
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)",
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *tokenRepository) Validate(ctx context.Context, token string) (int64, error) {
	tokenHash := hashToken(token)

	var userID int64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT user_id FROM refresh_tokens 
		 WHERE token_hash = ? AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL`,
		tokenHash,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		return 0, errors.New("invalid or expired refresh token")
	}
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *tokenRepository) Revoke(ctx context.Context, token string) error {
	tokenHash := hashToken(token)
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE token_hash = ?",
		tokenHash,
	)
	return err
}

func (r *tokenRepository) RevokeAllForUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = ? AND revoked_at IS NULL",
		userID,
	)
	return err
}

func (r *tokenRepository) CleanupExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(
		ctx,
		"DELETE FROM refresh_tokens WHERE expires_at < CURRENT_TIMESTAMP",
	)
	return err
}
