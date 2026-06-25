// ============================================================
// Base
// ============================================================

export interface ApiErrorBody {
  code: string;
  message: string;
}

// ============================================================
// Game / Status
// ============================================================

export interface StatusResponse {
  status: string;
  version: string;
  db_connected: boolean;
}

// ============================================================
// Master Data (Gamedata)
// ============================================================

export interface Resource {
  id: string;
  master_id: string;
  name: string;
  market_price: number;
  market_sale_qty: number;
}

export interface ProcessResource {
  resource_id: string;
  quantity: number;
}

export interface ProductionProcess {
  id: string;
  master_id: string;
  name: string;
  cycle_time_s: number;
  window_start_hour?: number;
  window_end_hour?: number;
  inputs: ProcessResource[];
  outputs: ProcessResource[];
}

export interface ProductionBuilding {
  id: string;
  master_id: string;
  name: string;
  construction_cost: number;
  construction_time_s: number;
  processes: ProductionProcess[];
}

export interface SaleResource {
  resource_id: string;
  units_sold_per_second: number;
  price_per_unit: number;
}

export interface SaleBuilding {
  id: string;
  master_id: string;
  name: string;
  construction_cost: number;
  construction_time_s: number;
  resources: SaleResource[];
}

export interface Gamedata {
  resources: Resource[];
  production_buildings: ProductionBuilding[];
  sale_buildings: SaleBuilding[];
}

// ============================================================
// Auth
// ============================================================

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  timezone: string;
}

export interface LoginResponse {
  user: User;
  session_expires_at: string;
}

export interface LogoutResponse {
  logged_out: boolean;
  user_id: string;
}

// ============================================================
// User
// ============================================================

export interface User {
  id: string;
  username: string;
  email: string;
  timezone: string;
  created_at: string;
}

export interface UpdateUserRequest {
  timezone?: string;
}

// ============================================================
// Company
// ============================================================

export interface Company {
  id: string;
  user_id: string;
  name: string;
  money: number;
  created_at: string;
}

export interface CreateCompanyRequest {
  name: string;
}

export interface UpdateCompanyRequest {
  name: string;
}

// ============================================================
// Inventory
// ============================================================

export interface InventoryItem {
  resource_id: string;
  quantity: number;
}

// ============================================================
// Market
// ============================================================

export interface MarketBuyRequest {
  resource_id: string;
  quantity: number;
}

export interface MarketSellRequest {
  resource_id: string;
  quantity: number;
}

export interface MarketTransactionResponse {
  company: Company;
  inventory: InventoryItem[];
}

// ============================================================
// Production
// ============================================================

export interface ProductionRun {
  id: string;
  process_id: string;
  production_cycles: number;
  started_at: string;
  ends_at: string;
  is_collected: boolean;
}

export interface CompanyProductionBuilding {
  id: string;
  production_building_id: string;
  level: number;
  construction_ends_at: string | null;
  active_run: ProductionRun | null;
}

export interface BuildProductionBuildingRequest {
  production_building_id: string;
}

export interface UpgradeBuildingRequest {
  levels: number;
}

export interface UpgradeProductionBuildingResponse {
  building: CompanyProductionBuilding;
  company: Pick<Company, 'id' | 'money'>;
}

export interface StartProductionRequest {
  process_id: string;
  cycles: number;
}

export interface StartProductionResponse {
  building: CompanyProductionBuilding;
  inventory: InventoryItem[];
}

export interface CollectProductionResponse {
  building: CompanyProductionBuilding;
  inventory: InventoryItem[];
}

// ============================================================
// Sale
// ============================================================

export interface SaleRun {
  id: string;
  resource_id: string;
  units_to_sell: number;
  started_at: string;
  ends_at: string;
  is_collected: boolean;
}

export interface CompanySaleBuilding {
  id: string;
  sale_building_id: string;
  level: number;
  construction_ends_at: string | null;
  active_run: SaleRun | null;
}

export interface BuildSaleBuildingRequest {
  sale_building_id: string;
}

export interface UpgradeSaleBuildingResponse {
  building: CompanySaleBuilding;
  company: Pick<Company, 'id' | 'money'>;
}

export interface StartSaleRequest {
  resource_id: string;
  units: number;
}

export interface StartSaleResponse {
  building: CompanySaleBuilding;
  inventory: InventoryItem[];
}

export interface CollectSaleResponse {
  building: CompanySaleBuilding;
  company: Company;
}
