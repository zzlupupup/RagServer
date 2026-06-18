export interface ApiKey {
  id: number;
  name: string;
  status: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
  api_key?: string;
}

