import { api } from './client';
import type { User } from '../types/user';

export const usersApi = {
  list: () => api.get<User[]>('/api/v1/users'),
};
