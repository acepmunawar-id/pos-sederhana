import { useEffect, useState } from 'react';
import { Minus, Plus, Trash2 } from 'lucide-react';
import { Button, EmptyState, Modal, Textarea, useToast } from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { ProductPicker } from '@/features/products/components/ProductPicker';
import { ProductThumb } from '@/features/products/components/ProductThumb';
import type { Product } from '@/features/products/types';
import { useStockIn } from '../hooks/useStock';

interface Line {
  product: Product;
  quantity: number;
}

export function StockInModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const toast = useToast();
  const stockIn = useStockIn();
  const [lines, setLines] = useState<Line[]>([]);
  const [note, setNote] = useState('');

  useEffect(() => {
    if (open) {
      setLines([]);
      setNote('');
    }
  }, [open]);

  const setQty = (id: number, qty: number) =>
    setLines((ls) => ls.map((l) => (l.product.id === id ? { ...l, quantity: Math.max(1, Math.min(1_000_000, qty || 1)) } : l)));

  const submit = () =>
    stockIn.mutate(
      { items: lines.map((l) => ({ product_id: l.product.id, quantity: l.quantity })), note: note.trim() || null },
      {
        onSuccess: () => {
          toast.success(`Stok masuk dicatat untuk ${lines.length} produk`);
          onClose();
        },
        onError: (err) => toast.error(errorMessage(err)),
      },
    );

  const totalQty = lines.reduce((s, l) => s + l.quantity, 0);

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Stok Masuk"
      description="Catat barang yang datang dari supplier."
      size="lg"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Batal</Button>
          <Button onClick={submit} disabled={lines.length === 0} loading={stockIn.isPending}>
            Simpan ({totalQty} item)
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <ProductPicker
          excludeIds={lines.map((l) => l.product.id)}
          onSelect={(p) => setLines((ls) => [...ls, { product: p, quantity: 1 }])}
        />
        {lines.length === 0 ? (
          <EmptyState className="py-8" title="Belum ada produk" description="Cari dan pilih produk yang stoknya bertambah." />
        ) : (
          <ul className="divide-y divide-separator/60 overflow-hidden rounded-lg bg-fill/40">
            {lines.map((l) => (
              <li key={l.product.id} className="flex flex-wrap items-center gap-3 px-3 py-2">
                <ProductThumb url={l.product.image_url} name={l.product.name} />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-subheadline font-medium">{l.product.name}</p>
                  <p className="text-caption text-text-secondary">
                    Stok {l.product.stock} → <b className="text-success">{l.product.stock + l.quantity}</b>
                  </p>
                </div>
                <div className="flex items-center gap-1">
                  <button type="button" aria-label="Kurangi" onClick={() => setQty(l.product.id, l.quantity - 1)} className="flex h-touch w-touch items-center justify-center rounded-md bg-card text-primary">
                    <Minus className="h-4 w-4" />
                  </button>
                  <input
                    type="number"
                    aria-label={`Jumlah ${l.product.name}`}
                    value={l.quantity}
                    onChange={(e) => setQty(l.product.id, Number(e.target.value))}
                    className="h-touch w-20 rounded-md bg-card text-center text-headline tabular-nums focus:outline-none focus:ring-2 focus:ring-primary/40"
                  />
                  <button type="button" aria-label="Tambah" onClick={() => setQty(l.product.id, l.quantity + 1)} className="flex h-touch w-touch items-center justify-center rounded-md bg-card text-primary">
                    <Plus className="h-4 w-4" />
                  </button>
                </div>
                <button
                  type="button"
                  onClick={() => setLines((ls) => ls.filter((x) => x.product.id !== l.product.id))}
                  className="inline-flex h-touch items-center gap-1 rounded-md px-2 text-footnote font-semibold text-danger hover:bg-card"
                >
                  <Trash2 className="h-4 w-4" /> Hapus
                </button>
              </li>
            ))}
          </ul>
        )}
        <Textarea label="Catatan" placeholder="Contoh: Pembelian dari Supplier A, faktur #123" value={note} onChange={(e) => setNote(e.target.value)} maxLength={500} />
      </div>
    </Modal>
  );
}
