import { api } from '@/services/api';
import type {
  CompanySaleBuilding,
  BuildSaleBuildingRequest,
  UpgradeBuildingRequest,
  UpgradeSaleBuildingResponse,
  StartSaleRequest,
  StartSaleResponse,
  CollectSaleResponse,
} from '@/types/api';

export const saleApi = {
  getBuildings: () => api.get<CompanySaleBuilding[]>('/api/v1/company/sale/buildings'),
  build: (data: BuildSaleBuildingRequest) =>
    api.post<CompanySaleBuilding>('/api/v1/company/sale/buildings', data),
  upgrade: (id: string, data: UpgradeBuildingRequest) =>
    api.post<UpgradeSaleBuildingResponse>(`/api/v1/company/sale/buildings/${id}/upgrade`, data),
  start: (id: string, data: StartSaleRequest) =>
    api.post<StartSaleResponse>(`/api/v1/company/sale/buildings/${id}/start`, data),
  collect: (id: string) =>
    api.post<CollectSaleResponse>(`/api/v1/company/sale/buildings/${id}/collect`),
};
