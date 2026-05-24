-- internal/auth/queries.sql - Operaciones CRUD para tabla user_sessions

-- name: CreateSession :exec
INSERT INTO user_sessions (id, user_id, session_id, verification_string, token_hash, expires_at, created_at, is_deleted, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 0, NULL);

-- name: GetSessionByID :one
SELECT id, user_id, session_id, verification_string, token_hash, expires_at, revoked_at, created_at, is_deleted, deleted_at
FROM user_sessions
WHERE id = ? AND is_deleted = 0;

-- name: GetSessionBySessionID :one
SELECT id, user_id, session_id, verification_string, token_hash, expires_at, revoked_at, created_at, is_deleted, deleted_at
FROM user_sessions
WHERE session_id = ? AND is_deleted = 0;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, session_id, verification_string, token_hash, expires_at, revoked_at, created_at, is_deleted, deleted_at
FROM user_sessions
WHERE token_hash = ? AND is_deleted = 0;

-- name: RevokeSession :exec
UPDATE user_sessions
SET revoked_at = CURRENT_TIMESTAMP
WHERE session_id = ? AND is_deleted = 0;

-- name: DeleteExpiredSessions :exec
UPDATE user_sessions
SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
WHERE (expires_at < CURRENT_TIMESTAMP OR revoked_at IS NOT NULL) AND is_deleted = 0;

-- name: CountActiveSessionsByUserID :one
SELECT COUNT(*) as count
FROM user_sessions
WHERE user_id = ? AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP AND is_deleted = 0;

-- name: GetUserSessionsByUserID :many
SELECT id, user_id, session_id, verification_string, token_hash, expires_at, revoked_at, created_at, is_deleted, deleted_at
FROM user_sessions
WHERE user_id = ? AND is_deleted = 0;
