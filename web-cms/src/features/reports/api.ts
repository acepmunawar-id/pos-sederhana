import { download, get } from '@/lib/api';
import type { CashierRow, Dashboard, DateRange, ExportType, PaymentMethodRow, SalesReport, SalesSummary, TopProduct } from './types';

export const reportApi = {
  dashboard: () => get<Dashboard>('/reports/dashboard'),
  sales: (range: DateRange, group: 'daily' | 'monthly') => get<SalesReport>('/reports/sales', { ...range, group }),
  grossProfit: (range: DateRange) => get<SalesSummary>('/reports/gross-profit', { ...range }),
  topProducts: (range: DateRange, limit = 20) => get<TopProduct[]>('/reports/top-products', { ...range, limit }),
  paymentMethods: (range: DateRange) => get<PaymentMethodRow[]>('/reports/payment-methods', { ...range }),
  cashiers: (range: DateRange) => get<CashierRow[]>('/reports/cashiers', { ...range }),
  export: (type: ExportType, format: 'csv' | 'xlsx', range: DateRange) =>
    download('/reports/export', { ...range, type, format }, `laporan-${type}.${format}`),
};
