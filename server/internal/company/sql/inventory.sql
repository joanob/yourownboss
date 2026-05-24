-- name: GetInventoryByCompanyID :many
SELECT id, company_id, resource_id, quantity, is_deleted, deleted_at
FROM company_inventory
WHERE company_id = ? AND is_deleted = 0
ORDER BY resource_id;

-- name: GetInventoryItem :one
SELECT id, company_id, resource_id, quantity, is_deleted, deleted_at
FROM company_inventory
WHERE company_id = ? AND resource_id = ? AND is_deleted = 0;

-- name: CreateInventoryItem :one
INSERT INTO company_inventory (id, company_id, resource_id, quantity)
VALUES (?, ?, ?, ?)
RETURNING id, company_id, resource_id, quantity, is_deleted, deleted_at;

-- name: UpdateInventoryQuantity :one
UPDATE company_inventory
SET quantity = ?
WHERE id = ? AND is_deleted = 0
RETURNING id, company_id, resource_id, quantity, is_deleted, deleted_at;

-- name: AddToInventory :one
INSERT INTO company_inventory (id, company_id, resource_id, quantity)
VALUES (?, ?, ?, ?)
ON CONFLICT(company_id, resource_id) DO UPDATE SET
  quantity = quantity + excluded.quantity,
  is_deleted = 0,
  deleted_at = NULL
RETURNING id, company_id, resource_id, quantity, is_deleted, deleted_at;

-- name: UpsertInventoryItem :one
INSERT INTO company_inventory (id, company_id, resource_id, quantity)
VALUES (?, ?, ?, ?)
ON CONFLICT(company_id, resource_id) DO UPDATE SET
  quantity = excluded.quantity,
  is_deleted = 0,
  deleted_at = NULL
RETURNING id, company_id, resource_id, quantity, is_deleted, deleted_at;

-- name: RemoveFromInventory :one
UPDATE company_inventory
SET quantity = ?
WHERE company_id = ? AND resource_id = ? AND is_deleted = 0
RETURNING id, company_id, resource_id, quantity, is_deleted, deleted_at;
