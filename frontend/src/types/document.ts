export interface DocumentItem {
  id: number;
  kb_id: number;
  uploaded_by_user_id: number;
  uploaded_by_user_display_name?: string;
  filename: string;
  original_filename: string;
  file_ext: string;
  mime_type: string;
  file_size: number;
  file_hash: string;
  storage_path: string;
  index_status: string;
  index_error?: string;
  chunk_count: number;
  can_delete: boolean;
  created_at: string;
  updated_at: string;
}
