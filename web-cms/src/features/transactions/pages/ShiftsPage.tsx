import { useState, type ReactNode } from 'react';
import { Timer } from 'lucide-react';
import { PageHeader } from '@/components/layout/PageHeader';
import {
  Badge,
  Card,
  EmptyState,
  ErrorState,
  Input,
  Modal,
  Pagination,
  SegmentedControl,
  Select,
  Skeleton,
  Table,
  type Column,
} from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { cn } from '@/lib/cn';
import { formatDateTime, formatRupiah, paymentMethodLabel } from '@/lib/format';
import { useListState } from '@/lib/use-list-state';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { useUserOptions } from '@/features/users/hooks/useUsers';
import { useShift, useShifts } from '../hooks/useTransactions';
import type { Shift } from '../types';

function Difference({ value }: { value: number | null }) {
  if (value === null) return <span className="text-text-secondary">-</span>;
  return (
    <span className={cn('font-semibold tabular-nums', value === 0 ? 'text-success' : value < 0 ? 'text-danger' : 'text-warning')}>
      {value > 0 ? '+' : ''}
      {formatRupiah(value)}
    </span>
  );
}

export default function ShiftsPage() {
  const { hasRole } = useAuth();
  const isManager = hasRole('owner', 'admin');
  const users = useUserOptions();
  const [status, setStatus] = useState<'' | 'OPEN' | 'CLOSED'>('');
  const list = useListState<{ status?: string; user_id?: number; date_from?: string; date_to?: string }>({}, { sort: 'opened_at', order: 'desc' });
  const query = useShifts(list.params);
  const [selected, setSelected] = useState<number | null>(null);

  const columns: Column<Shift>[] = [
    { key: 'opened', header: 'Dibuka', sortKey: 'opened_at', render: (s) => <span className="whitespace-nowrap">{formatDateTime(s.opened_at)}</span> },
    { key: 'closed', header: 'Ditutup', sortKey: 'closed_at', render: (s) => <span className="whitespace-nowrap">{formatDateTime(s.closed_at)}</span>, hideOnMobile: true },
    { key: 'cashier', header: 'Kasir', sortKey: 'cashier', render: (s) => s.user.name },
    { key: 'status', header: 'Status', render: (s) => (s.status === 'OPEN' ? <Badge tone="primary">Berjalan</Badge> : <Badge>Selesai</Badge>) },
    { key: 'sales', header: 'Penjualan', align: 'right', render: (s) => <span className="tabular-nums">{s.total_sales === null ? '-' : formatRupiah(s.total_sales)}</span> },
    { key: 'diff', header: 'Selisih Kas', align: 'right', render: (s) => <Difference value={s.cash_difference} /> },
  ];

  return (
    <>
      <PageHeader title="Shift Kasir" description="Rekap kas awal, penjualan, dan selisih kas setiap shift." />
      <Card padded={false}>
        <div className="flex flex-wrap items-end gap-3 p-4">
          <SegmentedControl
            value={status}
            onChange={(v) => {
              setStatus(v);
              list.setFilter('status', v || undefined);
            }}
            segments={[
              { value: '', label: 'Semua' },
              { value: 'OPEN', label: 'Berjalan' },
              { value: 'CLOSED', label: 'Selesai' },
            ]}
          />
          {isManager && (
            <Select
              className="w-full sm:w-44"
              aria-label="Kasir"
              placeholder="Semua kasir"
              value={list.filters.user_id ?? ''}
              onChange={(e) => list.setFilter('user_id', e.target.value ? Number(e.target.value) : undefined)}
              options={(users.data?.items ?? []).map((u) => ({ value: u.id, label: u.name }))}
            />
          )}
          <Input className="w-full sm:w-40" type="date" aria-label="Dari tanggal" value={list.filters.date_from ?? ''} onChange={(e) => list.setFilter('date_from', e.target.value || undefined)} />
          <Input className="w-full sm:w-40" type="date" aria-label="Sampai tanggal" value={list.filters.date_to ?? ''} onChange={(e) => list.setFilter('date_to', e.target.value || undefined)} />
        </div>
        {query.isError ? (
          <ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} />
        ) : (
          <Table
            columns={columns}
            rows={query.data?.items}
            rowKey={(s) => s.id}
            loading={query.isLoading}
            sort={list.sort}
            onSort={list.toggleSort}
            onRowClick={(s) => setSelected(s.id)}
            empty={<EmptyState icon={<Timer className="h-7 w-7" />} title="Belum ada shift" description="Shift dibuka dari aplikasi kasir." />}
          />
        )}
        <Pagination meta={query.data?.meta} onChange={list.setPage} />
      </Card>
      <ShiftDetailModal id={selected} onClose={() => setSelected(null)} />
    </>
  );
}

function ShiftDetailModal({ id, onClose }: { id: number | null; onClose: () => void }) {
  const query = useShift(id);
  const s = query.data;
  const row = (label: string, value: ReactNode) => (
    <div className="flex items-center justify-between py-2">
      <span className="text-subheadline text-text-secondary">{label}</span>
      <span className="text-subheadline tabular-nums">{value}</span>
    </div>
  );
  return (
    <Modal open={!!id} onClose={onClose} title="Detail Shift" description={s ? `${s.user.name} · ${formatDateTime(s.opened_at)}` : undefined} size="sm">
      {query.isLoading ? (
        <Skeleton className="h-64 w-full" />
      ) : query.isError || !s ? (
        <ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} />
      ) : (
        <div className="divide-y divide-separator/50">
          <div className="pb-2">
            {row('Kas awal', formatRupiah(s.opening_cash))}
            {row('Penjualan tunai', formatRupiah(s.summary.cash_sales))}
            {row('Kas seharusnya', <b>{formatRupiah(s.expected_cash ?? s.summary.expected_cash)}</b>)}
            {s.actual_cash !== null && row('Kas aktual', formatRupiah(s.actual_cash))}
            {s.status === 'CLOSED' && row('Selisih', <Difference value={s.cash_difference} />)}
          </div>
          <div className="py-2">
            {row('Total penjualan', <b>{formatRupiah(s.summary.total_sales)}</b>)}
            {row('Jumlah transaksi', s.summary.total_transactions)}
            {row('Transaksi dibatalkan', s.summary.voided_count)}
            {Object.entries(s.summary.payment_totals).map(([m, v]) => row(paymentMethodLabel[m] ?? m, formatRupiah(v ?? 0)))}
          </div>
          {(s.opening_note || s.closing_note) && (
            <div className="pt-2 text-footnote text-text-secondary">
              {s.opening_note && <p>Catatan buka: {s.opening_note}</p>}
              {s.closing_note && <p>Catatan tutup: {s.closing_note}</p>}
            </div>
          )}
        </div>
      )}
    </Modal>
  );
}
