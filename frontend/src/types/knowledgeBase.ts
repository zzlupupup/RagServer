export interface KnowledgeBase {
  id: number;
  owner_user_id: number;
  owner_user_display_name?: string;
  name: string;
  description: string;
  visibility: 'public' | 'private' | string;
  status: string;
  document_count: number;
  can_manage: boolean;
  created_at: string;
  updated_at: string;
}
