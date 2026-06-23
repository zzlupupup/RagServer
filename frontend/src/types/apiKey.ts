export interface ApiKey {
  id: number;
  created_by_user_id: number;
  bound_user_id: number;
  bound_user_email?: string;
  bound_user_display_name?: string;
  bound_user_role?: string;
  name: string;
  status: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
  api_key?: string;
}
