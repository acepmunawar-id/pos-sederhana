import { getPage, post } from '@/lib/api';
import type { ListParams } from '@/types/api';
import type { Product } from '@/features/products/types';
import type { StockAdjustmentInput, StockInInput, StockMovement } from './types';

export const stockApi = {
  movements: (params: ListParams) => getPage<StockMovement>('/stock/movements', params),
  low: (params: ListParams) => getPage<Product>('/stock/low', params),
  stockIn: (input: StockInInput) => post<StockMovement[]>('/stock/in', input),
  adjust: (input: StockAdjustmentInput) => post<StockMovement>('/stock/adjustments', input),
};
