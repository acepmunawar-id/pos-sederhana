import type { IdName } from '@/types/api';

export type MovementType = 'IN' | 'ADJUSTMENT' | 'SALE' | 'VOID';

export interface StockMovement {
  id: number;
  product: { id: number; name: string; sku: string };
  user: IdName;
  type: MovementType;
  quantity: number;
  stock_before: number;
  stock_after: number;
  reference_type: string | null;
  reference_id: number | null;
  note: string | null;
  created_at: string;
}

export interface StockInInput {
  items: { product_id: number; quantity: number }[];
  note?: string | null;
}

export interface StockAdjustmentInput {
  product_id: number;
  actual_stock: number;
  note: string;
}

export const movementLabel: Record<MovementType, string> = {
  IN: 'Stok masuk',
  ADJUSTMENT: 'Penyesuaian',
  SALE: 'Penjualan',
  VOID: 'Void',
};
