-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Refresh tokens table (for session management and revocation)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    revoked_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Companies table
CREATE TABLE IF NOT EXISTS companies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    name TEXT NOT NULL,
    money INTEGER NOT NULL DEFAULT 50000000, -- Stored in thousandths (50,000.000 = 50000000)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_companies_user_id ON companies(user_id);
-- Resources table (available resources in the game)
CREATE TABLE IF NOT EXISTS resources (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    price INTEGER NOT NULL, -- Price in thousandths for pack_size units
    pack_size INTEGER NOT NULL DEFAULT 1 -- Number of units per pack
);

-- Production buildings table (available production buildings in the game)
CREATE TABLE IF NOT EXISTS production_buildings (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    cost INTEGER NOT NULL
);

-- Production processes table (available production processes in the game)
CREATE TABLE IF NOT EXISTS production_processes (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    processing_time_ms INTEGER NOT NULL,
    building_id INTEGER NOT NULL,
    window_start_hour INTEGER,
    window_end_hour INTEGER,
    FOREIGN KEY (building_id) REFERENCES production_buildings(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_production_processes_building_id ON production_processes(building_id);

-- Company inventory table (resources owned by companies)
CREATE TABLE IF NOT EXISTS company_inventory (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id INTEGER NOT NULL,
    resource_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, resource_id),
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_company_inventory_company_id ON company_inventory(company_id);
CREATE INDEX IF NOT EXISTS idx_company_inventory_resource_id ON company_inventory(resource_id);