import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation } from '@tanstack/react-query';
import { Button, Input, Modal, useToast } from '@/components/ui';
import { applyServerErrors } from '@/lib/form';
import { authApi } from '../api';
import { useAuth } from '../hooks/useAuth';

const schema = z
  .object({
    old_password: z.string().min(1, 'Password lama wajib diisi'),
    new_password: z.string().min(8, 'Minimal 8 karakter').max(72, 'Maksimal 72 karakter'),
    confirm: z.string(),
  })
  .refine((v) => v.new_password === v.confirm, { path: ['confirm'], message: 'Konfirmasi password tidak cocok' })
  .refine((v) => v.new_password !== v.old_password, { path: ['new_password'], message: 'Harus berbeda dari password lama' });
type FormValues = z.infer<typeof schema>;

export function ChangePasswordModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const { logout } = useAuth();
  const { register, handleSubmit, setError, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { old_password: '', new_password: '', confirm: '' },
  });

  const mutation = useMutation({
    mutationFn: (v: FormValues) => authApi.changePassword({ old_password: v.old_password, new_password: v.new_password }),
    onSuccess: async () => {
      toast.success('Password berhasil diganti. Silakan login kembali.');
      reset();
      onClose();
      await logout();
    },
    onError: (err) => {
      const msg = applyServerErrors(err, setError, ['old_password', 'new_password']);
      if (msg) toast.error(msg);
    },
  });

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Ganti Password"
      description="Setelah diganti, Anda akan keluar dari semua perangkat."
      size="sm"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Batal
          </Button>
          <Button form="change-password-form" type="submit" loading={mutation.isPending}>
            Simpan
          </Button>
        </>
      }
    >
      <form id="change-password-form" className="space-y-4" onSubmit={handleSubmit((v) => mutation.mutate(v))} noValidate>
        <Input label="Password lama" type="password" autoComplete="current-password" error={errors.old_password?.message} {...register('old_password')} />
        <Input label="Password baru" type="password" autoComplete="new-password" hint="Minimal 8 karakter" error={errors.new_password?.message} {...register('new_password')} />
        <Input label="Ulangi password baru" type="password" autoComplete="new-password" error={errors.confirm?.message} {...register('confirm')} />
      </form>
    </Modal>
  );
}
