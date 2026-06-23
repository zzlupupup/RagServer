import { api } from './client';
import type { KnowledgeBase } from '../types/knowledgeBase';

export const knowledgeBasesApi = {
  list: () => api.get<KnowledgeBase[]>('/api/v1/kbs'),
  create: (body: { name: string }) => api.post<KnowledgeBase>('/api/v1/kbs', body),
  update: (id: number, body: { name?: string }) => api.patch<KnowledgeBase>(`/api/v1/kbs/${id}`, body),
  remove: (id: number) => api.delete<{ deleted: boolean }>(`/api/v1/kbs/${id}`),
};

