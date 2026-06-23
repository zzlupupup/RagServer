export type UserRole = 'teacher' | 'student';

export interface User {
  id: number;
  email: string;
  display_name: string;
  role: UserRole;
  status?: string;
  last_login_at?: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  display_name: string;
  role: UserRole;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: User;
}
