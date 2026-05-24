-- Sale Buildings queries
-- name: CreateSaleBuilding :exec
INSERT INTO sale_buildings (id, master_id, name, construction_cost, construction_time_s)
VALUES (?, ?, ?, ?, ?);

-- name: GetSaleBuildingByID :one
SELECT id, master_id, name, construction_cost, construction_time_s
FROM sale_buildings
WHERE id = ?;

-- name: GetSaleBuildingByMasterID :one
SELECT id, master_id, name, construction_cost, construction_time_s
FROM sale_buildings
WHERE master_id = ?;

-- name: GetAllSaleBuildings :many
SELECT id, master_id, name, construction_cost, construction_time_s
FROM sale_buildings
ORDER BY name;

-- name: UpsertSaleBuilding :exec
INSERT INTO sale_buildings (id, master_id, name, construction_cost, construction_time_s)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(master_id) DO UPDATE SET
  id = excluded.id,
  name = excluded.name,
  construction_cost = excluded.construction_cost,
  construction_time_s = excluded.construction_time_s;

-- Sale Resources queries
-- name: CreateSaleResource :exec
INSERT INTO sale_resources (sale_building_id, resource_id, price_per_unit, units_sold_per_second)
VALUES (?, ?, ?, ?);

-- name: GetSaleResourcesByBuildingID :many
SELECT sale_building_id, resource_id, price_per_unit, units_sold_per_second
FROM sale_resources
WHERE sale_building_id = ?
ORDER BY resource_id;

-- name: DeleteSaleResources :exec
DELETE FROM sale_resources
WHERE sale_building_id = ?;
