import { useEffect, type ReactNode } from 'react';

export function PageHeader({ title, description, actions }: { title: string; description?: string; actions?: ReactNode }) {
  useEffect(() => {
    document.title = `${title} · POS UMKM`;
  }, [title]);

  return (
    <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="text-large-title">{title}</h1>
        {description && <p className="mt-1 text-subheadline text-text-secondary">{description}</p>}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
    </div>
  );
}
