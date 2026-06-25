import { useQuery } from '@tanstack/react-query';
import { companyApi } from '../api';

export function useInventory() {
  return useQuery({
    queryKey: ['inventory'],
    queryFn: companyApi.getInventory,
    staleTime: 10_000,
  });
}
