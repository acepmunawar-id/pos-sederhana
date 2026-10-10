import { keepPreviousData, useMutation, useQuery } from '@tanstack/react-query';
import type { ListParams } from '@/types/api';
import { useInvalidateProducts } from '@/features/products/hooks/useProducts';
import { stockApi } from '../api';
import type { StockAdjustmentInput, StockInInput } from '../types';

export function useStockMovements(params: ListParams) {
  return useQuery({
    queryKey: ['stock', 'movements', params],
    queryFn: () => stockApi.movements(params),
    placeholderData: keepPreviousData,
  });
}

export function useLowStock(params: ListParams) {
  return useQuery({ queryKey: ['stock', 'low', params], queryFn: () => stockApi.low(params), placeholderData: keepPreviousData });
}

export function useStockIn() {
  const invalidate = useInvalidateProducts();
  return useMutation({ mutationFn: (input: StockInInput) => stockApi.stockIn(input), onSuccess: invalidate });
}

export function useStockAdjustment() {
  const invalidate = useInvalidateProducts();
  return useMutation({ mutationFn: (input: StockAdjustmentInput) => stockApi.adjust(input), onSuccess: invalidate });
}
