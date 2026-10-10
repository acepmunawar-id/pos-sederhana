import { cn } from '@/lib/cn';

export function Skeleton({ className }: { className?: string }) {
  return <div className={cn('animate-pulse rounded-sm bg-fill', className)} aria-hidden />;
}

/** Placeholder for a card with a few lines. */
export function SkeletonCard({ lines = 3, className }: { lines?: number; className?: string }) {
  return (
    <div className={cn('rounded-xl bg-card p-5 shadow-card', className)}>
      <Skeleton className="mb-4 h-4 w-1/3" />
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton key={i} className="mb-2 h-4 w-full last:mb-0" />
      ))}
    </div>
  );
}
