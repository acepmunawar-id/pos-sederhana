import type { Product } from '@/features/products/types';

export interface SalesSummary {
  transaction_count: number;
  gross_sales: number;
  discount_total: number;
  tax_total: number;
  net_sales: number;
  total_revenue: number;
  cost_total: number;
  gross_profit: number;
  margin_bp: number;
}

export interface SalesPeriod extends SalesSummary {
  period: string;
}

export interface SalesReport {
  from: string;
  to: string;
  group: 'daily' | 'monthly';
  summary: SalesSummary;
  rows: SalesPeriod[];
}

export interface TopProduct {
  product_id: number;
  product_name: string;
  sku: string;
  quantity: number;
  revenue: number;
  profit: number;
}

export interface PaymentMethodRow {
  method: string;
  transaction_count: number;
  amount: number;
}

export interface CashierRow {
  user_id: number;
  user_name: string;
  transaction_count: number;
  total_sales: number;
}

export interface Dashboard {
  today: SalesSummary;
  yesterday: SalesSummary;
  last_7_days: { date: string; total_revenue: number; transaction_count: number }[];
  top_products: TopProduct[];
  low_stock: Product[];
  low_stock_count: number;
}

export interface DateRange {
  date_from: string;
  date_to: string;
}

export type ExportType = 'sales-daily' | 'sales-monthly' | 'top-products' | 'payment-methods' | 'cashiers';
