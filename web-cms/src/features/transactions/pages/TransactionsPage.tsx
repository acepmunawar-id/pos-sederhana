import { useNavigate } from 'react-router-dom';
import { Receipt } from 'lucide-react';
import { PageHeader } from '@/components/layout/PageHeader';
import {
  Card,
  EmptyState,
  ErrorState,
  Input,
  Pagination,
  SearchInput,
  Select,
  Table,
  type Column,
} from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { formatDateTime, formatRupiah, paymentMethodLabel } from '@/lib/format';
import { useListState } from '@/lib/use-list-state';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { useUserOptions } from '@/features/users/hooks/useUsers';
import { TransactionStatusBadge } from '../components/StatusBadge';
import { useTransactions } from '../hooks/useTransactions';
import type { Transaction } from '../types';

export default function TransactionsPage() {
  const navigate = useNavigate();
  const { hasRole } = useAuth();
  const isManager = hasRole('owner', 'admin');
  const users = useUserOptions();
  const list = useListState<{ status?: string; method?: string; user_id?: number; date_from?: string; date_to?: string }>(
    {},
    { sort: 'created_at', order: 'desc' },
  );
  const query = useTransactions(list.params);

  const columns: Column<Transaction>[] = [
    {
      key: 'invoice',
      header: 'No. Invoice',
      sortKey: 'invoice_number',
      render: (t) => <span className="font-semibold tabular-nums text-primary">{t.invoice_number}</span>,
    },
    { key: 'date', header: 'Waktu', sortKey: 'created_at', render: (t) => <span className="whitespace-nowrap">{formatDateTime(t.created_at)}</span> },
    { key: 'cashier', header: 'Kasir', sortKey: 'cashier', render: (t) => t.cashier.name, hideOnMobile: true },
    {
      key: 'method',
      header: 'Pembayaran',
      render: (t) => t.payments.map((p) => paymentMethodLabel[p.method]).join(' + '),
      hideOnMobile: true,
    },
    { key: 'total', header: 'Total', sortKey: 'total', align: 'right', render: (t) => <span className="font-semibold tabular-nums">{formatRupiah(t.total)}</span> },
    { key: 'status', header: 'Status', render: (t) => <TransactionStatusBadge status={t.status} /> },
  ];

  return (
    <>
      <PageHeader title="Transaksi" description={isManager ? 'Semua transaksi penjualan toko.' : 'Transaksi yang Anda buat.'} />
      <Card padded={false}>
        <div className="flex flex-wrap items-end gap-3 p-4">
          <SearchInput value={list.search} onChange={list.setSearch} placeholder="Cari nomor invoice…" />
          <Select
            className="w-full sm:w-40"
            aria-label="Status"
            placeholder="Semua status"
            value={list.filters.status ?? ''}
            onChange={(e) => list.setFilter('status', e.target.value || undefined)}
            options={[{ value: 'COMPLETED', label: 'Selesai' }, { value: 'VOIDED', label: 'Dibatalkan' }]}
          />
          <Select
            className="w-full sm:w-40"
            aria-label="Metode bayar"
            placeholder="Semua metode"
            value={list.filters.method ?? ''}
            onChange={(e) => list.setFilter('method', e.target.value || undefined)}
            options={Object.entries(paymentMethodLabel).map(([value, label]) => ({ value, label }))}
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
            rowKey={(t) => t.id}
            loading={query.isLoading}
            sort={list.sort}
            onSort={list.toggleSort}
            onRowClick={(t) => navigate(`/transactions/${t.id}`)}
            empty={<EmptyState icon={<Receipt className="h-7 w-7" />} title="Belum ada transaksi" description="Transaksi dari aplikasi kasir akan muncul di sini." />}
          />
        )}
        <Pagination meta={query.data?.meta} onChange={list.setPage} />
      </Card>
    </>
  );
}
