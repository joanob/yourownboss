import { useQuery } from '@tanstack/react-query';
import { companyApi } from '../api';
import { ApiError } from '@/services/api';

export function useCompany() {
  return useQuery({
    queryKey: ['company'],
    queryFn: companyApi.get,
    retry: (failureCount, error) => {
      if (error instanceof ApiError && error.status < 500) return false;
      return failureCount < 2;
    },
  });
}
