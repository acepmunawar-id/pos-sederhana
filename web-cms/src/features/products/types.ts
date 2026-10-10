import type { IdName } from '@/types/api';

export interface Product {
  id: number;
  sku: string;
  barcode: string | null;
  name: string;
  category: IdName;
  unit: IdName;
  price: number;
  cost: number;
  stock: number;
  min_stock: number;
  is_low_stock: boolean;
  image_url: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProductInput {
  category_id: number;
  unit_id: number;
  sku: string;
  barcode: string | null;
  name: string;
  price: number;
  cost: number;
  min_stock: number;
  is_active: boolean;
}

export interface CreateProductInput extends ProductInput {
  initial_stock: number;
}
