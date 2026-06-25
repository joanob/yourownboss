import { useQuery } from '@tanstack/react-query';
import { saleApi } from '../api';

export function useSaleBuildings() {
  return useQuery({
    queryKey: ['sale-buildings'],
    queryFn: saleApi.getBuildings,
    refetchInterval: 10_000,
  });
}
