import { useState } from 'react';
import { FileDown, FileSpreadsheet } from 'lucide-react';
import { PageHeader } from '@/components/layout/PageHeader';
import {
  Button,
  Card,
  EmptyState,
  ErrorState,
  SegmentedControl,
  SkeletonCard,
  StatCard,
  Table,
  useToast,
  type Column,
} from '@/components/ui';
import { BarChart } from '@/components/ui/BarChart';
import { ProportionBars } from '@/components/ui/ProportionBars';
import { errorMessage } from '@/lib/api';
import { formatNumber, formatPercentBP, formatPeriod, formatRupiah, formatShortDay, paymentMethodLabel } from '@/lib/format';
import { reportApi } from '../api';
import { DateRangeFilter, presetRange } from '../components/DateRangeFilter';
import { useCashierReport, usePaymentMethods, useSalesReport, useTopProducts } from '../hooks/useReports';
import type { CashierRow, DateRange, ExportType, SalesPeriod, TopProduct } from '../types';

type Tab = 'sales' | 'products' | 'methods' | 'cashiers';

export default function ReportsPage() {
  const toast = useToast();
  const [tab, setTab] = useState<Tab>('sales');
  const [range, setRange] = useState<DateRange>(presetRange('30d'));
  const [group, setGroup] = useState<'daily' | 'monthly'>('daily');
  const [exporting, setExporting] = useState<'csv' | 'xlsx' | null>(null);

  const exportType: ExportType =
    tab === 'sales' ? (group === 'daily' ? 'sales-daily' : 'sales-monthly') : tab === 'products' ? 'top-products' : tab === 'methods' ? 'payment-methods' : 'cashiers';

  const doExport = async (format: 'csv' | 'xlsx') => {
    setExporting(format);
    try {
      await reportApi.export(exportType, format, range);
      toast.success('File laporan berhasil diunduh');
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setExporting(null);
    }
  };

  return (
    <>
      <PageHeader
        title="Laporan"
        description="Analisis penjualan, produk, metode bayar, dan kinerja kasir."
        actions={
          <>
            <Button variant="secondary" icon={<FileDown className="h-4 w-4" />} loading={exporting === 'csv'} onClick={() => doExport('csv')}>
              Export CSV
            </Button>
            <Button variant="secondary" icon={<FileSpreadsheet className="h-4 w-4" />} loading={exporting === 'xlsx'} onClick={() => doExport('xlsx')}>
              Export Excel
            </Button>
          </>
        }
      />
      <div className="mb-4 flex flex-col gap-3">
        <DateRangeFilter value={range} onChange={setRange} />
        <SegmentedControl
          value={tab}
          onChange={setTab}
          segments={[
            { value: 'sales', label: 'Penjualan & Laba' },
            { value: 'products', label: 'Produk Terlaris' },
            { value: 'methods', label: 'Metode Bayar' },
            { value: 'cashiers', label: 'Per Kasir' },
          ]}
        />
      </div>
      {tab === 'sales' && <SalesTab range={range} group={group} onGroup={setGroup} />}
      {tab === 'products' && <TopProductsTab range={range} />}
      {tab === 'methods' && <MethodsTab range={range} />}
      {tab === 'cashiers' && <CashiersTab range={range} />}
    </>
  );
}

function SalesTab({ range, group, onGroup }: { range: DateRange; group: 'daily' | 'monthly'; onGroup: (g: 'daily' | 'monthly') => void }) {
  const q = useSalesReport(range, group);
  if (q.isLoading) return <LoadingGrid />;
  if (q.isError || !q.data) return <Card><ErrorState message={errorMessage(q.error)} onRetry={() => q.refetch()} /></Card>;
  const { summary, rows } = q.data;

  const columns: Column<SalesPeriod>[] = [
    { key: 'period', header: 'Periode', render: (r) => formatPeriod(r.period) },
    { key: 'count', header: 'Transaksi', align: 'right', render: (r) => formatNumber(r.transaction_count) },
    { key: 'gross', header: 'Penjualan Kotor', align: 'right', render: (r) => formatRupiah(r.gross_sales), hideOnMobile: true },
    { key: 'disc', header: 'Diskon', align: 'right', render: (r) => formatRupiah(r.discount_total), hideOnMobile: true },
    { key: 'net', header: 'Penjualan Bersih', align: 'right', render: (r) => formatRupiah(r.net_sales) },
    { key: 'tax', header: 'Pajak', align: 'right', render: (r) => formatRupiah(r.tax_total), hideOnMobile: true },
    { key: 'cost', header: 'HPP', align: 'right', render: (r) => formatRupiah(r.cost_total), hideOnMobile: true },
    { key: 'profit', header: 'Laba Kotor', align: 'right', render: (r) => <span className="font-semibold">{formatRupiah(r.gross_profit)}</span> },
  ];

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3 sm:gap-4 xl:grid-cols-4">
        <StatCard label="Total diterima" value={formatRupiah(summary.total_revenue)} hint={`${formatNumber(summary.transaction_count)} transaksi`} />
        <StatCard label="Penjualan bersih" value={formatRupiah(summary.net_sales)} hint={`Diskon ${formatRupiah(summary.discount_total)}`} />
        <StatCard label="HPP (modal)" value={formatRupiah(summary.cost_total)} hint={`Pajak ${formatRupiah(summary.tax_total)}`} tone="warning" />
        <StatCard label="Laba kotor" value={formatRupiah(summary.gross_profit)} hint={`Margin ${formatPercentBP(summary.margin_bp)}`} tone="success" />
      </div>
      <Card
        title={group === 'daily' ? 'Penjualan bersih per hari' : 'Penjualan bersih per bulan'}
        action={
          <SegmentedControl
            value={group}
            onChange={onGroup}
            segments={[
              { value: 'daily', label: 'Harian' },
              { value: 'monthly', label: 'Bulanan' },
            ]}
          />
        }
      >
        {rows.length === 0 ? (
          <EmptyState className="py-14" title="Tidak ada penjualan" description="Belum ada transaksi pada rentang tanggal ini." />
        ) : (
          <BarChart
            ariaLabel="Grafik penjualan bersih"
            formatValue={formatRupiah}
            data={rows.map((r) => ({
              label: group === 'daily' ? formatShortDay(r.period) : formatPeriod(r.period),
              value: r.net_sales,
              detail: `${r.transaction_count} transaksi · laba ${formatRupiah(r.gross_profit)}`,
            }))}
          />
        )}
      </Card>
      {rows.length > 0 && (
        <Card title="Rincian" padded={false}>
          <Table columns={columns} rows={rows} rowKey={(r) => r.period} />
        </Card>
      )}
    </div>
  );
}

