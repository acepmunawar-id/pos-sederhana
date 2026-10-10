import { useState } from 'react';
import { Package, Plus } from 'lucide-react';
import { PageHeader } from '@/components/layout/PageHeader';
import {
  Badge,
  Button,
  Card,
  ConfirmDialog,
  EmptyState,
  ErrorState,
  Pagination,
  RowActions,
  SearchInput,
  SegmentedControl,
  Select,
  Table,
  useToast,
  type Column,
} from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { formatRupiah } from '@/lib/format';
import { useListState } from '@/lib/use-list-state';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { useCategoryOptions } from '@/features/categories/hooks/useCategories';
import { ProductFormModal } from '../components/ProductFormModal';
import { ProductThumb } from '../components/ProductThumb';
import { useDeleteProduct, useProducts } from '../hooks/useProducts';
import type { Product } from '../types';

type StatusFilter = 'all' | 'active' | 'inactive' | 'low';

export default function ProductsPage() {
  const toast = useToast();
  const { hasRole } = useAuth();
  const canManage = hasRole('owner', 'admin');
  const categories = useCategoryOptions();
  const [status, setStatus] = useState<StatusFilter>('all');
  const list = useListState<{ category_id?: number; is_active?: boolean; low_stock?: boolean }>({}, { sort: 'name', order: 'asc' });
  const query = useProducts(list.params);
  const remove = useDeleteProduct();
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Product | null>(null);
  const [deleting, setDeleting] = useState<Product | null>(null);

  const changeStatus = (s: StatusFilter) => {
    setStatus(s);
    list.setFilter('is_active', s === 'active' ? true : s === 'inactive' ? false : undefined);
    list.setFilter('low_stock', s === 'low' ? true : undefined);
  };

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const columns: Column<Product>[] = [
    {
      key: 'name',
      header: 'Produk',
      sortKey: 'name',
      render: (p) => (
        <div className="flex items-center gap-3">
          <ProductThumb url={p.image_url} name={p.name} />
          <div className="min-w-0">
            <p className="truncate font-semibold">{p.name}</p>
            <p className="text-caption text-text-secondary">
              {p.sku}
              {p.barcode && ` · ${p.barcode}`}
            </p>
          </div>
        </div>
      ),
    },
    { key: 'category', header: 'Kategori', sortKey: 'category', render: (p) => p.category.name, hideOnMobile: true },
    { key: 'price', header: 'Harga Jual', sortKey: 'price', align: 'right', render: (p) => <span className="tabular-nums">{formatRupiah(p.price)}</span> },
    ...(canManage
      ? [{ key: 'cost', header: 'Modal', sortKey: 'cost', align: 'right' as const, hideOnMobile: true, render: (p: Product) => <span className="tabular-nums text-text-secondary">{formatRupiah(p.cost)}</span> }]
      : []),
    {
      key: 'stock',
      header: 'Stok',
      sortKey: 'stock',
      align: 'right',
      render: (p) => (
        <span className={p.is_low_stock ? 'font-semibold text-danger' : 'tabular-nums'}>
          {p.stock} {p.unit.name}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (p) => (
        <div className="flex flex-wrap gap-1">
          {p.is_active ? <Badge tone="success">Aktif</Badge> : <Badge>Nonaktif</Badge>}
          {p.is_low_stock && <Badge tone="warning">Stok menipis</Badge>}
        </div>
      ),
    },
    ...(canManage
      ? [{
          key: 'actions',
          header: '',
          align: 'right' as const,
          render: (p: Product) => (
            <RowActions
              onEdit={() => {
                setEditing(p);
                setFormOpen(true);
              }}
              onDelete={() => setDeleting(p)}
            />
          ),
        }]
      : []),
  ];

  return (
    <>
      <PageHeader
        title="Produk"
        description="Kelola daftar produk, harga, dan stok minimum."
        actions={canManage && <Button icon={<Plus className="h-4 w-4" />} onClick={openCreate}>Tambah Produk</Button>}
      />
      <Card padded={false}>
        <div className="flex flex-wrap items-center gap-3 p-4">
          <SearchInput value={list.search} onChange={list.setSearch} placeholder="Cari nama, SKU, barcode…" />
          <Select
            className="w-full sm:w-52"
            aria-label="Filter kategori"
            placeholder="Semua kategori"
            value={list.filters.category_id ?? ''}
            onChange={(e) => list.setFilter('category_id', e.target.value ? Number(e.target.value) : undefined)}
            options={(categories.data?.items ?? []).map((c) => ({ value: c.id, label: c.name }))}
          />
          <SegmentedControl
            value={status}
            onChange={changeStatus}
            segments={[
              { value: 'all', label: 'Semua' },
              { value: 'active', label: 'Aktif' },
              { value: 'inactive', label: 'Nonaktif' },
              { value: 'low', label: 'Stok menipis' },
            ]}
          />
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
            empty={
              <EmptyState
                icon={<Package className="h-7 w-7" />}
                title={list.search || status !== 'all' || list.filters.category_id ? 'Produk tidak ditemukan' : 'Belum ada produk'}
                description={list.search ? 'Coba kata kunci atau filter lain.' : undefined}
                action={canManage && !list.search && status === 'all' && <Button onClick={openCreate}>Tambah Produk</Button>}
              />
            }
          />
        )}
        <Pagination meta={query.data?.meta} onChange={list.setPage} />
      </Card>

      <ProductFormModal open={formOpen} onClose={() => setFormOpen(false)} product={editing} />
      <ConfirmDialog
        open={!!deleting}
        title="Hapus produk?"
        message={<>Produk <b>{deleting?.name}</b> akan dihapus dari daftar. Riwayat transaksi tetap tersimpan.</>}
        confirmLabel="Hapus"
        loading={remove.isPending}
        onClose={() => setDeleting(null)}
        onConfirm={() =>
          remove.mutate(deleting!.id, {
            onSuccess: () => {
              toast.success('Produk dihapus');
              setDeleting(null);
            },
            onError: (err) => toast.error(errorMessage(err)),
          })
        }
      />
    </>
  );
}
