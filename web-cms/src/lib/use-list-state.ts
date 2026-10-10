import { useEffect, useState } from 'react';
import type { ListParams } from '@/types/api';

export function useDebounced<T>(value: T, delay = 350): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

export interface SortState {
  sort?: string;
  order: 'asc' | 'desc';
}

/**
 * Page, search, sort and filter state for list pages.
 * Changing search, sort or a filter resets to page 1.
 */
export function useListState<F extends Record<string, string | number | boolean | undefined>>(
  initialFilters: F,
  initialSort: SortState = { order: 'desc' },
  limit = 10,
) {
  const [page, setPage] = useState(1);
  const [search, setSearchRaw] = useState('');
  const [sort, setSortRaw] = useState<SortState>(initialSort);
  const [filters, setFiltersRaw] = useState<F>(initialFilters);
  const debouncedSearch = useDebounced(search);

  useEffect(() => setPage(1), [debouncedSearch]);

  const params: ListParams = { page, limit, search: debouncedSearch, sort: sort.sort, order: sort.order, ...filters };

  return {
    params,
    page,
    setPage,
    search,
    setSearch: setSearchRaw,
    sort,
    toggleSort: (field: string) => {
      setSortRaw((s) => ({ sort: field, order: s.sort === field && s.order === 'asc' ? 'desc' : 'asc' }));
      setPage(1);
    },
    filters,
    setFilter: <K extends keyof F>(key: K, value: F[K]) => {
      setFiltersRaw((f) => ({ ...f, [key]: value }));
      setPage(1);
    },
  };
}
