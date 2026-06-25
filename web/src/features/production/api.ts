import { api } from '@/services/api';
import type {
  CompanyProductionBuilding,
  BuildProductionBuildingRequest,
  UpgradeBuildingRequest,
  UpgradeProductionBuildingResponse,
  StartProductionRequest,
  StartProductionResponse,
  CollectProductionResponse,
} from '@/types/api';

export const productionApi = {
  getBuildings: () => api.get<CompanyProductionBuilding[]>('/api/v1/company/production/buildings'),
  build: (data: BuildProductionBuildingRequest) =>
    api.post<CompanyProductionBuilding>('/api/v1/company/production/buildings', data),
  upgrade: (id: string, data: UpgradeBuildingRequest) =>
    api.post<UpgradeProductionBuildingResponse>(
      `/api/v1/company/production/buildings/${id}/upgrade`,
      data
    ),
  start: (id: string, data: StartProductionRequest) =>
    api.post<StartProductionResponse>(`/api/v1/company/production/buildings/${id}/start`, data),
  collect: (id: string) =>
    api.post<CollectProductionResponse>(`/api/v1/company/production/buildings/${id}/collect`),
};
