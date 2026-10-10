import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListParams } from '@/types/api';
import { categoryApi, unitApi } from '../api';
import type { CategoryInput, UnitInput } from '../types';

export const categoryKeys = {
  all: ['categories'] as const,
  list: (p: ListParams) => ['categories', 'list', p] as const,
};
export const unitKeys = {
  all: ['units'] as const,
  list: (p: ListParams) => ['units', 'list', p] as const,
};

export function useCategories(params: ListParams) {
  return useQuery({ queryKey: categoryKeys.list(params), queryFn: () => categoryApi.list(params), placeholderData: keepPreviousData });
}

/** All categories for dropdowns (max 100). */
export function useCategoryOptions() {
  return useCategories({ limit: 100, sort: 'name', order: 'asc' });
}

export function useSaveCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id?: number; input: CategoryInput }) =>
      id ? categoryApi.update(id, input) : categoryApi.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: categoryKeys.all }),
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => categoryApi.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: categoryKeys.all }),
  });
}

export function useUnits(params: ListParams) {
  return useQuery({ queryKey: unitKeys.list(params), queryFn: () => unitApi.list(params), placeholderData: keepPreviousData });
}

export function useUnitOptions() {
  return useUnits({ limit: 100, sort: 'name', order: 'asc' });
}

export function useSaveUnit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id?: number; input: UnitInput }) => (id ? unitApi.update(id, input) : unitApi.create(input)),
    onSuccess: () => qc.invalidateQueries({ queryKey: unitKeys.all }),
  });
}

export function useDeleteUnit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unitApi.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: unitKeys.all }),
  });
}
