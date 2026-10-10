import type { ReactNode } from 'react';
import { ArrowDown, ArrowUp, ArrowUpDown } from 'lucide-react';
import { cn } from '@/lib/cn';
import type { SortState } from '@/lib/use-list-state';
import { Skeleton } from './Skeleton';

export interface Column<T> {
  key: string;
  header: string;
  render: (row: T) => ReactNode;
  /** Server-side sort field; enables the sort button. */
  sortKey?: string;
  align?: 'left' | 'right' | 'center';
  className?: string;
  /** Hide on narrow (tablet portrait) screens. */
  hideOnMobile?: boolean;
  /** Role of the column in the phone card layout: the card title, or the actions footer. */
  mobile?: 'title' | 'actions' | 'hidden';
}

export interface TableProps<T> {
  columns: Column<T>[];
  rows: T[] | undefined;
  rowKey: (row: T) => string | number;
  loading?: boolean;
  sort?: SortState;
  onSort?: (field: string) => void;
  onRowClick?: (row: T) => void;
  empty?: ReactNode;
  skeletonRows?: number;
}

const alignClass = { left: 'text-left', right: 'text-right', center: 'text-center' };

export function Table<T>({
  columns,
  rows,
  rowKey,
  loading,
  sort,
  onSort,
  onRowClick,
  empty,
  skeletonRows = 6,
}: TableProps<T>) {
  const showEmpty = !loading && rows && rows.length === 0;
  return (
    <>
      <MobileList columns={columns} rows={rows} rowKey={rowKey} loading={loading} onRowClick={onRowClick} skeletonRows={skeletonRows} />
      <div className="hidden overflow-x-auto md:block">
      <table className="w-full min-w-[640px] border-collapse">
        <thead>
          <tr className="border-b border-separator/60">
            {columns.map((c) => {
              const active = sort?.sort === c.sortKey;
              const SortIcon = !active ? ArrowUpDown : sort?.order === 'asc' ? ArrowUp : ArrowDown;
              return (
                <th
                  key={c.key}
                  scope="col"
                  className={cn(
                    'whitespace-nowrap px-4 py-3 text-footnote font-semibold uppercase tracking-wide text-text-secondary',
                    alignClass[c.align ?? 'left'],
                    c.hideOnMobile && 'hidden lg:table-cell',
                    c.className,
                  )}
                >
                  {c.sortKey && onSort ? (
                    <button
                      type="button"
                      onClick={() => onSort(c.sortKey!)}
                      className={cn('inline-flex items-center gap-1 uppercase hover:text-text', active && 'text-primary')}
                    >
                      {c.header}
                      <SortIcon className="h-3.5 w-3.5" />
                    </button>
                  ) : (
                    c.header
                  )}
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody>
          {loading &&
            Array.from({ length: skeletonRows }).map((_, i) => (
              <tr key={i} className="border-b border-separator/40 last:border-0">
                {columns.map((c) => (
                  <td key={c.key} className={cn('px-4 py-3.5', c.hideOnMobile && 'hidden lg:table-cell')}>
                    <Skeleton className="h-4 w-full max-w-[160px]" />
                  </td>
                ))}
              </tr>
            ))}
          {!loading &&
            rows?.map((row) => (
              <tr
                key={rowKey(row)}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
                className={cn(
                  'border-b border-separator/40 last:border-0',
                  onRowClick && 'cursor-pointer hover:bg-fill/60',
                )}
              >
                {columns.map((c) => (
                  <td
                    key={c.key}
                    className={cn(
                      'px-4 py-3 text-subheadline',
                      alignClass[c.align ?? 'left'],
                      c.hideOnMobile && 'hidden lg:table-cell',
                      c.className,
                    )}
                  >
                    {c.render(row)}
                  </td>
                ))}
              </tr>
            ))}
        </tbody>
      </table>
      </div>
      {showEmpty && empty}
    </>
  );
}

/**
 * Phone layout (< 768px): each row becomes a card. The first column (or the one marked
 * `mobile: 'title'`) is the heading, other columns become label/value lines, and the
 * `actions` column sits at the bottom.
 */
function MobileList<T>({
  columns,
  rows,
  rowKey,
  loading,
  onRowClick,
  skeletonRows = 6,
}: Pick<TableProps<T>, 'columns' | 'rows' | 'rowKey' | 'loading' | 'onRowClick' | 'skeletonRows'>) {
  const title = columns.find((c) => c.mobile === 'title') ?? columns[0];
  const actions = columns.find((c) => c.mobile === 'actions' || (c.key === 'actions' && !c.header));
  const fields = columns.filter((c) => c !== title && c !== actions && c.mobile !== 'hidden' && c.header);

  if (loading) {
    return (
      <ul className="divide-y divide-separator/50 md:hidden">
        {Array.from({ length: skeletonRows }).map((_, i) => (
          <li key={i} className="space-y-2 px-4 py-4">
            <Skeleton className="h-4 w-2/3" />
            <Skeleton className="h-3 w-1/2" />
          </li>
        ))}
      </ul>
    );
  }
  return (
    <ul className="divide-y divide-separator/50 md:hidden">
      {rows?.map((row) => (
        <li
          key={rowKey(row)}
          onClick={onRowClick ? () => onRowClick(row) : undefined}
          className={cn('px-4 py-3.5', onRowClick && 'cursor-pointer active:bg-fill/60')}
        >
          <div className="mb-2 text-subheadline">{title.render(row)}</div>
          <dl className="grid grid-cols-2 gap-x-4 gap-y-1.5">
            {fields.map((c) => (
              <div key={c.key} className="min-w-0">
                <dt className="text-caption uppercase tracking-wide text-text-secondary">{c.header}</dt>
                <dd className="truncate text-subheadline">{c.render(row)}</dd>
              </div>
            ))}
          </dl>
          {actions && <div className="mt-2 border-t border-separator/40 pt-1">{actions.render(row)}</div>}
        </li>
      ))}
    </ul>
  );
}
