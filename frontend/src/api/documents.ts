import { api } from './client';
import type { DocumentItem } from '../types/document';
import type { PaginatedResponse } from '../types/pagination';

export const documentsApi = {
  list: (kbId: number, page = 1, pageSize = 10) =>
    api.get<PaginatedResponse<DocumentItem>>(`/api/v1/kbs/${kbId}/documents?page=${page}&page_size=${pageSize}`),
  upload: (kbId: number, file: File) => {
    const form = new FormData();
    form.append('file', file);
    return api.post<DocumentItem>(`/api/v1/kbs/${kbId}/documents/upload`, form);
  },
  remove: (id: number) => api.delete<{ deleted: boolean }>(`/api/v1/documents/${id}`),
};
