import { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button, ImagePicker, Input, Modal, Select, Switch, useToast } from '@/components/ui';
import { applyServerErrors, numberField, optionalNumberField } from '@/lib/form';
import { formatRupiah } from '@/lib/format';
import { useCategoryOptions, useUnitOptions } from '@/features/categories/hooks/useCategories';
import { useCreateProduct, useUpdateProduct, useUploadProductImage } from '../hooks/useProducts';
import type { Product } from '../types';

const money = (label: string) => numberField(`${label} wajib diisi`).pipe(z.number().int('Harus bilangan bulat').min(0, 'Tidak boleh negatif'));
const count = () => optionalNumberField().pipe(z.number().int('Harus bilangan bulat').min(0, 'Tidak boleh negatif'));

const schema = z.object({
    name: z.string().trim().min(2, 'Minimal 2 karakter').max(150),
    sku: z.string().trim().min(1, 'SKU wajib diisi').max(50),
    barcode: z.string().trim().max(50).optional(),
    category_id: numberField('Pilih kategori'),
    unit_id: numberField('Pilih satuan'),
    price: money('Harga jual'),
    cost: money('Harga modal'),
    min_stock: count(),
    initial_stock: count(),
    is_active: z.boolean(),
});
type FormValues = z.infer<typeof schema>;

const fields = ['name', 'sku', 'barcode', 'category_id', 'unit_id', 'price', 'cost', 'min_stock', 'initial_stock'];

export function ProductFormModal({ open, onClose, product }: { open: boolean; onClose: () => void; product?: Product | null }) {
  const toast = useToast();
  const categories = useCategoryOptions();
  const units = useUnitOptions();
  const create = useCreateProduct();
  const update = useUpdateProduct();
  const uploadImage = useUploadProductImage();
  const [image, setImage] = useState<File | null>(null);
  const [imageError, setImageError] = useState<string>();

  const { register, handleSubmit, reset, setError, control, watch, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
  });

  useEffect(() => {
    if (!open) return;
    setImage(null);
    setImageError(undefined);
    reset({
      name: product?.name ?? '',
      sku: product?.sku ?? '',
      barcode: product?.barcode ?? '',
      category_id: product?.category.id ?? ('' as unknown as number),
      unit_id: product?.unit.id ?? ('' as unknown as number),
      price: product?.price ?? ('' as unknown as number),
      cost: product?.cost ?? ('' as unknown as number),
      min_stock: product?.min_stock ?? 0,
      initial_stock: 0,
      is_active: product?.is_active ?? true,
    });
  }, [open, product, reset]);

  const price = Number(watch('price')) || 0;
  const cost = Number(watch('cost')) || 0;
  const margin = price > 0 ? Math.round(((price - cost) / price) * 1000) / 10 : 0;

  const onSubmit = handleSubmit(async (v) => {
    const input = {
      name: v.name,
      sku: v.sku,
      barcode: v.barcode || null,
      category_id: v.category_id,
      unit_id: v.unit_id,
      price: v.price,
      cost: v.cost,
      min_stock: v.min_stock,
      is_active: v.is_active,
    };
    try {
      const saved = product
        ? await update.mutateAsync({ id: product.id, input })
        : await create.mutateAsync({ ...input, initial_stock: v.initial_stock });
      if (image) {
        try {
          await uploadImage.mutateAsync({ id: saved.id, file: image });
        } catch (err) {
          toast.error(`Produk tersimpan, tetapi gambar gagal diunggah: ${(err as Error).message}`);
          onClose();
          return;
        }
      }
      toast.success(product ? 'Produk diperbarui' : 'Produk ditambahkan');
      onClose();
    } catch (err) {
      const msg = applyServerErrors(err, setError, fields);
      if (msg) toast.error(msg);
    }
  });

  const saving = create.isPending || update.isPending || uploadImage.isPending;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={product ? 'Ubah Produk' : 'Tambah Produk'}
      description={product ? 'Stok diubah lewat menu Stok (stok masuk / penyesuaian).' : undefined}
      size="lg"
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={saving}>Batal</Button>
          <Button type="submit" form="product-form" loading={saving}>Simpan</Button>
        </>
      }
    >
      <form id="product-form" onSubmit={onSubmit} className="grid gap-4 sm:grid-cols-[auto_1fr]" noValidate>
        <ImagePicker
          label="Gambar"
          currentUrl={product?.image_url}
          file={image}
          error={imageError}
          onChange={(f, err) => {
            setImage(f);
            setImageError(err);
          }}
        />
        <div className="grid gap-4 sm:grid-cols-2">
          <Input className="sm:col-span-2" label="Nama produk" required autoFocus error={errors.name?.message} {...register('name')} />
          <Input label="SKU" required placeholder="MNM-001" error={errors.sku?.message} {...register('sku')} />
          <Input label="Barcode" placeholder="Opsional" error={errors.barcode?.message} {...register('barcode')} />
          <Select
            label="Kategori"
            required
            placeholder={categories.isLoading ? 'Memuat…' : 'Pilih kategori'}
            options={(categories.data?.items ?? []).map((c) => ({ value: c.id, label: c.name }))}
            error={errors.category_id?.message}
            {...register('category_id')}
          />
          <Select
            label="Satuan"
            required
            placeholder={units.isLoading ? 'Memuat…' : 'Pilih satuan'}
            options={(units.data?.items ?? []).map((u) => ({ value: u.id, label: `${u.name} (${u.abbreviation})` }))}
            error={errors.unit_id?.message}
            {...register('unit_id')}
          />
          <Input label="Harga jual" prefix="Rp" type="number" inputMode="numeric" min={0} required error={errors.price?.message} {...register('price')} />
          <Input
            label="Harga modal"
            prefix="Rp"
            type="number"
            inputMode="numeric"
            min={0}
            required
            hint={price > 0 ? `Margin ${margin}% · laba ${formatRupiah(price - cost)} per item` : undefined}
            error={errors.cost?.message}
            {...register('cost')}
          />
          {!product && (
            <Input label="Stok awal" type="number" inputMode="numeric" min={0} error={errors.initial_stock?.message} {...register('initial_stock')} />
          )}
          <Input
            label="Stok minimum"
            type="number"
            inputMode="numeric"
            min={0}
            hint="Peringatan stok menipis muncul di angka ini"
            error={errors.min_stock?.message}
            {...register('min_stock')}
          />
          <div className="sm:col-span-2">
            <Controller
              control={control}
              name="is_active"
              render={({ field }) => (
                <Switch
                  checked={field.value}
                  onChange={field.onChange}
                  label="Produk aktif"
                  description="Produk nonaktif tidak tampil di aplikasi kasir"
                />
              )}
            />
          </div>
        </div>
      </form>
    </Modal>
  );
}
