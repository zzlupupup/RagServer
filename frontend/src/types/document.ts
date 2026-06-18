export interface DocumentItem {
  id: number;
  kb_id: number;
  filename: string;
  original_filename: string;
  file_ext: string;
  mime_type: string;
  file_size: number;
  index_status: string;
  index_error?: string;
  chunk_count: number;
  created_at: string;
  updated_at: string;
}

