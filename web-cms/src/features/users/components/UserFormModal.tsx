import { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button, Input, Modal, Select, Switch, useToast } from '@/components/ui';
import { applyServerErrors, numberField } from '@/lib/form';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { useRoles, useSaveUser } from '../hooks/useUsers';
import type { User } from '../types';

const makeSchema = (isEdit: boolean) =>
  z.object({
    name: z.string().trim().min(2, 'Minimal 2 karakter').max(100),
    email: z.string().trim().min(1, 'Email wajib diisi').email('Format email tidak valid'),
    role_id: numberField('Pilih role'),
    password: isEdit
      ? z.string().max(72).refine((v) => v === '' || v.length >= 8, 'Minimal 8 karakter').optional()
      : z.string().min(8, 'Minimal 8 karakter').max(72),
    is_active: z.boolean(),
  });
type FormValues = z.infer<ReturnType<typeof makeSchema>>;

export function UserFormModal({ open, onClose, user }: { open: boolean; onClose: () => void; user?: User | null }) {
  const toast = useToast();
  const { user: me, hasRole } = useAuth();
  const roles = useRoles();
  const save = useSaveUser();
  const isEdit = !!user;
  const isSelf = user?.id === me?.id;
  const { register, handleSubmit, reset, setError, control, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(makeSchema(isEdit)),
  });

  useEffect(() => {
    if (!open) return;
    reset({
      name: user?.name ?? '',
      email: user?.email ?? '',
      role_id: user?.role.id ?? ('' as unknown as number),
      password: '',
      is_active: user?.is_active ?? true,
    });
  }, [open, user, reset]);

  // Only an Owner may grant the Owner role.
  const roleOptions = (roles.data ?? [])
    .filter((r) => hasRole('owner') || r.code !== 'owner')
    .map((r) => ({ value: r.id, label: r.name }));

  const onSubmit = handleSubmit((v) => {
    const payload = { name: v.name, email: v.email, role_id: v.role_id, is_active: v.is_active };
    const req = user
      ? { id: user.id, input: { ...payload, password: v.password || undefined } }
      : { input: { ...payload, password: v.password ?? '' } };
    save.mutate(req, {
      onSuccess: () => {
        toast.success(isEdit ? 'Pengguna diperbarui' : 'Pengguna ditambahkan');
        onClose();
      },
      onError: (err) => {
        const msg = applyServerErrors(err, setError, ['name', 'email', 'role_id', 'password']);
        if (msg) toast.error(msg);
      },
    });
  });

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isEdit ? 'Ubah Pengguna' : 'Tambah Pengguna'}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Batal</Button>
          <Button type="submit" form="user-form" loading={save.isPending}>Simpan</Button>
        </>
      }
    >
      <form id="user-form" onSubmit={onSubmit} className="space-y-4" noValidate>
        <Input label="Nama" required autoFocus error={errors.name?.message} {...register('name')} />
        <Input label="Email" type="email" required autoComplete="off" error={errors.email?.message} {...register('email')} />
        <Select
          label="Role"
          required
          disabled={isSelf}
          hint={isSelf ? 'Anda tidak dapat mengubah role akun sendiri' : undefined}
          placeholder="Pilih role"
          options={roleOptions}
          error={errors.role_id?.message}
          {...register('role_id')}
        />
        <Input
          label={isEdit ? 'Password baru' : 'Password'}
          type="password"
          autoComplete="new-password"
          required={!isEdit}
          hint={isEdit ? 'Kosongkan jika tidak ingin mengganti. Pengguna akan logout dari semua perangkat.' : 'Minimal 8 karakter'}
          error={errors.password?.message}
          {...register('password')}
        />
        {!isSelf && (
          <Controller
            control={control}
            name="is_active"
            render={({ field }) => (
              <Switch checked={field.value} onChange={field.onChange} label="Akun aktif" description="Akun nonaktif tidak bisa login" />
            )}
          />
        )}
      </form>
    </Modal>
  );
}
