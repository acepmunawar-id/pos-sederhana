import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { reportApi } from '../api';
import type { DateRange } from '../types';

export function useDashboard() {
  return useQuery({ queryKey: ['dashboard'], queryFn: reportApi.dashboard, refetchInterval: 60_000 });
}

export function useSalesReport(range: DateRange, group: 'daily' | 'monthly') {
  return useQuery({ queryKey: ['reports', 'sales', range, group], queryFn: () => reportApi.sales(range, group), placeholderData: keepPreviousData });
}

export function useTopProducts(range: DateRange) {
  return useQuery({ queryKey: ['reports', 'top', range], queryFn: () => reportApi.topProducts(range), placeholderData: keepPreviousData });
}

export function usePaymentMethods(range: DateRange) {
  return useQuery({ queryKey: ['reports', 'methods', range], queryFn: () => reportApi.paymentMethods(range), placeholderData: keepPreviousData });
}

export function useCashierReport(range: DateRange) {
  return useQuery({ queryKey: ['reports', 'cashiers', range], queryFn: () => reportApi.cashiers(range), placeholderData: keepPreviousData });
}
