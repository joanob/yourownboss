import { api } from '@/services/api';
import type { User, UpdateUserRequest } from '@/types/api';

export const usersApi = {
  getMe: () => api.get<User>('/api/v1/users/me'),
  updateMe: (data: UpdateUserRequest) => api.put<User>('/api/v1/users/me', data),
};
