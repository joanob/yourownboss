import { api } from '@/services/api';
import type {
  Company,
  CreateCompanyRequest,
  UpdateCompanyRequest,
  InventoryItem,
} from '@/types/api';

export const companyApi = {
  get: () => api.get<Company>('/api/v1/company'),
  create: (data: CreateCompanyRequest) => api.post<Company>('/api/v1/company', data),
  update: (data: UpdateCompanyRequest) => api.put<Company>('/api/v1/company', data),
  delete: () => api.delete<void>('/api/v1/company'),
  getInventory: () => api.get<InventoryItem[]>('/api/v1/company/inventory'),
};
