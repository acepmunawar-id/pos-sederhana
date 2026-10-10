import { ChevronLeft, ChevronRight } from 'lucide-react';
import type { PageMeta } from '@/types/api';
import { formatNumber } from '@/lib/format';
import { Button } from './Button';

export function Pagination({ meta, onChange }: { meta?: PageMeta; onChange: (page: number) => void }) {
  if (!meta || meta.total === 0) return null;
  const pages = Math.max(1, Math.ceil(meta.total / meta.limit));
  const from = (meta.page - 1) * meta.limit + 1;
  const to = Math.min(meta.page * meta.limit, meta.total);
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 border-t border-separator/60 px-4 py-3">
      <p className="text-footnote text-text-secondary">
        Menampilkan {formatNumber(from)}–{formatNumber(to)} dari {formatNumber(meta.total)} data
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="secondary"
          size="sm"
          disabled={meta.page <= 1}
          onClick={() => onChange(meta.page - 1)}
          icon={<ChevronLeft className="h-4 w-4" />}
          aria-label="Halaman sebelumnya"
        >
          Sebelumnya
        </Button>
        <span className="text-footnote tabular-nums text-text-secondary">
          {meta.page} / {pages}
        </span>
        <Button
          variant="secondary"
          size="sm"
          disabled={meta.page >= pages}
          onClick={() => onChange(meta.page + 1)}
          aria-label="Halaman berikutnya"
        >
          Berikutnya
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
