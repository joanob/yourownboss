import { api } from '@/services/api';
import type {
  User,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  LogoutResponse,
} from '@/types/api';

export const authApi = {
  register: (data: RegisterRequest) => api.post<User>('/api/v1/auth/register', data),
  login: (data: LoginRequest) => api.post<LoginResponse>('/api/v1/auth/login', data),
  logout: () => api.post<LogoutResponse>('/api/v1/auth/logout'),
};
