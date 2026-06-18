import { api } from './client';
import type { DocumentItem } from '../types/document';

export const documentsApi = {
  list: (kbId: number) => api.get<DocumentItem[]>(`/api/v1/kbs/${kbId}/documents`),
  upload: (kbId: number, file: File) => {
    const form = new FormData();
    form.append('file', file);
    return api.post<DocumentItem>(`/api/v1/kbs/${kbId}/documents/upload`, form);
  },
  remove: (id: number) => api.delete<{ deleted: boolean }>(`/api/v1/documents/${id}`),
};

