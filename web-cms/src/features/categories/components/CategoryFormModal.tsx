import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button, Input, Modal, Textarea, useToast } from '@/components/ui';
import { applyServerErrors } from '@/lib/form';
import { useSaveCategory } from '../hooks/useCategories';
import type { Category } from '../types';

const schema = z.object({
  name: z.string().trim().min(2, 'Minimal 2 karakter').max(100, 'Maksimal 100 karakter'),
  description: z.string().max(500, 'Maksimal 500 karakter').optional(),
});
type FormValues = z.infer<typeof schema>;

export function CategoryFormModal({ open, onClose, category }: { open: boolean; onClose: () => void; category?: Category | null }) {
  const toast = useToast();
  const save = useSaveCategory();
  const { register, handleSubmit, reset, setError, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema) });

  useEffect(() => {
    if (open) reset({ name: category?.name ?? '', description: category?.description ?? '' });
  }, [open, category, reset]);

  const onSubmit = handleSubmit((v) =>
    save.mutate(
      { id: category?.id, input: { name: v.name, description: v.description || null } },
      {
        onSuccess: () => {
          toast.success(category ? 'Kategori diperbarui' : 'Kategori ditambahkan');
          onClose();
        },
        onError: (err) => {
          const msg = applyServerErrors(err, setError, ['name', 'description']);
          if (msg) toast.error(msg);
        },
      },
    ),
  );

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={category ? 'Ubah Kategori' : 'Tambah Kategori'}
      size="sm"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Batal</Button>
          <Button type="submit" form="category-form" loading={save.isPending}>Simpan</Button>
        </>
      }
    >
      <form id="category-form" onSubmit={onSubmit} className="space-y-4" noValidate>
        <Input label="Nama kategori" required autoFocus error={errors.name?.message} {...register('name')} />
        <Textarea label="Deskripsi" error={errors.description?.message} {...register('description')} />
      </form>
    </Modal>
  );
}
