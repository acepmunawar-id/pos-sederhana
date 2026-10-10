import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListParams } from '@/types/api';
import { shiftApi, transactionApi } from '../api';

export function useTransactions(params: ListParams) {
  return useQuery({
    queryKey: ['transactions', 'list', params],
    queryFn: () => transactionApi.list(params),
    placeholderData: keepPreviousData,
  });
}

export function useTransaction(id: number) {
  return useQuery({ queryKey: ['transactions', 'detail', id], queryFn: () => transactionApi.get(id), enabled: id > 0 });
}

export function useReceipt(id: number, enabled: boolean) {
  return useQuery({ queryKey: ['transactions', 'receipt', id], queryFn: () => transactionApi.receipt(id), enabled: enabled && id > 0 });
}

export function useVoidTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, reason }: { id: number; reason: string }) => transactionApi.void(id, reason),
    onSuccess: () =>
      Promise.all(
        ['transactions', 'products', 'stock', 'dashboard', 'reports', 'shifts'].map((k) => qc.invalidateQueries({ queryKey: [k] })),
      ),
  });
}

export function useShifts(params: ListParams) {
  return useQuery({ queryKey: ['shifts', 'list', params], queryFn: () => shiftApi.list(params), placeholderData: keepPreviousData });
}

export function useShift(id: number | null) {
  return useQuery({ queryKey: ['shifts', 'detail', id], queryFn: () => shiftApi.get(id!), enabled: !!id });
}
