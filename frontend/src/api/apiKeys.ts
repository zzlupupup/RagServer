import { api } from './client';
import type { ApiKey } from '../types/apiKey';

export const apiKeysApi = {
  list: () => api.get<ApiKey[]>('/api/v1/api-keys'),
  create: (name: string) => api.post<ApiKey>('/api/v1/api-keys', { name }),
  reveal: (id: number) => api.post<{ api_key: string }>(`/api/v1/api-keys/${id}/reveal`),
  disable: (id: number) => api.post<{ disabled: boolean }>(`/api/v1/api-keys/${id}/disable`),
  remove: (id: number) => api.delete<{ deleted: boolean }>(`/api/v1/api-keys/${id}`),
};

