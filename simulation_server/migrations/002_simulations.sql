-- migrations/002_simulations.sql

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS simulations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  process_id INTEGER NOT NULL,
  time_ms INTEGER NOT NULL,
  benefit_per_hour REAL NOT NULL,
  FOREIGN KEY (process_id) REFERENCES production_processes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS simulation_resources (
  simulation_id INTEGER NOT NULL,
  resource_id INTEGER NOT NULL,
  is_output INTEGER NOT NULL,
  price INTEGER NOT NULL,
  quantity INTEGER NOT NULL,
  PRIMARY KEY (simulation_id, resource_id, is_output),
  FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE,
  FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_simulation_resources_simulation_id ON simulation_resources(simulation_id);
