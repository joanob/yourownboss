import { api } from '@/services/api';
import type { Gamedata, Resource, ProductionBuilding, SaleBuilding } from '@/types/api';

export const gamedataApi = {
  get: () => api.get<Gamedata>('/api/v1/gamedata'),
  getResources: () => api.get<Resource[]>('/api/v1/resources'),
  getProductionBuildings: () => api.get<ProductionBuilding[]>('/api/v1/production/buildings'),
  getSaleBuildings: () => api.get<SaleBuilding[]>('/api/v1/sale/buildings'),
};
