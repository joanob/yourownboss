import { api } from '@/services/api';
import type { MarketBuyRequest, MarketSellRequest, MarketTransactionResponse } from '@/types/api';

export const marketApi = {
  buy: (data: MarketBuyRequest) => api.post<MarketTransactionResponse>('/api/v1/market/buy', data),
  sell: (data: MarketSellRequest) =>
    api.post<MarketTransactionResponse>('/api/v1/market/sell', data),
};
