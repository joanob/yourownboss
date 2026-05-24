-- Production Buildings queries
-- name: CreateProductionBuilding :exec
INSERT INTO production_buildings (id, master_id, name, construction_cost, construction_time_s)
VALUES (?, ?, ?, ?, ?);

-- name: GetProductionBuildingByID :one
SELECT id, master_id, name, construction_cost, construction_time_s
FROM production_buildings
WHERE id = ?;

-- name: GetProductionBuildingByMasterID :one
SELECT id, master_id, name, construction_cost, construction_time_s
FROM production_buildings
WHERE master_id = ?;

-- name: GetAllProductionBuildings :many
SELECT id, master_id, name, construction_cost, construction_time_s
FROM production_buildings
ORDER BY name;

-- name: UpsertProductionBuilding :exec
INSERT INTO production_buildings (id, master_id, name, construction_cost, construction_time_s)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(master_id) DO UPDATE SET
  id = excluded.id,
  name = excluded.name,
  construction_cost = excluded.construction_cost,
  construction_time_s = excluded.construction_time_s;

-- Production Processes queries
-- name: CreateProductionProcess :exec
INSERT INTO production_processes (id, master_id, production_building_id, name, cycle_time_s, window_start_hour, window_end_hour)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetProductionProcessByID :one
SELECT id, master_id, production_building_id, name, cycle_time_s, window_start_hour, window_end_hour
FROM production_processes
WHERE id = ?;

-- name: GetProductionProcessByMasterID :one
SELECT id, master_id, production_building_id, name, cycle_time_s, window_start_hour, window_end_hour
FROM production_processes
WHERE master_id = ?;

-- name: GetProductionProcessesByBuildingID :many
SELECT id, master_id, production_building_id, name, cycle_time_s, window_start_hour, window_end_hour
FROM production_processes
WHERE production_building_id = ?
ORDER BY name;

-- name: UpsertProductionProcess :exec
INSERT INTO production_processes (id, master_id, production_building_id, name, cycle_time_s, window_start_hour, window_end_hour)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(master_id) DO UPDATE SET
  id = excluded.id,
  production_building_id = excluded.production_building_id,
  name = excluded.name,
  cycle_time_s = excluded.cycle_time_s,
  window_start_hour = excluded.window_start_hour,
  window_end_hour = excluded.window_end_hour;

-- Production Process Resources queries
-- name: CreateProductionProcessResource :exec
INSERT INTO production_process_resources (process_id, resource_id, is_output, quantity)
VALUES (?, ?, ?, ?);

-- name: GetProductionProcessResources :many
SELECT process_id, resource_id, is_output, quantity
FROM production_process_resources
WHERE process_id = ?
ORDER BY is_output DESC, resource_id;

-- name: DeleteProductionProcessResources :exec
DELETE FROM production_process_resources
WHERE process_id = ?;
