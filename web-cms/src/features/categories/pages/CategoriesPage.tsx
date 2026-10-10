import { useState } from 'react';
import { Plus, Tags, Ruler } from 'lucide-react';
import { PageHeader } from '@/components/layout/PageHeader';
import {
  Button,
  Card,
  ConfirmDialog,
  EmptyState,
  ErrorState,
  Pagination,
  SearchInput,
  SegmentedControl,
  Table,
  RowActions,
  useToast,
  type Column,
} from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { formatDate } from '@/lib/format';
import { useListState } from '@/lib/use-list-state';
import { CategoryFormModal } from '../components/CategoryFormModal';
import { UnitFormModal } from '../components/UnitFormModal';
import { useCategories, useDeleteCategory, useDeleteUnit, useUnits } from '../hooks/useCategories';
import type { Category, Unit } from '../types';

type Tab = 'categories' | 'units';

export default function CategoriesPage() {
  const [tab, setTab] = useState<Tab>('categories');
  return (
    <>
      <PageHeader
        title="Kategori & Satuan"
        description="Kelompokkan produk dan tentukan satuan jualnya."
        actions={
          <SegmentedControl
            value={tab}
            onChange={setTab}
            segments={[
              { value: 'categories', label: 'Kategori' },
              { value: 'units', label: 'Satuan' },
            ]}
          />
        }
      />
      {tab === 'categories' ? <CategoriesTab /> : <UnitsTab />}
    </>
  );
}

function CategoriesTab() {
  const toast = useToast();
  const list = useListState({}, { sort: 'name', order: 'asc' });
  const query = useCategories(list.params);
  const remove = useDeleteCategory();
  const [editing, setEditing] = useState<Category | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [deleting, setDeleting] = useState<Category | null>(null);

  const columns: Column<Category>[] = [
    { key: 'name', header: 'Nama', sortKey: 'name', render: (c) => <span className="font-semibold">{c.name}</span> },
    { key: 'desc', header: 'Deskripsi', render: (c) => <span className="text-text-secondary">{c.description || '-'}</span>, hideOnMobile: true },
    { key: 'count', header: 'Jumlah Produk', align: 'right', render: (c) => c.product_count },
    { key: 'created', header: 'Dibuat', sortKey: 'created_at', render: (c) => formatDate(c.created_at), hideOnMobile: true },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (c) => (
        <RowActions
          onEdit={() => {
            setEditing(c);
            setFormOpen(true);
          }}
          onDelete={() => setDeleting(c)}
        />
      ),
    },
  ];

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  return (
    <Card padded={false}>
      <div className="flex flex-wrap items-center justify-between gap-3 p-4">
        <SearchInput value={list.search} onChange={list.setSearch} placeholder="Cari kategori…" />
        <Button icon={<Plus className="h-4 w-4" />} onClick={openCreate}>Tambah Kategori</Button>
      </div>
      {query.isError ? (
        <ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} />
      ) : (
        <Table
          columns={columns}
          rows={query.data?.items}
          rowKey={(c) => c.id}
          loading={query.isLoading}
          sort={list.sort}
          onSort={list.toggleSort}
          empty={
            <EmptyState
              icon={<Tags className="h-7 w-7" />}
              title={list.search ? 'Kategori tidak ditemukan' : 'Belum ada kategori'}
              description={list.search ? 'Coba kata kunci lain.' : 'Tambahkan kategori pertama untuk mengelompokkan produk.'}
              action={!list.search && <Button onClick={openCreate}>Tambah Kategori</Button>}
            />
          }
        />
      )}
      <Pagination meta={query.data?.meta} onChange={list.setPage} />

      <CategoryFormModal open={formOpen} onClose={() => setFormOpen(false)} category={editing} />
      <ConfirmDialog
        open={!!deleting}
        title="Hapus kategori?"
        message={<>Kategori <b>{deleting?.name}</b> akan dihapus. Kategori yang masih dipakai produk tidak dapat dihapus.</>}
        confirmLabel="Hapus"
        loading={remove.isPending}
        onClose={() => setDeleting(null)}
        onConfirm={() =>
          remove.mutate(deleting!.id, {
            onSuccess: () => {
              toast.success('Kategori dihapus');
              setDeleting(null);
            },
            onError: (err) => toast.error(errorMessage(err)),
          })
        }
      />
    </Card>
  );
}

function UnitsTab() {
  const toast = useToast();
  const list = useListState({}, { sort: 'name', order: 'asc' });
  const query = useUnits(list.params);
  const remove = useDeleteUnit();
  const [editing, setEditing] = useState<Unit | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [deleting, setDeleting] = useState<Unit | null>(null);

  const columns: Column<Unit>[] = [
    { key: 'name', header: 'Nama', sortKey: 'name', render: (u) => <span className="font-semibold">{u.name}</span> },
    { key: 'abbr', header: 'Singkatan', sortKey: 'abbreviation', render: (u) => u.abbreviation },
    { key: 'created', header: 'Dibuat', sortKey: 'created_at', render: (u) => formatDate(u.created_at), hideOnMobile: true },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (u) => (
        <RowActions
          onEdit={() => {
            setEditing(u);
            setFormOpen(true);
          }}
          onDelete={() => setDeleting(u)}
        />
      ),
    },
  ];

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  return (
    <Card padded={false}>
      <div className="flex flex-wrap items-center justify-between gap-3 p-4">
        <SearchInput value={list.search} onChange={list.setSearch} placeholder="Cari satuan…" />
        <Button icon={<Plus className="h-4 w-4" />} onClick={openCreate}>Tambah Satuan</Button>
      </div>
      {query.isError ? (
        <ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} />
      ) : (
        <Table
          columns={columns}
          rows={query.data?.items}
          rowKey={(u) => u.id}
          loading={query.isLoading}
          sort={list.sort}
          onSort={list.toggleSort}
          empty={
            <EmptyState
              icon={<Ruler className="h-7 w-7" />}
              title={list.search ? 'Satuan tidak ditemukan' : 'Belum ada satuan'}
              action={!list.search && <Button onClick={openCreate}>Tambah Satuan</Button>}
            />
          }
        />
      )}
      <Pagination meta={query.data?.meta} onChange={list.setPage} />

      <UnitFormModal open={formOpen} onClose={() => setFormOpen(false)} unit={editing} />
      <ConfirmDialog
        open={!!deleting}
        title="Hapus satuan?"
        message={<>Satuan <b>{deleting?.name}</b> akan dihapus. Satuan yang masih dipakai produk tidak dapat dihapus.</>}
        confirmLabel="Hapus"
        loading={remove.isPending}
        onClose={() => setDeleting(null)}
        onConfirm={() =>
          remove.mutate(deleting!.id, {
            onSuccess: () => {
              toast.success('Satuan dihapus');
              setDeleting(null);
            },
            onError: (err) => toast.error(errorMessage(err)),
          })
        }
      />
    </Card>
  );
}
