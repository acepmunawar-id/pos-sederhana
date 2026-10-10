import { Package } from 'lucide-react';

export function ProductThumb({ url, name }: { url: string | null; name: string }) {
  return url ? (
    <img src={url} alt={name} className="h-10 w-10 shrink-0 rounded-md bg-fill object-cover" loading="lazy" />
  ) : (
    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-fill text-gray" aria-hidden>
      <Package className="h-5 w-5" />
    </span>
  );
}
