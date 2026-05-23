-- Your Own Boss - Database Schema
-- Base de datos SQLite
-- Esto contiene todas las tablas necesarias para el juego

-- DATOS MAESTROS

CREATE TABLE IF NOT EXISTS resources (
  id              TEXT PRIMARY KEY,
  master_id       TEXT    NOT NULL UNIQUE,
  name            TEXT    NOT NULL,
  market_price    INTEGER NOT NULL,
  market_sale_qty INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS production_buildings (
  id                  TEXT PRIMARY KEY,
  master_id           TEXT    NOT NULL UNIQUE,
  name                TEXT    NOT NULL,
  construction_cost   INTEGER NOT NULL,
  construction_time_s INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS production_processes (
  id                     TEXT PRIMARY KEY,
  master_id              TEXT    NOT NULL UNIQUE,
  production_building_id TEXT    NOT NULL REFERENCES production_buildings(id),
  name                   TEXT    NOT NULL,
  cycle_time_s           INTEGER NOT NULL,
  window_start_hour      INTEGER,
  window_end_hour        INTEGER
);

CREATE TABLE IF NOT EXISTS production_process_resources (
  process_id  TEXT    NOT NULL REFERENCES production_processes(id),
  resource_id TEXT    NOT NULL REFERENCES resources(id),
  is_output   BOOLEAN NOT NULL,
  quantity    INTEGER NOT NULL,
  PRIMARY KEY (process_id, resource_id, is_output)
);

CREATE TABLE IF NOT EXISTS sale_buildings (
  id                  TEXT PRIMARY KEY,
  master_id           TEXT    NOT NULL UNIQUE,
  name                TEXT    NOT NULL,
  construction_cost   INTEGER NOT NULL,
  construction_time_s INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sale_resources (
  sale_building_id        TEXT    NOT NULL REFERENCES sale_buildings(id),
  resource_id             TEXT    NOT NULL REFERENCES resources(id),
  price_per_unit          INTEGER NOT NULL,
  units_sold_per_second   INTEGER NOT NULL,
  PRIMARY KEY (sale_building_id, resource_id)
);

-- USUARIOS Y AUTENTICACIÓN

CREATE TABLE IF NOT EXISTS users (
  id                              TEXT     PRIMARY KEY,
  username                        TEXT     NOT NULL UNIQUE,
  email                           TEXT     NOT NULL UNIQUE,
  password_hash                   TEXT     NOT NULL,
  role                            TEXT     NOT NULL DEFAULT 'P',
  timezone                        TEXT,
  last_timezone_modification_at   DATETIME,
  created_at                      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted                      INTEGER  NOT NULL DEFAULT 0,
  deleted_at                      DATETIME
);

CREATE TABLE IF NOT EXISTS user_sessions (
  id                  TEXT     PRIMARY KEY,
  user_id             TEXT     NOT NULL REFERENCES users(id),
  session_id          TEXT     NOT NULL UNIQUE,
  verification_string TEXT     NOT NULL,
  token_hash          TEXT     NOT NULL UNIQUE,
  expires_at          DATETIME NOT NULL,
  revoked_at          DATETIME,
  created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted          INTEGER  NOT NULL DEFAULT 0,
  deleted_at          DATETIME
);

-- EMPRESAS

CREATE TABLE IF NOT EXISTS companies (
  id         TEXT     PRIMARY KEY,
  user_id    TEXT     NOT NULL UNIQUE REFERENCES users(id),
  name       TEXT     NOT NULL,
  money      INTEGER  NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted INTEGER  NOT NULL DEFAULT 0,
  deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS company_inventory (
  id          TEXT    PRIMARY KEY,
  company_id  TEXT    NOT NULL REFERENCES companies(id),
  resource_id TEXT    NOT NULL REFERENCES resources(id),
  quantity    INTEGER NOT NULL DEFAULT 0,
  is_deleted  INTEGER NOT NULL DEFAULT 0,
  deleted_at  DATETIME,
  UNIQUE (company_id, resource_id)
);

-- PRODUCCIÓN

CREATE TABLE IF NOT EXISTS company_production_buildings (
  id                     TEXT     PRIMARY KEY,
  company_id             TEXT     NOT NULL REFERENCES companies(id),
  production_building_id TEXT     NOT NULL REFERENCES production_buildings(id),
  level                  INTEGER  NOT NULL DEFAULT 1,
  construction_ends_at   DATETIME,
  created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted             INTEGER  NOT NULL DEFAULT 0,
  deleted_at             DATETIME
);

CREATE TABLE IF NOT EXISTS production_runs (
  id                 TEXT     PRIMARY KEY,
  company_building_id TEXT    NOT NULL REFERENCES company_production_buildings(id),
  process_id         TEXT     NOT NULL REFERENCES production_processes(id),
  production_cycles  INTEGER  NOT NULL,
  started_at         DATETIME NOT NULL,
  ends_at            DATETIME NOT NULL,
  is_collected       INTEGER  NOT NULL DEFAULT 0,
  collected_at       DATETIME,
  is_deleted         INTEGER  NOT NULL DEFAULT 0,
  deleted_at         DATETIME,
  UNIQUE (company_building_id, is_collected) WHERE is_collected = 0
);

-- VENTA

CREATE TABLE IF NOT EXISTS company_sale_buildings (
  id                 TEXT     PRIMARY KEY,
  company_id         TEXT     NOT NULL REFERENCES companies(id),
  sale_building_id   TEXT     NOT NULL REFERENCES sale_buildings(id),
  level              INTEGER  NOT NULL DEFAULT 1,
  construction_ends_at DATETIME,
  created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted         INTEGER  NOT NULL DEFAULT 0,
  deleted_at         DATETIME
);

CREATE TABLE IF NOT EXISTS sale_runs (
  id                        TEXT     PRIMARY KEY,
  company_sale_building_id  TEXT     NOT NULL REFERENCES company_sale_buildings(id),
  resource_id               TEXT     NOT NULL REFERENCES resources(id),
  units_to_sell             INTEGER  NOT NULL,
  started_at                DATETIME NOT NULL,
  ends_at                   DATETIME NOT NULL,
  is_collected              INTEGER  NOT NULL DEFAULT 0,
  collected_at              DATETIME,
  is_deleted                INTEGER  NOT NULL DEFAULT 0,
  deleted_at                DATETIME,
  UNIQUE (company_sale_building_id, is_collected) WHERE is_collected = 0
);

-- RATE LIMITING

CREATE TABLE IF NOT EXISTS login_attempts (
  id             TEXT     PRIMARY KEY,
  user_id        TEXT,
  username       TEXT     NOT NULL,
  failed_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted     INTEGER  NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts(username, failed_at);

-- AUDITORÍA

CREATE TABLE IF NOT EXISTS audit_log (
  id             TEXT     PRIMARY KEY,
  user_id        TEXT     REFERENCES users(id) ON DELETE SET NULL,
  company_id     TEXT     REFERENCES companies(id) ON DELETE SET NULL,
  action         TEXT     NOT NULL,
  resource_type  TEXT,
  resource_id    TEXT,
  changes        TEXT,
  timestamp      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ÍNDICES PARA OPTIMIZACIÓN DE QUERIES

CREATE INDEX IF NOT EXISTS idx_companies_user_id ON companies(user_id);
CREATE INDEX IF NOT EXISTS idx_production_buildings_company ON company_production_buildings(company_id);
CREATE INDEX IF NOT EXISTS idx_sale_buildings_company ON company_sale_buildings(company_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id, revoked_at);
CREATE INDEX IF NOT EXISTS idx_audit_log_company ON audit_log(company_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user_id, timestamp);

-- TRIGGERS PARA SOFT-DELETE

-- Trigger para usuarios
CREATE TRIGGER IF NOT EXISTS users_before_delete
BEFORE DELETE ON users
FOR EACH ROW
BEGIN
  UPDATE users
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  UPDATE user_sessions
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE user_id = OLD.id;

  UPDATE companies
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE user_id = OLD.id;

  UPDATE company_inventory
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);

  UPDATE company_production_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);

  UPDATE company_sale_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);

  UPDATE production_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_building_id IN (
    SELECT id FROM company_production_buildings
    WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id)
  );

  UPDATE sale_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_sale_building_id IN (
    SELECT id FROM company_sale_buildings
    WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id)
  );

  SELECT RAISE(IGNORE);
END;

-- Trigger para empresas
CREATE TRIGGER IF NOT EXISTS companies_before_delete
BEFORE DELETE ON companies
FOR EACH ROW
BEGIN
  UPDATE companies
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  UPDATE company_inventory
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id = OLD.id;

  UPDATE company_production_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id = OLD.id;

  UPDATE company_sale_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id = OLD.id;

  UPDATE production_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_building_id IN (
    SELECT id FROM company_production_buildings WHERE company_id = OLD.id
  );

  UPDATE sale_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_sale_building_id IN (
    SELECT id FROM company_sale_buildings WHERE company_id = OLD.id
  );

  SELECT RAISE(IGNORE);
END;

-- Trigger para edificios de producción
CREATE TRIGGER IF NOT EXISTS company_production_buildings_before_delete
BEFORE DELETE ON company_production_buildings
FOR EACH ROW
BEGIN
  UPDATE company_production_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  UPDATE production_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_building_id = OLD.id;

  SELECT RAISE(IGNORE);
END;

-- Trigger para edificios de venta
CREATE TRIGGER IF NOT EXISTS company_sale_buildings_before_delete
BEFORE DELETE ON company_sale_buildings
FOR EACH ROW
BEGIN
  UPDATE company_sale_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  UPDATE sale_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_sale_building_id = OLD.id;

  SELECT RAISE(IGNORE);
END;
