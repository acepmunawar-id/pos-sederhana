import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { PageHeader } from '@/components/layout/PageHeader';
import { Button, Card, ErrorState, ImagePicker, Input, SkeletonCard, Textarea, useToast } from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { applyServerErrors, optionalNumberField } from '@/lib/form';
import { formatDateTime, percentToBP } from '@/lib/format';
import { ReceiptPreview } from '../components/ReceiptPreview';
import { useSettings, useUpdateSettings, useUploadLogo } from '../hooks/useSettings';

const schema = z.object({
  name: z.string().trim().min(2, 'Minimal 2 karakter').max(100),
  address: z.string().max(500).optional(),
  phone: z.string().max(30).optional(),
  tax_percent: optionalNumberField().pipe(z.number().min(0, 'Minimal 0%').max(100, 'Maksimal 100%')),
  receipt_footer: z.string().max(500).optional(),
});
type FormValues = z.infer<typeof schema>;

export default function SettingsPage() {
  const toast = useToast();
  const query = useSettings();
  const update = useUpdateSettings();
  const uploadLogo = useUploadLogo();
  const [logoError, setLogoError] = useState<string>();
  const { register, handleSubmit, reset, watch, setError, formState: { errors, isDirty } } = useForm<FormValues>({ resolver: zodResolver(schema) });

  useEffect(() => {
    if (!query.data) return;
    reset({
      name: query.data.name,
      address: query.data.address ?? '',
      phone: query.data.phone ?? '',
      tax_percent: query.data.tax_rate_bp / 100,
      receipt_footer: query.data.receipt_footer ?? '',
    });
  }, [query.data, reset]);

  const values = watch();

  const onSubmit = handleSubmit((v) =>
    update.mutate(
      {
        name: v.name,
        address: v.address || null,
        phone: v.phone || null,
        tax_rate_bp: percentToBP(v.tax_percent),
        receipt_footer: v.receipt_footer || null,
      },
      {
        onSuccess: () => toast.success('Pengaturan toko disimpan'),
        onError: (err) => {
          const msg = applyServerErrors(err, setError, ['name', 'address', 'phone', 'receipt_footer']);
          if (msg) toast.error(msg);
        },
      },
    ),
  );

  return (
    <>
      <PageHeader title="Pengaturan Toko" description="Informasi ini tampil di struk dan aplikasi kasir." />
      {query.isLoading ? (
        <div className="grid gap-4 lg:grid-cols-3">
          <SkeletonCard className="lg:col-span-2" lines={8} />
          <SkeletonCard lines={6} />
        </div>
      ) : query.isError || !query.data ? (
        <Card><ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} /></Card>
      ) : (
        <div className="grid gap-4 lg:grid-cols-3">
          <Card title="Informasi toko" subtitle={`Terakhir diubah ${formatDateTime(query.data.updated_at)}`} className="lg:col-span-2">
            <form onSubmit={onSubmit} className="space-y-4" noValidate>
              <ImagePicker
                label="Logo toko"
                currentUrl={query.data.logo_url}
                file={null}
                error={logoError}
                onChange={(file, err) => {
                  setLogoError(err);
                  if (!file) return;
                  uploadLogo.mutate(file, {
                    onSuccess: () => toast.success('Logo berhasil diunggah'),
                    onError: (e) => toast.error(errorMessage(e)),
                  });
                }}
              />
              <Input label="Nama toko" required error={errors.name?.message} {...register('name')} />
              <Textarea label="Alamat" rows={2} error={errors.address?.message} {...register('address')} />
              <div className="grid gap-4 sm:grid-cols-2">
                <Input label="Nomor telepon" type="tel" error={errors.phone?.message} {...register('phone')} />
                <Input
                  label="Pajak (%)"
                  type="number"
                  inputMode="decimal"
                  step="0.01"
                  min={0}
                  max={100}
                  hint="Isi 0 jika toko tidak memungut pajak. Contoh: 11 untuk PPN 11%"
                  error={errors.tax_percent?.message}
                  {...register('tax_percent')}
                />
              </div>
              <Textarea label="Catatan kaki struk" rows={3} placeholder="Terima kasih atas kunjungan Anda" error={errors.receipt_footer?.message} {...register('receipt_footer')} />
              <div className="flex justify-end gap-2">
                <Button variant="secondary" disabled={!isDirty || update.isPending} onClick={() => reset()}>Batalkan perubahan</Button>
                <Button type="submit" loading={update.isPending} disabled={!isDirty}>Simpan</Button>
              </div>
            </form>
          </Card>
          <Card title="Pratinjau struk">
            <ReceiptPreview
              name={values.name ?? ''}
              address={values.address}
              phone={values.phone}
              footer={values.receipt_footer}
              taxBP={percentToBP(Number(values.tax_percent) || 0)}
              logoUrl={query.data.logo_url}
            />
          </Card>
        </div>
      )}
    </>
  );
}
