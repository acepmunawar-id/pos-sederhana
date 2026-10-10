import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button, Input, Modal, useToast } from '@/components/ui';
import { applyServerErrors } from '@/lib/form';
import { useSaveUnit } from '../hooks/useCategories';
import type { Unit } from '../types';

const schema = z.object({
  name: z.string().trim().min(1, 'Wajib diisi').max(50, 'Maksimal 50 karakter'),
  abbreviation: z.string().trim().min(1, 'Wajib diisi').max(10, 'Maksimal 10 karakter'),
});
type FormValues = z.infer<typeof schema>;

export function UnitFormModal({ open, onClose, unit }: { open: boolean; onClose: () => void; unit?: Unit | null }) {
  const toast = useToast();
  const save = useSaveUnit();
  const { register, handleSubmit, reset, setError, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema) });

  useEffect(() => {
    if (open) reset({ name: unit?.name ?? '', abbreviation: unit?.abbreviation ?? '' });
  }, [open, unit, reset]);

  const onSubmit = handleSubmit((v) =>
    save.mutate(
      { id: unit?.id, input: v },
      {
        onSuccess: () => {
          toast.success(unit ? 'Satuan diperbarui' : 'Satuan ditambahkan');
          onClose();
        },
        onError: (err) => {
          const msg = applyServerErrors(err, setError, ['name', 'abbreviation']);
          if (msg) toast.error(msg);
        },
      },
    ),
  );

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={unit ? 'Ubah Satuan' : 'Tambah Satuan'}
      size="sm"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Batal</Button>
          <Button type="submit" form="unit-form" loading={save.isPending}>Simpan</Button>
        </>
      }
    >
      <form id="unit-form" onSubmit={onSubmit} className="space-y-4" noValidate>
        <Input label="Nama satuan" placeholder="Kilogram" required autoFocus error={errors.name?.message} {...register('name')} />
        <Input label="Singkatan" placeholder="kg" required error={errors.abbreviation?.message} {...register('abbreviation')} />
      </form>
    </Modal>
  );
}
