import type { RoleCode } from '@/types/api';

export interface Role {
  id: number;
  code: RoleCode;
  name: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  role: Role;
  is_active: boolean;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface TokenResponse {
  access_token: string;
  access_token_expires_at: string;
  refresh_token: string;
  refresh_token_expires_at: string;
  token_type: 'Bearer';
  user: User;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface ChangePasswordInput {
  old_password: string;
  new_password: string;
}
