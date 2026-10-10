import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/cn';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  title?: string;
  subtitle?: string;
  action?: ReactNode;
  padded?: boolean;
}

export function Card({ title, subtitle, action, padded = true, className, children, ...rest }: CardProps) {
  return (
    <section className={cn('rounded-xl bg-card shadow-card', className)} {...rest}>
      {(title || action) && (
        <header className="flex flex-wrap items-center justify-between gap-3 px-5 pb-1 pt-4">
          <div>
            {title && <h3 className="text-headline">{title}</h3>}
            {subtitle && <p className="text-footnote text-text-secondary">{subtitle}</p>}
          </div>
          {action}
        </header>
      )}
      <div className={cn(padded && 'p-5', padded && (title || action) && 'pt-3')}>{children}</div>
    </section>
  );
}

/** Large-number summary tile used on dashboard & reports. */
export function StatCard({
  label,
  value,
  hint,
  icon,
  tone = 'primary',
}: {
  label: string;
  value: ReactNode;
  hint?: ReactNode;
  icon?: ReactNode;
  tone?: 'primary' | 'success' | 'warning' | 'danger';
}) {
  const tones = {
    primary: 'bg-primary/10 text-primary',
    success: 'bg-success/10 text-success',
    warning: 'bg-warning/10 text-warning',
    danger: 'bg-danger/10 text-danger',
  };
  return (
    <div className="flex items-start gap-4 rounded-xl bg-card p-4 shadow-card sm:p-5">
      {icon && <span className={cn('hidden h-11 w-11 shrink-0 items-center justify-center rounded-lg sm:flex', tones[tone])}>{icon}</span>}
      <div className="min-w-0">
        <p className="text-footnote text-text-secondary">{label}</p>
        <p className="mt-0.5 truncate text-[20px] font-bold leading-7 tabular-nums sm:text-title">{value}</p>
        {hint && <p className="mt-0.5 text-footnote text-text-secondary">{hint}</p>}
      </div>
    </div>
  );
}
