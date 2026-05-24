-- Resources queries
-- name: CreateResource :exec
INSERT INTO resources (id, master_id, name, market_price, market_sale_qty)
VALUES (?, ?, ?, ?, ?);

-- name: GetResourceByID :one
SELECT id, master_id, name, market_price, market_sale_qty
FROM resources
WHERE id = ?;

-- name: GetResourceByMasterID :one
SELECT id, master_id, name, market_price, market_sale_qty
FROM resources
WHERE master_id = ?;

-- name: GetAllResources :many
SELECT id, master_id, name, market_price, market_sale_qty
FROM resources
ORDER BY name;

-- name: UpsertResource :exec
INSERT INTO resources (id, master_id, name, market_price, market_sale_qty)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(master_id) DO UPDATE SET
  id = excluded.id,
  name = excluded.name,
  market_price = excluded.market_price,
  market_sale_qty = excluded.market_sale_qty;
