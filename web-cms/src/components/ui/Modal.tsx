import { useEffect, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { X } from 'lucide-react';
import { cn } from '@/lib/cn';

export interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  size?: 'sm' | 'md' | 'lg';
}

const widths = { sm: 'max-w-sm', md: 'max-w-lg', lg: 'max-w-2xl' };

export function Modal({ open, onClose, title, description, children, footer, size = 'md' }: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    document.addEventListener('keydown', onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = prev;
    };
  }, [open, onClose]);

  if (!open) return null;
  return createPortal(
    <div className="fixed inset-0 z-50 flex items-end justify-center sm:items-center sm:p-4" role="dialog" aria-modal="true">
      <div className="absolute inset-0 animate-fade-in bg-black/40" onClick={onClose} aria-hidden />
      <div
        className={cn(
          'relative flex max-h-[92vh] w-full animate-slide-up flex-col rounded-t-xl bg-card shadow-modal sm:rounded-xl',
          widths[size],
        )}
      >
        <header className="flex items-start justify-between gap-4 border-b border-separator/60 px-5 py-4">
          <div>
            <h2 className="text-headline">{title}</h2>
            {description && <p className="mt-0.5 text-footnote text-text-secondary">{description}</p>}
          </div>
          <button
            type="button"
            onClick={onClose}
            className="-mr-2 flex h-touch w-touch items-center justify-center rounded-full text-gray hover:bg-fill"
            aria-label="Tutup"
          >
            <X className="h-5 w-5" />
          </button>
        </header>
        <div className="flex-1 overflow-y-auto px-5 py-4">{children}</div>
        {footer && <footer className="flex justify-end gap-2 border-t border-separator/60 px-5 py-3">{footer}</footer>}
      </div>
    </div>,
    document.body,
  );
}
