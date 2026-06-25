import { useQuery } from '@tanstack/react-query';
import { productionApi } from '../api';

export function useProductionBuildings() {
  return useQuery({
    queryKey: ['production-buildings'],
    queryFn: productionApi.getBuildings,
    refetchInterval: 10_000,
  });
}
