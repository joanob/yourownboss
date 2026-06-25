import { useQuery } from '@tanstack/react-query';
import { gamedataApi } from '../api';

export function useGamedata() {
  return useQuery({
    queryKey: ['gamedata'],
    queryFn: gamedataApi.get,
    staleTime: 10 * 60_000,
  });
}
