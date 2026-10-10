import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button, Input, Modal, Textarea, useToast } from '@/components/ui';
import { cn } from '@/lib/cn';
import { applyServerErrors, numberField } from '@/lib/form';
import { ProductPicker } from '@/features/products/components/ProductPicker';
import { ProductThumb } from '@/features/products/components/ProductThumb';
import type { Product } from '@/features/products/types';
import { useStockAdjustment } from '../hooks/useStock';

const schema = z.object({
  actual_stock: numberField('Stok aktual wajib diisi').pipe(z.number().int('Harus bilangan bulat').min(0, 'Tidak boleh negatif')),
  note: z.string().trim().min(3, 'Tulis alasan penyesuaian (minimal 3 karakter)').max(500),
});
type FormValues = z.infer<typeof schema>;

export function StockAdjustmentModal({ open, onClose, initialProduct }: { open: boolean; onClose: () => void; initialProduct?: Product | null }) {
  const toast = useToast();
  const adjust = useStockAdjustment();
  const [product, setProduct] = useState<Product | null>(null);
  const { register, handleSubmit, reset, watch, setError, formState: { errors } } = useForm<FormValues>({ resolver: zodResolver(schema) });

  useEffect(() => {
    if (!open) return;
    setProduct(initialProduct ?? null);
    reset({ actual_stock: initialProduct?.stock ?? ('' as unknown as number), note: '' });
  }, [open, initialProduct, reset]);

  const actual = watch('actual_stock');
  const diff = product && actual !== undefined && String(actual) !== '' ? Number(actual) - product.stock : null;

  const onSubmit = handleSubmit((v) => {
    if (!product) return;
    adjust.mutate(
      { product_id: product.id, actual_stock: v.actual_stock, note: v.note },
      {
        onSuccess: () => {
          toast.success(`Stok ${product.name} disesuaikan menjadi ${v.actual_stock}`);
          onClose();
        },
        onError: (err) => {
          const msg = applyServerErrors(err, setError, ['actual_stock', 'note']);
          if (msg) toast.error(msg);
        },
      },
    );
  });

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Penyesuaian Stok"
      description="Samakan stok sistem dengan hasil hitung fisik (stock opname)."
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Batal</Button>
          <Button type="submit" form="adjust-form" disabled={!product || diff === 0} loading={adjust.isPending}>Simpan</Button>
        </>
      }
    >
      <form id="adjust-form" onSubmit={onSubmit} className="space-y-4" noValidate>
        {product ? (
          <div className="flex items-center gap-3 rounded-lg bg-fill/60 p-3">
            <ProductThumb url={product.image_url} name={product.name} />
            <div className="min-w-0 flex-1">
              <p className="truncate text-subheadline font-semibold">{product.name}</p>
              <p className="text-caption text-text-secondary">Stok di sistem: {product.stock} {product.unit.name}</p>
            </div>
            <Button variant="ghost" size="sm" onClick={() => setProduct(null)}>Ganti</Button>
          </div>
        ) : (
          <ProductPicker
            onSelect={(p) => {
              setProduct(p);
              reset({ actual_stock: p.stock, note: '' });
            }}
          />
        )}
        <Input
          label="Stok aktual (hasil hitung)"
          type="number"
          inputMode="numeric"
          min={0}
          required
          disabled={!product}
          error={errors.actual_stock?.message}
          {...register('actual_stock')}
        />
        {diff !== null && (
          <p className={cn('rounded-md px-3 py-2 text-footnote font-medium', diff === 0 ? 'bg-fill text-text-secondary' : diff > 0 ? 'bg-success/10 text-success' : 'bg-danger/10 text-danger')}>
            {diff === 0 ? 'Tidak ada selisih' : `Selisih ${diff > 0 ? '+' : ''}${diff} ${product?.unit.name ?? ''}`}
          </p>
        )}
        <Textarea label="Alasan" required placeholder="Contoh: barang rusak, kedaluwarsa, salah hitung" error={errors.note?.message} {...register('note')} />
      </form>
    </Modal>
  );
}
