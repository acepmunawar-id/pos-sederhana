import { useEffect, useRef, useState } from 'react';
import { ImagePlus } from 'lucide-react';
import { cn } from '@/lib/cn';

const MAX_BYTES = 2 * 1024 * 1024;
const ACCEPT = ['image/jpeg', 'image/png', 'image/webp'];

/** Square image picker with preview and client-side type/size validation. */
export function ImagePicker({
  label,
  currentUrl,
  file,
  onChange,
  error,
  className,
}: {
  label: string;
  currentUrl?: string | null;
  file: File | null;
  onChange: (file: File | null, error?: string) => void;
  error?: string;
  className?: string;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [preview, setPreview] = useState<string | null>(null);

  useEffect(() => {
    if (!file) {
      setPreview(null);
      return;
    }
    const url = URL.createObjectURL(file);
    setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  const src = preview ?? currentUrl ?? null;

  return (
    <div className={cn('flex flex-col gap-1.5', className)}>
      <span className="text-footnote font-medium text-text-secondary">{label}</span>
      <button
        type="button"
        onClick={() => inputRef.current?.click()}
        className={cn(
          'group relative flex aspect-square w-32 items-center justify-center overflow-hidden rounded-lg bg-fill text-gray',
          'hover:ring-2 hover:ring-primary/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40',
          error && 'ring-2 ring-danger/50',
        )}
      >
        {src ? (
          <img src={src} alt="" className="h-full w-full object-cover" />
        ) : (
          <span className="flex flex-col items-center gap-1 text-caption">
            <ImagePlus className="h-6 w-6" /> Pilih gambar
          </span>
        )}
        {src && (
          <span className="absolute inset-x-0 bottom-0 bg-black/50 py-1 text-center text-caption font-semibold text-white opacity-0 transition group-hover:opacity-100">
            Ganti
          </span>
        )}
      </button>
      <input
        ref={inputRef}
        type="file"
        accept={ACCEPT.join(',')}
        className="hidden"
        onChange={(e) => {
          const f = e.target.files?.[0] ?? null;
          e.target.value = '';
          if (!f) return;
          if (!ACCEPT.includes(f.type)) return onChange(null, 'Format harus JPG, PNG, atau WEBP');
          if (f.size > MAX_BYTES) return onChange(null, 'Ukuran maksimal 2 MB');
          onChange(f);
        }}
      />
      {error ? (
        <p className="text-caption text-danger">{error}</p>
      ) : (
        <p className="text-caption text-text-secondary">JPG/PNG/WEBP, maks 2 MB</p>
      )}
    </div>
  );
}
