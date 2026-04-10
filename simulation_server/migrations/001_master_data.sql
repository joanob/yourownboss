-- migrations/001_master_data.sql

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS resources (
  id INTEGER NOT NULL PRIMARY KEY,
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS production_buildings (
  id INTEGER NOT NULL PRIMARY KEY,
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS production_processes (
  id INTEGER NOT NULL PRIMARY KEY,
  building_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  start_hour INTEGER,
  end_hour INTEGER,
  FOREIGN KEY (building_id) REFERENCES production_buildings(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS process_resources (
  process_id INTEGER NOT NULL,
  resource_id INTEGER NOT NULL,
  is_output INTEGER NOT NULL,
  PRIMARY KEY (process_id, resource_id, is_output),
  FOREIGN KEY (process_id) REFERENCES production_processes(id) ON DELETE CASCADE,
  FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE RESTRICT
);
