import { api } from './client';
import type { LoginResponse, RegisterRequest, User } from '../types/user';

export const authApi = {
  register: (body: RegisterRequest) => api.post<User>('/api/v1/auth/register', body),
  login: (body: { email: string; password: string }) => api.post<LoginResponse>('/api/v1/auth/login', body),
};
