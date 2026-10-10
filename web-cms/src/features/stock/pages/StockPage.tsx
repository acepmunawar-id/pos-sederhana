import { useState } from 'react';
import { AlertTriangle, ClipboardCheck, History, PackagePlus } from 'lucide-react';
import { PageHeader } from '@/components/layout/PageHeader';
import {
  Badge,
  Button,
  Card,
  EmptyState,
  ErrorState,
  Input,
  Pagination,
  SearchInput,
  SegmentedControl,
  Select,
  Table,
  type Column,
} from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { cn } from '@/lib/cn';
import { formatDateTime } from '@/lib/format';
import { useListState } from '@/lib/use-list-state';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { ProductThumb } from '@/features/products/components/ProductThumb';
import type { Product } from '@/features/products/types';
import { StockAdjustmentModal } from '../components/StockAdjustmentModal';
import { StockInModal } from '../components/StockInModal';
import { useLowStock, useStockMovements } from '../hooks/useStock';
import { movementLabel, type MovementType, type StockMovement } from '../types';

type Tab = 'history' | 'low';

const typeTone: Record<MovementType, 'success' | 'primary' | 'gray' | 'warning'> = {
  IN: 'success',
  ADJUSTMENT: 'warning',
  SALE: 'primary',
  VOID: 'gray',
};

export default function StockPage() {
  const { hasRole } = useAuth();
  const canManage = hasRole('owner', 'admin');
  const [tab, setTab] = useState<Tab>('history');
  const [stockInOpen, setStockInOpen] = useState(false);
  const [adjustOpen, setAdjustOpen] = useState(false);
  const [adjustProduct, setAdjustProduct] = useState<Product | null>(null);

  return (
    <>
      <PageHeader
        title="Stok"
        description="Pantau pergerakan stok dan produk yang hampir habis."
        actions={
          canManage && (
            <>
              <Button variant="secondary" icon={<ClipboardCheck className="h-4 w-4" />} onClick={() => { setAdjustProduct(null); setAdjustOpen(true); }}>
                Penyesuaian
              </Button>
              <Button icon={<PackagePlus className="h-4 w-4" />} onClick={() => setStockInOpen(true)}>
                Stok Masuk
              </Button>
            </>
          )
        }
      />
      <SegmentedControl
        className="mb-4"
        value={tab}
        onChange={setTab}
        segments={[
          { value: 'history', label: 'Riwayat Pergerakan' },
          { value: 'low', label: 'Stok Menipis' },
        ]}
      />
      {tab === 'history' ? (
        <MovementsTab />
      ) : (
        <LowStockTab canManage={canManage} onAdjust={(p) => { setAdjustProduct(p); setAdjustOpen(true); }} />
      )}
      <StockInModal open={stockInOpen} onClose={() => setStockInOpen(false)} />
      <StockAdjustmentModal open={adjustOpen} onClose={() => setAdjustOpen(false)} initialProduct={adjustProduct} />
    </>
  );
}

