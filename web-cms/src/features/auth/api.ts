import { get, post, put } from '@/lib/api';
import type { ChangePasswordInput, LoginInput, TokenResponse, User } from './types';

export const authApi = {
  login: (input: LoginInput) => post<TokenResponse>('/auth/login', input),
  logout: (refreshToken: string) => post<null>('/auth/logout', { refresh_token: refreshToken }),
  me: () => get<User>('/auth/me'),
  changePassword: (input: ChangePasswordInput) => put<null>('/auth/password', input),
};
