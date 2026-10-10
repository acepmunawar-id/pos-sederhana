import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListParams } from '@/types/api';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { userApi } from '../api';
import type { CreateUserInput, UpdateUserInput } from '../types';

export function useUsers(params: ListParams, enabled = true) {
  return useQuery({ queryKey: ['users', 'list', params], queryFn: () => userApi.list(params), placeholderData: keepPreviousData, enabled });
}

/** Users for filter dropdowns; only managers can list users. */
export function useUserOptions() {
  const { hasRole } = useAuth();
  return useUsers({ limit: 100, sort: 'name', order: 'asc' }, hasRole('owner', 'admin'));
}

export function useRoles() {
  return useQuery({ queryKey: ['roles'], queryFn: userApi.roles, staleTime: Infinity });
}

export function useSaveUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (v: { id: number; input: UpdateUserInput } | { id?: undefined; input: CreateUserInput }) =>
      v.id ? userApi.update(v.id, v.input) : userApi.create(v.input as CreateUserInput),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  });
}

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (id: number) => userApi.remove(id), onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }) });
}