function MovementsTab() {
  const list = useListState<{ type?: string; date_from?: string; date_to?: string }>({}, { sort: 'created_at', order: 'desc' });
  const query = useStockMovements(list.params);

  const columns: Column<StockMovement>[] = [
    { key: 'date', header: 'Waktu', sortKey: 'created_at', render: (m) => <span className="whitespace-nowrap">{formatDateTime(m.created_at)}</span> },
    {
      key: 'product',
      header: 'Produk',
      sortKey: 'product',
      render: (m) => (
        <div>
          <p className="font-semibold">{m.product.name}</p>
          <p className="text-caption text-text-secondary">{m.product.sku}</p>
        </div>
      ),
    },
    { key: 'type', header: 'Jenis', sortKey: 'type', render: (m) => <Badge tone={typeTone[m.type]}>{movementLabel[m.type]}</Badge> },
    {
      key: 'qty',
      header: 'Jumlah',
      sortKey: 'quantity',
      align: 'right',
      render: (m) => (
        <span className={cn('font-semibold tabular-nums', m.quantity > 0 ? 'text-success' : 'text-danger')}>
          {m.quantity > 0 ? '+' : ''}
          {m.quantity}
        </span>
      ),
    },
    { key: 'flow', header: 'Stok', align: 'right', render: (m) => <span className="tabular-nums text-text-secondary">{m.stock_before} → {m.stock_after}</span> },
    { key: 'user', header: 'Oleh', render: (m) => m.user.name, hideOnMobile: true },
    { key: 'note', header: 'Catatan', render: (m) => <span className="text-text-secondary">{m.note ?? '-'}</span>, hideOnMobile: true },
  ];

  return (
    <Card padded={false}>
      <div className="flex flex-wrap items-end gap-3 p-4">
        <SearchInput value={list.search} onChange={list.setSearch} placeholder="Cari produk atau catatan…" />
        <Select
          className="w-full sm:w-44"
          aria-label="Jenis pergerakan"
          placeholder="Semua jenis"
          value={list.filters.type ?? ''}
          onChange={(e) => list.setFilter('type', e.target.value || undefined)}
          options={(Object.keys(movementLabel) as MovementType[]).map((t) => ({ value: t, label: movementLabel[t] }))}
        />
        <Input className="w-full sm:w-44" type="date" aria-label="Dari tanggal" value={list.filters.date_from ?? ''} onChange={(e) => list.setFilter('date_from', e.target.value || undefined)} />
        <Input className="w-full sm:w-44" type="date" aria-label="Sampai tanggal" value={list.filters.date_to ?? ''} onChange={(e) => list.setFilter('date_to', e.target.value || undefined)} />
      </div>
      {query.isError ? (
        <ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} />
      ) : (
        <Table
          columns={columns}
          rows={query.data?.items}
          rowKey={(m) => m.id}
          loading={query.isLoading}
          sort={list.sort}
          onSort={list.toggleSort}
          empty={<EmptyState icon={<History className="h-7 w-7" />} title="Belum ada pergerakan stok" description="Coba ubah filter atau rentang tanggal." />}
        />
      )}
      <Pagination meta={query.data?.meta} onChange={list.setPage} />
    </Card>
  );
}

function LowStockTab({ canManage, onAdjust }: { canManage: boolean; onAdjust: (p: Product) => void }) {
  const list = useListState({}, { sort: 'stock', order: 'asc' });
  const query = useLowStock(list.params);

  const columns: Column<Product>[] = [
    {
      key: 'name',
      header: 'Produk',
      render: (p) => (
        <div className="flex items-center gap-3">
          <ProductThumb url={p.image_url} name={p.name} />
          <div>
            <p className="font-semibold">{p.name}</p>
            <p className="text-caption text-text-secondary">{p.category.name}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'stock',
      header: 'Stok',
      align: 'right',
      sortKey: 'stock',
      render: (p) => <span className={cn('font-semibold tabular-nums', p.stock === 0 ? 'text-danger' : 'text-warning')}>{p.stock} {p.unit.name}</span>,
    },
    { key: 'min', header: 'Minimum', align: 'right', render: (p) => <span className="tabular-nums">{p.min_stock}</span> },
    { key: 'status', header: 'Status', render: (p) => (p.stock === 0 ? <Badge tone="danger">Habis</Badge> : <Badge tone="warning">Menipis</Badge>) },
    ...(canManage
      ? [{
          key: 'action',
          header: '',
          align: 'right' as const,
          render: (p: Product) => <Button size="sm" variant="secondary" onClick={() => onAdjust(p)}>Sesuaikan</Button>,
        }]
      : []),
  ];

  return (
    <Card padded={false}>
      <div className="p-4">
        <SearchInput value={list.search} onChange={list.setSearch} placeholder="Cari produk…" />
      </div>
      {query.isError ? (
        <ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} />
      ) : (
        <Table
          columns={columns}
          rows={query.data?.items}
          rowKey={(p) => p.id}
          loading={query.isLoading}
          sort={list.sort}
          onSort={list.toggleSort}
          empty={<EmptyState icon={<AlertTriangle className="h-7 w-7" />} title="Semua stok aman" description="Tidak ada produk yang berada di bawah stok minimum." />}
        />
      )}
      <Pagination meta={query.data?.meta} onChange={list.setPage} />
    </Card>
  );
}
