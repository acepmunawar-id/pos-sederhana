import { get, getPage, post } from '@/lib/api';
import type { ListParams } from '@/types/api';
import type { Receipt, Shift, ShiftDetail, Transaction } from './types';

export const transactionApi = {
  list: (params: ListParams) => getPage<Transaction>('/transactions', params),
  get: (id: number) => get<Transaction>(`/transactions/${id}`),
  receipt: (id: number) => get<Receipt>(`/transactions/${id}/receipt`),
  void: (id: number, reason: string) => post<Transaction>(`/transactions/${id}/void`, { reason }),
};

export const shiftApi = {
  list: (params: ListParams) => getPage<Shift>('/shifts', params),
  get: (id: number) => get<ShiftDetail>(`/shifts/${id}`),
};
