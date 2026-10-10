import { del, get, getPage, post, put } from '@/lib/api';
import type { ListParams } from '@/types/api';
import type { CreateUserInput, Role, UpdateUserInput, User } from './types';

export const userApi = {
  list: (params: ListParams) => getPage<User>('/users', params),
  roles: () => get<Role[]>('/roles'),
  create: (input: CreateUserInput) => post<User>('/users', input),
  update: (id: number, input: UpdateUserInput) => put<User>(`/users/${id}`, input),
  remove: (id: number) => del(`/users/${id}`),
};
