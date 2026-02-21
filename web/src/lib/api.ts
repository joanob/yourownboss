import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  withCredentials: true, // Important for sending cookies
  headers: {
    'Content-Type': 'application/json',
  },
});

// Flag to prevent multiple refresh attempts
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: Error | null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve();
    }
  });
  failedQueue = [];
};

// Response interceptor for handling 401 errors and auto-refresh
api.interceptors.response.use(
  (response: any) => response,
  async (error: any) => {
    const originalRequest = error.config;

    // Don't retry for auth endpoints or if no response
    if (!error.response || originalRequest.url?.includes('/auth/')) {
      return Promise.reject(error);
    }

    // If error is 401 and we haven't tried to refresh yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // If already refreshing, queue this request
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(() => api(originalRequest))
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Try to refresh the token
        await api.post('/auth/refresh');
        processQueue(null);
        return api(originalRequest);
      } catch (refreshError) {
        processQueue(new Error('Token refresh failed'));
        // Don't redirect here, let the app handle it
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

// Types
export interface Company {
  id: number;
  user_id: number;
  name: string;
  money: number; // int64 in thousandths
  created_at: string;
  updated_at: string;
}

export interface Resource {
  id: number;
  name: string;
  icon: string;
  description: string;
  price: number; // int64 in thousandths, price per pack
  pack_size: number; // units per pack
}

export interface InventoryItem {
  id: number;
  resource_id: number;
  name: string;
  icon: string;
  quantity: number; // total units owned
  price: number; // int64 in thousandths, price per pack
  pack_size: number; // units per pack
}

// API functions
export const authAPI = {
  register: (username: string, password: string) =>
    api.post('/auth/register', { username, password }),
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),
  logout: () => api.post('/auth/logout'),
  me: () => api.get('/auth/me'),
  refresh: () => api.post('/auth/refresh'),
};

export const companyAPI = {
  getMyCompany: () => api.get<Company>('/companies/me'),
  createCompany: (name: string) => api.post<Company>('/companies', { name }),
};

export const inventoryAPI = {
  getInventory: () => api.get<InventoryItem[]>('/inventory'),
  getResources: () => api.get<Resource[]>('/resources'),
};

export const marketAPI = {
  buy: (resourceId: number, packCount: number) =>
    api.post('/market/buy', { resource_id: resourceId, pack_count: packCount }),
  sell: (resourceId: number, packCount: number) =>
    api.post('/market/sell', {
      resource_id: resourceId,
      pack_count: packCount,
    }),
};

export default api;
