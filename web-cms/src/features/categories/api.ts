import { del, getPage, post, put } from '@/lib/api';
import type { ListParams } from '@/types/api';
import type { Category, CategoryInput, Unit, UnitInput } from './types';

export const categoryApi = {
  list: (params: ListParams) => getPage<Category>('/categories', params),
  create: (input: CategoryInput) => post<Category>('/categories', input),
  update: (id: number, input: CategoryInput) => put<Category>(`/categories/${id}`, input),
  remove: (id: number) => del(`/categories/${id}`),
};

export const unitApi = {
  list: (params: ListParams) => getPage<Unit>('/units', params),
  create: (input: UnitInput) => post<Unit>('/units', input),
  update: (id: number, input: UnitInput) => put<Unit>(`/units/${id}`, input),
  remove: (id: number) => del(`/units/${id}`),
};
