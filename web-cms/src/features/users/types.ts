export type { Role, User } from '@/features/auth/types';

export interface CreateUserInput {
  name: string;
  email: string;
  password: string;
  role_id: number;
  is_active: boolean;
}

export interface UpdateUserInput {
  name: string;
  email: string;
  role_id: number;
  is_active: boolean;
  password?: string;
}
