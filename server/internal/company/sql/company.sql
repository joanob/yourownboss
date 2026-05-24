-- name: CreateCompany :one
INSERT INTO companies (id, user_id, name, money, created_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING id, user_id, name, money, created_at, is_deleted, deleted_at;

-- name: GetCompanyByID :one
SELECT id, user_id, name, money, created_at, is_deleted, deleted_at
FROM companies
WHERE id = ? AND is_deleted = 0;

-- name: GetCompanyByUserID :one
SELECT id, user_id, name, money, created_at, is_deleted, deleted_at
FROM companies
WHERE user_id = ? AND is_deleted = 0;

-- name: UpdateCompanyName :one
UPDATE companies
SET name = ?
WHERE id = ? AND is_deleted = 0
RETURNING id, user_id, name, money, created_at, is_deleted, deleted_at;

-- name: UpdateCompanyMoney :one
UPDATE companies
SET money = ?
WHERE id = ? AND is_deleted = 0
RETURNING id, user_id, name, money, created_at, is_deleted, deleted_at;

-- name: CheckCompanyExists :one
SELECT COUNT(*) as count
FROM companies
WHERE user_id = ? AND is_deleted = 0;

-- name: SoftDeleteCompany :exec
UPDATE companies
SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
WHERE id = ?;
