import type { IdName } from '@/types/api';

export type PaymentMethod = 'CASH' | 'QRIS' | 'TRANSFER' | 'DEBIT';
export type TransactionStatus = 'COMPLETED' | 'VOIDED';

export interface TransactionItem {
  id: number;
  product_id: number;
  product_name: string;
  sku: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
}

export interface Payment {
  method: PaymentMethod;
  amount: number;
  tendered_amount: number;
  reference_number: string | null;
}

export interface Transaction {
  id: number;
  invoice_number: string;
  shift_id: number;
  cashier: IdName;
  subtotal: number;
  discount_type: 'NOMINAL' | 'PERCENT' | null;
  discount_value: number;
  discount_amount: number;
  tax_rate_bp: number;
  tax_amount: number;
  total: number;
  paid_amount: number;
  change_amount: number;
  gross_profit?: number;
  status: TransactionStatus;
  note: string | null;
  voided_at: string | null;
  voided_by: IdName | null;
  void_reason: string | null;
  item_count: number;
  items?: TransactionItem[];
  payments: Payment[];
  created_at: string;
}

export interface Receipt {
  store: { name: string; address: string | null; phone: string | null; logo_url: string | null; receipt_footer: string | null };
  transaction: Transaction;
  text: string;
}

export interface ShiftSummary {
  total_sales: number;
  total_transactions: number;
  voided_count: number;
  cash_sales: number;
  expected_cash: number;
  payment_totals: Partial<Record<PaymentMethod, number>>;
}

export interface Shift {
  id: number;
  user: IdName;
  status: 'OPEN' | 'CLOSED';
  opening_cash: number;
  expected_cash: number | null;
  actual_cash: number | null;
  cash_difference: number | null;
  total_sales: number | null;
  total_transactions: number | null;
  opening_note: string | null;
  closing_note: string | null;
  opened_at: string;
  closed_at: string | null;
}

export interface ShiftDetail extends Shift {
  summary: ShiftSummary;
}
