import type { ReactNode } from 'react';
import { AlertCircle, Inbox, RotateCw } from 'lucide-react';
import { cn } from '@/lib/cn';
import { Button } from './Button';

/** Shown when a list or page has no data. */
export function EmptyState({
  title,
  description,
  icon,
  action,
  className,
}: {
  title: string;
  description?: string;
  icon?: ReactNode;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn('flex flex-col items-center justify-center px-6 py-14 text-center', className)}>
      <span className="mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-fill text-gray">
        {icon ?? <Inbox className="h-7 w-7" />}
      </span>
      <p className="text-headline">{title}</p>
      {description && <p className="mt-1 max-w-sm text-subheadline text-text-secondary">{description}</p>}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}

/** Shown when a query fails. */
export function ErrorState({
  message,
  onRetry,
  className,
}: {
  message: string;
  onRetry?: () => void;
  className?: string;
}) {
  return (
    <div className={cn('flex flex-col items-center justify-center px-6 py-14 text-center', className)} role="alert">
      <span className="mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-danger/10 text-danger">
        <AlertCircle className="h-7 w-7" />
      </span>
      <p className="text-headline">Gagal memuat data</p>
      <p className="mt-1 max-w-sm text-subheadline text-text-secondary">{message}</p>
      {onRetry && (
        <Button variant="secondary" className="mt-4" icon={<RotateCw className="h-4 w-4" />} onClick={onRetry}>
          Coba lagi
        </Button>
      )}
    </div>
  );
}
