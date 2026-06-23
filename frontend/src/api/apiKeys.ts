import { api } from './client';
import type { ApiKey } from '../types/apiKey';
import type { PaginatedResponse } from '../types/pagination';

export const apiKeysApi = {
  list: (page = 1, pageSize = 10) => api.get<PaginatedResponse<ApiKey>>(`/api/v1/api-keys?page=${page}&page_size=${pageSize}`),
  create: (body: { bound_user_id: number }) => api.post<ApiKey>('/api/v1/api-keys', body),
  reveal: (id: number) => api.post<{ api_key: string }>(`/api/v1/api-keys/${id}/reveal`),
  disable: (id: number) => api.post<{ disabled: boolean }>(`/api/v1/api-keys/${id}/disable`),
  enable: (id: number) => api.post<{ enabled: boolean }>(`/api/v1/api-keys/${id}/enable`),
  remove: (id: number) => api.delete<{ deleted: boolean }>(`/api/v1/api-keys/${id}`),
};
