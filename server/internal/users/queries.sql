-- queries.sql - Operaciones CRUD para tabla users

-- name: CreateUser :exec
INSERT INTO users (id, username, email, password_hash, role, timezone, created_at, is_deleted, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 0, NULL);

-- GetUserByID obtiene un usuario por ID (no soft-deleted)
-- name: GetUserByID :one
SELECT id, username, email, password_hash, role, timezone, last_timezone_modification_at, created_at, is_deleted, deleted_at
FROM users
WHERE id = ? AND is_deleted = 0;

-- GetUserByUsername obtiene un usuario por username (no soft-deleted)
-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, role, timezone, last_timezone_modification_at, created_at, is_deleted, deleted_at
FROM users
WHERE username = ? AND is_deleted = 0;

-- GetUserByEmail obtiene un usuario por email (no soft-deleted)
-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, role, timezone, last_timezone_modification_at, created_at, is_deleted, deleted_at
FROM users
WHERE email = ? AND is_deleted = 0;

-- UpdateUser actualiza datos del usuario
-- name: UpdateUser :exec
UPDATE users
SET timezone = ?, last_timezone_modification_at = ?
WHERE id = ? AND is_deleted = 0;

-- SoftDeleteUser marca usuario como borrado (soft delete)
-- name: SoftDeleteUser :exec
UPDATE users
SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
WHERE id = ? AND is_deleted = 0;

-- CountUserByUsername verifica si username ya existe (no soft-deleted)
-- name: CountUserByUsername :one
SELECT COUNT(*) as count
FROM users
WHERE username = ? AND is_deleted = 0;

-- CountUserByEmail verifica si email ya existe (no soft-deleted)
-- name: CountUserByEmail :one
SELECT COUNT(*) as count
FROM users
WHERE email = ? AND is_deleted = 0;