type RankedProduct = TopProduct & { rank: number };

function TopProductsTab({ range }: { range: DateRange }) {
  const q = useTopProducts(range);
  const columns: Column<RankedProduct>[] = [
    { key: 'rank', header: '#', render: (r) => <span className="font-bold tabular-nums text-text-secondary">{r.rank}</span> },
    { key: 'name', header: 'Produk', render: (r) => <div><p className="font-semibold">{r.product_name}</p><p className="text-caption text-text-secondary">{r.sku}</p></div> },
    { key: 'qty', header: 'Terjual', align: 'right', render: (r) => formatNumber(r.quantity) },
    { key: 'rev', header: 'Pendapatan', align: 'right', render: (r) => formatRupiah(r.revenue) },
    { key: 'profit', header: 'Laba Kotor', align: 'right', render: (r) => <span className="text-success">{formatRupiah(r.profit)}</span> },
  ];
  if (q.isError) return <Card><ErrorState message={errorMessage(q.error)} onRetry={() => q.refetch()} /></Card>;
  return (
    <Card title="Produk terlaris" subtitle="Diurutkan berdasarkan jumlah terjual. Pendapatan sebelum diskon transaksi." padded={false}>
      <Table
        columns={columns}
        rows={q.data?.map((r, i) => ({ ...r, rank: i + 1 }))}
        rowKey={(r) => r.product_id}
        loading={q.isLoading}
        empty={<EmptyState title="Belum ada produk terjual" />}
      />
    </Card>
  );
}

function MethodsTab({ range }: { range: DateRange }) {
  const q = usePaymentMethods(range);
  if (q.isLoading) return <SkeletonCard lines={6} />;
  if (q.isError || !q.data) return <Card><ErrorState message={errorMessage(q.error)} onRetry={() => q.refetch()} /></Card>;
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card title="Porsi per metode bayar">
        {q.data.length === 0 ? (
          <EmptyState className="py-10" title="Belum ada pembayaran" />
        ) : (
          <ProportionBars
            formatValue={formatRupiah}
            rows={q.data.map((r) => ({ label: paymentMethodLabel[r.method] ?? r.method, value: r.amount, detail: `${formatNumber(r.transaction_count)} transaksi` }))}
          />
        )}
      </Card>
      <Card title="Tabel" padded={false}>
        <Table
          columns={[
            { key: 'm', header: 'Metode', render: (r) => paymentMethodLabel[r.method] ?? r.method },
            { key: 'c', header: 'Transaksi', align: 'right', render: (r) => formatNumber(r.transaction_count) },
            { key: 'a', header: 'Nominal', align: 'right', render: (r) => <span className="font-semibold">{formatRupiah(r.amount)}</span> },
          ]}
          rows={q.data}
          rowKey={(r) => r.method}
          empty={<EmptyState title="Belum ada pembayaran" />}
        />
      </Card>
    </div>
  );
}

function CashiersTab({ range }: { range: DateRange }) {
  const q = useCashierReport(range);
  const columns: Column<CashierRow>[] = [
    { key: 'name', header: 'Kasir', render: (r) => <span className="font-semibold">{r.user_name}</span> },
    { key: 'count', header: 'Transaksi', align: 'right', render: (r) => formatNumber(r.transaction_count) },
    { key: 'avg', header: 'Rata-rata / transaksi', align: 'right', render: (r) => formatRupiah(r.transaction_count ? Math.round(r.total_sales / r.transaction_count) : 0) },
    { key: 'total', header: 'Total Penjualan', align: 'right', render: (r) => <span className="font-semibold">{formatRupiah(r.total_sales)}</span> },
  ];
  if (q.isError) return <Card><ErrorState message={errorMessage(q.error)} onRetry={() => q.refetch()} /></Card>;
  return (
    <Card title="Penjualan per kasir" padded={false}>
      <Table columns={columns} rows={q.data} rowKey={(r) => r.user_id} loading={q.isLoading} empty={<EmptyState title="Belum ada transaksi" />} />
    </Card>
  );
}

function LoadingGrid() {
  return (
    <div className="grid grid-cols-2 gap-3 sm:gap-4 xl:grid-cols-4">
      {Array.from({ length: 4 }).map((_, i) => <SkeletonCard key={i} lines={1} />)}
      <SkeletonCard className="sm:col-span-2 xl:col-span-4" lines={8} />
    </div>
  );
}
