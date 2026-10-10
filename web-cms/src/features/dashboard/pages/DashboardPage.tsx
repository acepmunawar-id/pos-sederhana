import { Link } from 'react-router-dom';
import { AlertTriangle, Banknote, ChevronRight, Receipt, TrendingUp, Trophy } from 'lucide-react';
import { PageHeader } from '@/components/layout/PageHeader';
import { Badge, Card, EmptyState, ErrorState, SkeletonCard, StatCard } from '@/components/ui';
import { BarChart } from '@/components/ui/BarChart';
import { errorMessage } from '@/lib/api';
import { formatNumber, formatPercentBP, formatRupiah, formatShortDay } from '@/lib/format';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { ProductThumb } from '@/features/products/components/ProductThumb';
import { useDashboard } from '@/features/reports/hooks/useReports';

function Trend({ today, yesterday }: { today: number; yesterday: number }) {
  if (yesterday === 0) return <>Kemarin {formatRupiah(0)}</>;
  const pct = Math.round(((today - yesterday) / yesterday) * 100);
  return (
    <span className={pct >= 0 ? 'text-success' : 'text-danger'}>
      {pct >= 0 ? '▲' : '▼'} {Math.abs(pct)}% dari kemarin
    </span>
  );
}

export default function DashboardPage() {
  const { user } = useAuth();
  const query = useDashboard();
  const hour = new Date().getHours();
  const greeting = hour < 11 ? 'Selamat pagi' : hour < 15 ? 'Selamat siang' : hour < 18 ? 'Selamat sore' : 'Selamat malam';

  return (
    <>
      <PageHeader title="Dashboard" description={`${greeting}, ${user?.name.split(' ')[0]}. Ini ringkasan toko hari ini.`} />
      {query.isLoading ? (
        <div className="grid grid-cols-2 gap-3 sm:gap-4 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => <SkeletonCard key={i} lines={1} />)}
          <SkeletonCard className="sm:col-span-2 xl:col-span-3" lines={8} />
          <SkeletonCard lines={8} />
        </div>
      ) : query.isError || !query.data ? (
        <Card><ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} /></Card>
      ) : (
        <DashboardContent data={query.data} />
      )}
    </>
  );
}

function DashboardContent({ data }: { data: NonNullable<ReturnType<typeof useDashboard>['data']> }) {
  const { today, yesterday } = data;
  const weekTotal = data.last_7_days.reduce((s, d) => s + d.total_revenue, 0);
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3 sm:gap-4 xl:grid-cols-4">
        <StatCard
          label="Penjualan hari ini"
          value={formatRupiah(today.total_revenue)}
          hint={<Trend today={today.total_revenue} yesterday={yesterday.total_revenue} />}
          icon={<Banknote className="h-5 w-5" />}
        />
        <StatCard
          label="Jumlah transaksi"
          value={formatNumber(today.transaction_count)}
          hint={`Kemarin ${formatNumber(yesterday.transaction_count)} transaksi`}
          icon={<Receipt className="h-5 w-5" />}
          tone="primary"
        />
        <StatCard
          label="Laba kotor hari ini"
          value={formatRupiah(today.gross_profit)}
          hint={today.net_sales > 0 ? `Margin ${formatPercentBP(today.margin_bp)}` : 'Belum ada penjualan'}
          icon={<TrendingUp className="h-5 w-5" />}
          tone="success"
        />
        <StatCard
          label="Stok menipis"
          value={`${data.low_stock_count} produk`}
          hint={data.low_stock_count > 0 ? 'Segera lakukan stok masuk' : 'Semua stok aman'}
          icon={<AlertTriangle className="h-5 w-5" />}
          tone={data.low_stock_count > 0 ? 'warning' : 'success'}
        />
      </div>

      <div className="grid gap-4 xl:grid-cols-3">
        <Card title="Penjualan 7 hari terakhir" subtitle={`Total ${formatRupiah(weekTotal)}`} className="xl:col-span-2">
          {weekTotal === 0 ? (
            <EmptyState className="py-16" title="Belum ada penjualan" description="Grafik akan terisi setelah ada transaksi dari aplikasi kasir." />
          ) : (
            <BarChart
              ariaLabel="Grafik pendapatan harian 7 hari terakhir"
              formatValue={formatRupiah}
              data={data.last_7_days.map((d) => ({
                label: formatShortDay(d.date),
                value: d.total_revenue,
                detail: `${d.transaction_count} transaksi`,
              }))}
            />
          )}
        </Card>

        <Card
          title="Produk terlaris"
          subtitle="7 hari terakhir"
          padded={false}
          action={<Link to="/reports" className="inline-flex h-touch items-center text-footnote font-semibold text-primary">Laporan <ChevronRight className="h-4 w-4" /></Link>}
        >
          {data.top_products.length === 0 ? (
            <EmptyState className="py-12" icon={<Trophy className="h-7 w-7" />} title="Belum ada data" />
          ) : (
            <ol className="divide-y divide-separator/50 px-2 pb-2">
              {data.top_products.map((p, i) => (
                <li key={p.product_id} className="flex items-center gap-3 px-3 py-2.5">
                  <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-fill text-footnote font-bold tabular-nums">{i + 1}</span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-subheadline font-medium">{p.product_name}</p>
                    <p className="text-caption text-text-secondary">{formatNumber(p.quantity)} terjual</p>
                  </div>
                  <span className="text-subheadline tabular-nums">{formatRupiah(p.revenue)}</span>
                </li>
              ))}
            </ol>
          )}
        </Card>
      </div>

      <Card
        title="Stok menipis"
        padded={false}
        action={<Link to="/stock" className="inline-flex h-touch items-center text-footnote font-semibold text-primary">Kelola stok <ChevronRight className="h-4 w-4" /></Link>}
      >
        {data.low_stock.length === 0 ? (
          <EmptyState className="py-10" title="Semua stok aman" description="Tidak ada produk di bawah stok minimum." />
        ) : (
          <ul className="grid overflow-hidden rounded-b-xl sm:grid-cols-2 lg:grid-cols-3">
            {data.low_stock.map((p) => (
              <li key={p.id} className="flex items-center gap-3 border-t border-separator/50 px-5 py-3">
                <ProductThumb url={p.image_url} name={p.name} />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-subheadline font-medium">{p.name}</p>
                  <p className="text-caption text-text-secondary">Minimum {p.min_stock}</p>
                </div>
                <Badge tone={p.stock === 0 ? 'danger' : 'warning'}>Sisa {p.stock}</Badge>
              </li>
            ))}
          </ul>
        )}
      </Card>
    </div>
  );
}
