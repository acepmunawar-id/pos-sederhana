import { useState } from 'react';
import { Plus, Users } from 'lucide-react';
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
  Table,
  useToast,
  type Column,
} from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { formatDateTime } from '@/lib/format';
import { useListState } from '@/lib/use-list-state';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { UserFormModal } from '../components/UserFormModal';
import { useDeleteUser, useUsers } from '../hooks/useUsers';
import type { User } from '../types';

const roleTone = { owner: 'primary', admin: 'warning', cashier: 'gray' } as const;

export default function UsersPage() {
  const toast = useToast();
  const { user: me, hasRole } = useAuth();
  const list = useListState<{ role?: string }>({}, { sort: 'name', order: 'asc' });
  const query = useUsers(list.params);
  const remove = useDeleteUser();
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [deleting, setDeleting] = useState<User | null>(null);
  const [role, setRole] = useState('');

  // Admins cannot manage Owner accounts (enforced by the API as well).
  const canManage = (u: User) => hasRole('owner') || u.role.code !== 'owner';

  const columns: Column<User>[] = [
    {
      key: 'name',
      header: 'Nama',
      sortKey: 'name',
      render: (u) => (
        <div>
          <p className="font-semibold">
            {u.name} {u.id === me?.id && <span className="text-caption font-normal text-text-secondary">(Anda)</span>}
          </p>
          <p className="text-caption text-text-secondary">{u.email}</p>
        </div>
      ),
    },
    { key: 'role', header: 'Role', render: (u) => <Badge tone={roleTone[u.role.code]}>{u.role.name}</Badge> },
    { key: 'status', header: 'Status', render: (u) => (u.is_active ? <Badge tone="success">Aktif</Badge> : <Badge>Nonaktif</Badge>) },
    { key: 'login', header: 'Login terakhir', sortKey: 'last_login_at', render: (u) => formatDateTime(u.last_login_at), hideOnMobile: true },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (u) =>
        canManage(u) && (
          <RowActions
            onEdit={() => {
              setEditing(u);
              setFormOpen(true);
            }}
            onDelete={u.id !== me?.id ? () => setDeleting(u) : undefined}
          />
        ),
    },
  ];

  return (
    <>
      <PageHeader
        title="Pengguna"
        description="Kelola akun Owner, Admin, dan Kasir."
        actions={<Button icon={<Plus className="h-4 w-4" />} onClick={() => { setEditing(null); setFormOpen(true); }}>Tambah Pengguna</Button>}
      />
      <Card padded={false}>
        <div className="flex flex-wrap items-center gap-3 p-4">
          <SearchInput value={list.search} onChange={list.setSearch} placeholder="Cari nama atau email…" />
          <SegmentedControl
            value={role}
            onChange={(v) => {
              setRole(v);
              list.setFilter('role', v || undefined);
            }}
            segments={[
              { value: '', label: 'Semua' },
              { value: 'owner', label: 'Owner' },
              { value: 'admin', label: 'Admin' },
              { value: 'cashier', label: 'Kasir' },
            ]}
          />
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
            empty={<EmptyState icon={<Users className="h-7 w-7" />} title="Pengguna tidak ditemukan" />}
          />
        )}
        <Pagination meta={query.data?.meta} onChange={list.setPage} />
      </Card>

      <UserFormModal open={formOpen} onClose={() => setFormOpen(false)} user={editing} />
      <ConfirmDialog
        open={!!deleting}
        title="Hapus pengguna?"
        message={<>Akun <b>{deleting?.name}</b> akan dihapus dan langsung logout dari semua perangkat. Riwayat transaksinya tetap tersimpan.</>}
        confirmLabel="Hapus"
        loading={remove.isPending}
        onClose={() => setDeleting(null)}
        onConfirm={() =>
          remove.mutate(deleting!.id, {
            onSuccess: () => {
              toast.success('Pengguna dihapus');
              setDeleting(null);
            },
            onError: (err) => toast.error(errorMessage(err)),
          })
        }
      />
    </>
  );
}
