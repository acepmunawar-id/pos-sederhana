import { useEffect, useRef, useState } from 'react';
import { Loader2 } from 'lucide-react';
import { SearchInput } from '@/components/ui';
import { useDebounced } from '@/lib/use-list-state';
import { useProducts } from '../hooks/useProducts';
import type { Product } from '../types';
import { ProductThumb } from './ProductThumb';

/** Search-as-you-type product selector. */
export function ProductPicker({
  onSelect,
  excludeIds = [],
  placeholder = 'Cari produk (nama / SKU / barcode)…',
}: {
  onSelect: (p: Product) => void;
  excludeIds?: number[];
  placeholder?: string;
}) {
  const [term, setTerm] = useState('');
  const [open, setOpen] = useState(false);
  const search = useDebounced(term, 250);
  const ref = useRef<HTMLDivElement>(null);
  const query = useProducts({ search, limit: 8, sort: 'name', order: 'asc' });
  const items = (query.data?.items ?? []).filter((p) => !excludeIds.includes(p.id));

  useEffect(() => {
    const onClick = (e: MouseEvent) => !ref.current?.contains(e.target as Node) && setOpen(false);
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, []);

  return (
    <div className="relative" ref={ref} onFocus={() => setOpen(true)}>
      <SearchInput className="sm:w-full" value={term} onChange={(v) => { setTerm(v); setOpen(true); }} placeholder={placeholder} />
      {open && (
        <div className="absolute z-10 mt-1 max-h-72 w-full overflow-y-auto rounded-lg bg-card shadow-modal">
          {query.isFetching && items.length === 0 ? (
            <div className="flex items-center justify-center py-6 text-gray"><Loader2 className="h-5 w-5 animate-spin" /></div>
          ) : items.length === 0 ? (
            <p className="px-4 py-5 text-center text-footnote text-text-secondary">Produk tidak ditemukan</p>
          ) : (
            items.map((p) => (
              <button
                key={p.id}
                type="button"
                onClick={() => {
                  onSelect(p);
                  setTerm('');
                  setOpen(false);
                }}
                className="flex min-h-touch w-full items-center gap-3 px-3 py-2 text-left hover:bg-fill"
              >
                <ProductThumb url={p.image_url} name={p.name} />
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-subheadline font-medium">{p.name}</span>
                  <span className="block text-caption text-text-secondary">{p.sku}</span>
                </span>
                <span className="text-footnote tabular-nums text-text-secondary">Stok {p.stock}</span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
